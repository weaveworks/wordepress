package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/weaveworks/blackfriday"
	"github.com/weaveworks/wordepress"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	stdpath "path"
	"regexp"
	"strings"
)

var AnchorRegexp = regexp.MustCompile(`<a href="([^"]*)"`)
var ImgRegexp = regexp.MustCompile(`<img src="([^"]*)"`)

type PostRequest struct {
	Parent    int    `json:"parent"`
	MenuOrder int    `json:"menu_order"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Status    string `json:"status"`
	Slug      string `json:"slug"`
	Product   string `json:"wpcf-product"`
	Version   string `json:"wpcf-version"`
}

type Content struct {
	Rendered string `json:"rendered"`
}

type PostResponse struct {
	Id      int     `json:"id"`
	Slug    string  `json:"slug"`
	Content Content `json:"content"`
}

type Media struct {
	MediaDetails MediaDetails `json:"media_details"`
}

type MediaDetails struct {
	File string `json:"file"`
}

func WordepressMarkdown(input []byte) []byte {
	htmlFlags := blackfriday.HTML_USE_XHTML
	renderer := blackfriday.HtmlRenderer(htmlFlags, "", "")

	options := blackfriday.Options{
		Extensions: 0 |
			blackfriday.EXTENSION_NEWLINE_TO_SPACE |
			blackfriday.EXTENSION_FENCED_CODE,
	}

	return blackfriday.MarkdownOptions(input, renderer, options)
}

func transformBody(document wordepress.Document) string {
	html := WordepressMarkdown(document.Body())

	rewriteAnchors := func(bytes []byte) []byte {
		// This match must succeed or we wouldn't have been invoked
		href := string(AnchorRegexp.FindSubmatch(bytes)[1])

		// TODO extract string literals to config
		if strings.HasPrefix(href, "/site/") && strings.HasSuffix(href, ".md") {
			trimmed := strings.TrimPrefix(href, "/site/")
			trimmed = strings.TrimSuffix(trimmed, ".md")

			// Fully qualify each slug and reassemble the path
			var slugs []string
			for _, slug := range strings.Split(trimmed, "/") {
				slugs = append(slugs, fullyQualifiedSlug(slug))
			}

			return []byte(fmt.Sprintf(`<a href="/documentation/%s"`, strings.Join(slugs, "/")))
		}

		return bytes
	}

	rewriteImages := func(bytes []byte) []byte {
		// This match must succeed or we wouldn't have been invoked
		src := string(ImgRegexp.FindSubmatch(bytes)[1])

		qualifiedName := fmt.Sprintf("%s-%s-%s", product, version, src)

		if err := postAttachment(stdpath.Join(document.Dir(), src), qualifiedName); err != nil {
			log.Printf("Error posting attachment: %v", err)
			return bytes
		}

		return []byte(fmt.Sprintf(`<img src="/wp-content/uploads/%s"`, qualifiedName))
	}

	html = AnchorRegexp.ReplaceAllFunc(html, rewriteAnchors)
	html = ImgRegexp.ReplaceAllFunc(html, rewriteImages)
	return string(html)
}

func fullyQualifiedSlug(slug string) string {
	return fmt.Sprintf("%s-%s-%s", product, version, slug)
}

func sanitiseSlug(slug string) string {
	// WordPress does much more sanitisation than this, but the dots in the
	// embedded version strings are the main thing we're concerned with
	return strings.Replace(slug, ".", "-", -1)
}

func postAttachment(path, name string) error {
	body, err := os.Open(path)
	if err != nil {
		return err
	}

	url := baseURL + "/wp-json/wp/v2/media"
	request, err := http.NewRequest("POST", url, body)
	request.SetBasicAuth(user, password)
	request.Header.Set("Content-Type", "image/png")
	request.Header.Set("Content-Disposition", `attachment; filename="`+name+`"`)
	request.Header.Set("Accept", "application/json")

	log.Printf("Uploading image: %s", name)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	responseBytes, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("post failed: %v %s", response.Status, string(responseBytes))
	}

	var media Media
	err = json.Unmarshal(responseBytes, &media)
	if err != nil {
		return err
	}

	if media.MediaDetails.File != name {
		return fmt.Errorf("duplicate attachment: requested %s, response %s",
			name, media.MediaDetails.File)
	}

	return nil
}

func postDocument(parent int, document wordepress.Document) (int, error) {
	postRequest := PostRequest{
		Parent:    parent,
		MenuOrder: document.MenuOrder(),
		Slug:      fullyQualifiedSlug(document.Slug()),
		Title:     document.Title(),
		Content:   transformBody(document),
		Status:    "publish",
		Product:   product,
		Version:   version,
	}

	requestBytes, err := json.Marshal(postRequest)
	if err != nil {
		return 0, err
	}

	url := baseURL + "/wp-json/wp/v2/documentation"
	request, err := http.NewRequest("POST", url, bytes.NewReader(requestBytes))
	request.SetBasicAuth(user, password)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return 0, err
	}

	responseBytes, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return 0, err
	}

	if response.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("post failed: %v %s", response.Status, string(responseBytes))
	}

	var postResponse PostResponse
	err = json.Unmarshal(responseBytes, &postResponse)
	if err != nil {
		return 0, err
	}

	// Ensure server honoured our slug
	if postResponse.Slug != sanitiseSlug(postRequest.Slug) {
		return 0, fmt.Errorf("duplicate slug: requested %s, response %s",
			postRequest.Slug, postResponse.Slug)
	}

	return postResponse.Id, nil
}

func postDocuments(parent int, documents []wordepress.Document) error {
	for _, document := range documents {
		log.Printf("Uploading document: %s", document.Title())
		id, err := postDocument(parent, document)
		if err != nil {
			return err
		}
		err = postDocuments(id, document.Children())
		if err != nil {
			return err
		}
	}
	return nil
}

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a site into WordPress",
	Long:  `Publish a site into WordPress`,
	Run: func(cmd *cobra.Command, args []string) {
		if product == "" || version == "" || user == "" || password == "" {
			cmd.Usage()
			os.Exit(1)
		}
		for _, path := range args {
			documents, err := wordepress.ParseSite(path)
			if err != nil {
				log.Fatalf("Error parsing site: %v", err)
			}
			err = postDocuments(0, documents)
			if err != nil {
				log.Fatalf("Error uploading documents: %v", err)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(publishCmd)
}
