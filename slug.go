package wordepress

import (
	"fmt"
	"strings"
)

func sanitiseSlug(slug string) string {
	// WordPress does much more sanitisation than this, but the dots in the
	// embedded version strings are the main thing we're concerned with
	return strings.Replace(slug, ".", "-", -1)
}

func qualifySlug(product, version, base string) string {
	return fmt.Sprintf("%s-%s-%s", product, version, base)
}
