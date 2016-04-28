package wordepress

import (
	"fmt"
	"github.com/weaveworks/blackfriday"
	stdpath "path"
	"regexp"
	"strings"
)

var AnchorRegexp = regexp.MustCompile(`<a href="([^#"]*)`)
var ImgRegexp = regexp.MustCompile(`<img src="([^"]*)"`)
var BackslashRegexp = regexp.MustCompile(`\\`)

func convertToHTML(input []byte) []byte {
	htmlFlags := blackfriday.HTML_USE_XHTML
	renderer := blackfriday.HtmlRenderer(htmlFlags, "", "")

	options := blackfriday.Options{
		Extensions: 0 |
			blackfriday.EXTENSION_NEWLINE_TO_SPACE |
			blackfriday.EXTENSION_FENCED_CODE,
	}

	return blackfriday.MarkdownOptions(input, renderer, options)
}

func rewrite(product, version, srcdir string, markdown []byte) ([]byte, []*Image, error) {
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
				slugs = append(slugs, qualifySlug(product, version, slug))
			}

			return []byte(fmt.Sprintf(`<a href="/documentation/%s`, strings.Join(slugs, "/")))
		}

		return bytes
	}

	var images []*Image
	var rewriteErr error
	rewriteImages := func(bytes []byte) []byte {
		// This match must succeed or we wouldn't have been invoked
		src := string(ImgRegexp.FindSubmatch(bytes)[1])

		image, err := ReadImage(stdpath.Join(srcdir, src))
		if err != nil {
			rewriteErr = err
			return bytes
		}

		images = append(images, image)

		return []byte(fmt.Sprintf(`<img src="/wp-content/uploads/%s%s"`, image.Hash, image.Extension))
	}

	html := convertToHTML(markdown)
	html = AnchorRegexp.ReplaceAllFunc(html, rewriteAnchors)
	html = ImgRegexp.ReplaceAllFunc(html, rewriteImages)
	html = BackslashRegexp.ReplaceAllLiteral(html, []byte(`&#092;`))

	return html, images, rewriteErr
}
