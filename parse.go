package wordepress

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	stdpath "path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var AttributeRegexp = regexp.MustCompile(`^([[:word:]]+):[[:space:]]*(.+?)[[:space:]]*$`)

func parseReader(reader io.Reader) (map[string]string, []byte, error) {
	scanner := bufio.NewScanner(reader)

	if !scanner.Scan() || scanner.Text() != "---" {
		return nil, nil, fmt.Errorf("missing delimiter parsing header")
	}

	attributes := make(map[string]string)
	for {
		if !scanner.Scan() {
			return nil, nil, fmt.Errorf("unexpected EOF parsing header")
		}
		if scanner.Text() == "---" {
			break
		}
		matches := AttributeRegexp.FindStringSubmatch(scanner.Text())
		if matches == nil {
			return nil, nil, fmt.Errorf("unable to parse header attribute: %v", scanner.Text())
		}
		attributes[matches[1]] = matches[2]
	}

	var body []byte
	for scanner.Scan() {
		body = append(body, scanner.Bytes()...)
		body = append(body, '\n')
	}

	return attributes, body, nil
}

func validateAttributes(attributes map[string]string) (string, int, error) {
	title := attributes["title"]
	if title == "" {
		return "", 0, fmt.Errorf("missing or empty title attribute")
	}
	delete(attributes, "title")

	menuOrder, err := strconv.Atoi(attributes["menu_order"])
	if err != nil {
		return "", 0, fmt.Errorf(`invalid menu_order: "%s"`, attributes["menu_order"])
	}
	delete(attributes, "menu_order")

	if len(attributes) > 0 {
		return "", 0, fmt.Errorf("unknown attributes: %v", attributes)
	}

	return title, menuOrder, nil
}

func parseFile(product, version, path string, parent *Document) (*Document, []*Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("error open path: %v", err)
	}
	defer file.Close()

	attributes, markdown, err := parseReader(file)
	if err != nil {
		return nil, nil, err
	}

	title, menuOrder, err := validateAttributes(attributes)
	if err != nil {
		return nil, nil, err
	}

	content, images, err := rewrite(product, version, stdpath.Dir(path), markdown)
	if err != nil {
		return nil, nil, err
	}

	base := strings.TrimSuffix(stdpath.Base(path), ".md")
	slug := qualifySlug(product, version, base)

	return &Document{
		LocalParent: parent,
		Title:       Text{Raw: title},
		MenuOrder:   menuOrder,
		Product:     product,
		Version:     version,
		Slug:        slug,
		Content:     Text{Raw: string(content)},
		Status:      "publish"}, images, nil
}

func recursiveParseSite(product, version, path string, parent *Document) ([]*Document, []*Image, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, nil, err
	}
	if !fileInfo.IsDir() {
		return nil, nil, fmt.Errorf("path %v is not a directory", path)
	}

	glob := fmt.Sprintf("%s/*.md", path)
	files, err := filepath.Glob(glob)
	if err != nil {
		return nil, nil, err
	}

	log.Printf("Loading %d markdown files from %s", len(files), path)

	var documents []*Document
	var siteImages []*Image
	for _, file := range files {
		document, images, err := parseFile(product, version, file, parent)
		if err != nil {
			return nil, nil, fmt.Errorf("parse %v: %v", file, err)
		}

		documents = append(documents, document)
		siteImages = append(siteImages, images...)

		childPath := strings.TrimSuffix(file, ".md")
		if _, err := os.Stat(childPath); err == nil {
			children, images, err := recursiveParseSite(product, version, childPath, document)
			if err != nil {
				return nil, nil, err
			}
			documents = append(documents, children...)
			siteImages = append(siteImages, images...)
		}
	}

	return documents, siteImages, nil
}

func ParseSite(product, version, path string) ([]*Document, []*Image, error) {
	return recursiveParseSite(product, version, path, nil)
}
