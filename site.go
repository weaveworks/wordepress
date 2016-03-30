package wordepress

import "log"
import "bufio"
import "fmt"
import "io"
import "os"
import "regexp"
import stdpath "path"
import "path/filepath"
import "strings"
import "strconv"

var AttributeRegexp = regexp.MustCompile(`^([[:word:]]+):[[:space:]]*(.+?)[[:space:]]*$`)

type Document interface {
	Title() string
	MenuOrder() int
	Slug() string
	Body() []byte
	Children() []Document
	Dir() string
}

type document struct {
	slug       string
	attributes map[string]string
	body       []byte
	children   []Document
	dir        string
}

func (d *document) Slug() string {
	return d.slug
}

func (d *document) Title() string {
	return d.attributes["title"]
}

func (d *document) MenuOrder() int {
	if val, ok := d.attributes["menu_order"]; ok {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return 0
}

func (d *document) Body() []byte {
	return d.body
}

func (d *document) Children() []Document {
	return d.children
}

func (d *document) Dir() string {
	return d.dir
}

func parseReader(reader io.Reader) (*document, error) {
	scanner := bufio.NewScanner(reader)

	if !scanner.Scan() || scanner.Text() != "---" {
		return nil, fmt.Errorf("missing delimiter parsing header")
	}

	attributes := make(map[string]string)
	for {
		if !scanner.Scan() {
			return nil, fmt.Errorf("unexpected EOF parsing header")
		}
		if scanner.Text() == "---" {
			break
		}
		matches := AttributeRegexp.FindStringSubmatch(scanner.Text())
		if matches == nil {
			return nil, fmt.Errorf("unable to parse header attribute: %v", scanner.Text())
		}
		attributes[matches[1]] = matches[2]
	}

	var body []byte
	for scanner.Scan() {
		body = append(body, scanner.Bytes()...)
		body = append(body, '\n')
	}

	return &document{"", attributes, body, nil, ""}, nil
}

func parseFile(path string) (*document, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error open path: %v", err)
	}
	defer file.Close()

	document, err := parseReader(file)
	if err != nil {
		return nil, err
	}

	document.dir = stdpath.Dir(path)
	document.slug = strings.TrimSuffix(stdpath.Base(path), ".md")

	return document, nil
}

func ParseSite(path string) ([]Document, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !fileInfo.IsDir() {
		return nil, fmt.Errorf("path %v is not a directory", path)
	}

	glob := fmt.Sprintf("%s/*.md", path)
	files, err := filepath.Glob(glob)
	if err != nil {
		return nil, err
	}

	log.Printf("Loading %d markdown files from %s", len(files), path)

	var documents []Document
	for _, file := range files {
		document, err := parseFile(file)
		if err != nil {
			return nil, fmt.Errorf("parse %v: %v", file, err)
		}

		childPath := strings.TrimSuffix(file, ".md")
		if _, err := os.Stat(childPath); err == nil {
			document.children, err = ParseSite(childPath)
			if err != nil {
				return nil, err
			}
		}

		documents = append(documents, document)
	}

	return documents, nil
}
