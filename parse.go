package wordepress

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	stdpath "path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
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

const (
	templateStage1 = iota
	templateStage2
)

func parseTemplate(stage int, body []byte, sourceFileDir string) ([]byte, error) {
	var (
		data   interface{}
		right  string
		left   string
		output bytes.Buffer
	)

	defaultTagAttributes := map[string][]string{
		"details": {
			"style='margin-left: 1em; border-left: 1px solid gray; padding-left: 1em;'",
		},
	}

	openTagFunc := func(tag string, attributes ...string) string {
		if len(attributes) > 0 {
			return fmt.Sprintf("<%s %s>", tag, strings.Join(attributes, " "))
		}

		if v, ok := defaultTagAttributes[tag]; ok {
			return fmt.Sprintf("<%s %s>", tag, strings.Join(v, " "))
		}

		return fmt.Sprintf("<%s>", tag)
	}

	closeTagFunc := func(tag string) string {
		return fmt.Sprintf("</%s>", tag)
	}

	funcMap := template.FuncMap{
		"open_tag":  openTagFunc,
		"close_tag": closeTagFunc,
		"build_info": func() string {
			return os.Getenv("WORDEPRESS_CI_INFO")
		},
		"include": func(filePath string) (string, error) {
			content, err := ioutil.ReadFile(stdpath.Join(sourceFileDir, filePath))
			if err != nil {
				return "", err
			}

			return string(content), nil
		},
	}

	for _, t := range []string{"div", "details"} {
		func(t string) {
			funcMap["open_"+t] = func(attributes ...string) string { return openTagFunc(t, attributes...) }
			funcMap["close_"+t] = func() string { return closeTagFunc(t) }
		}(t)
	}

	switch stage {
	case templateStage1:
		right, left = "{{", "}}"
	case templateStage2:
		right, left = "[[", "]]"
	}

	data = nil

	t, err := template.New(fmt.Sprintf("stage%d", stage)).
		Delims(right, left).
		Funcs(funcMap).
		Parse(string(body))

	if err != nil {
		return nil, err
	}

	if err := t.Execute(&output, data); err != nil {
		return nil, err
	}

	return output.Bytes(), nil
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

	dir := stdpath.Dir(path)

	attributes, markdown, err := parseReader(file)
	if err != nil {
		return nil, nil, err
	}

	title, menuOrder, err := validateAttributes(attributes)
	if err != nil {
		return nil, nil, err
	}

	// includes and other macros can be done at pre-process stage, after that we render markdown
	// pre-processing stage uses standard `{{ ... }}` template expression delimiters
	// and the post-processing stage uses `[[ ... ]]` delimiters
	// backslash-escaped backtick quotes must be used inside post-processing macors, otherwise
	// the markdown library turns double qoutes into `&quot;` and makes template parser unhappy
	preProcessed, err := parseTemplate(templateStage1, markdown, dir)
	if err != nil {
		return nil, nil, err
	}

	content, images, err := rewrite(product, version, tag, dir, preProcessed)
	if err != nil {
		return nil, nil, err
	}

	// and after we have rendered some html, we re-execute the template processor
	postProcessed, err := parseTemplate(templateStage2, content, dir)
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
		Content:     Text{Raw: string(postProcessed)},
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
