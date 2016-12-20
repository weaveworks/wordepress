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

func parseFile(product, version, tag, path string, parent *CustomPost) (*CustomPost, []*Image, error) {
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

	content, images, err := rewrite(product, version, tag, stdpath.Dir(path), markdown)
	if err != nil {
		return nil, nil, err
	}

	base := strings.TrimSuffix(stdpath.Base(path), ".md")
	slug, err := sanitiseSlug(qualifySlug(product, tag, base))
	if err != nil {
		return nil, nil, err
	}

	return &CustomPost{
		LocalParent: parent,
		Title:       Text{Raw: title},
		MenuOrder:   menuOrder,
		Product:     product,
		Version:     version,
		Name:        base,
		Tag:         tag,
		Slug:        slug,
		Content:     Text{Raw: string(content)},
		Status:      "publish"}, images, nil
}

func recursiveParseDir(product, version, tag, path string, parent *CustomPost) ([]*CustomPost, []*Image, error) {
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

	var posts []*CustomPost
	var siteImages []*Image
	for _, file := range files {
		post, images, err := parseFile(product, version, tag, file, parent)
		if err != nil {
			return nil, nil, fmt.Errorf("parse %v: %v", file, err)
		}

		posts = append(posts, post)
		siteImages = append(siteImages, images...)

		childPath := strings.TrimSuffix(file, ".md")
		if _, err := os.Stat(childPath); err == nil {
			children, images, err := recursiveParseDir(product, version, tag, childPath, post)
			if err != nil {
				return nil, nil, err
			}
			posts = append(posts, children...)
			siteImages = append(siteImages, images...)
		}
	}

	return posts, siteImages, nil
}

func ParseDir(product, version, tag, path string) ([]*CustomPost, []*Image, error) {
	return recursiveParseDir(product, version, tag, path, nil)
}
