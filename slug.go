package wordepress

import (
	"fmt"
	"regexp"
	"strings"
)

var safeSlugRegexp = regexp.MustCompile(`^[a-z0-9-.]+$`)

func sanitiseSlug(slug string) (string, error) {
	// WordPress slug sanitisation is far more extensive than this, however in
	// practice the only unusual characters we're likely to encounter here are
	// periods embedded in the slug as part of a product version string. The
	// regexp prevents us from causing downstream errors in the event that
	// we're passed something we haven't been taught to deal with yet.

	if !safeSlugRegexp.Match([]byte(slug)) {
		return "", fmt.Errorf(`slug sanitisation failure: unknown characters: "%s"`, slug)
	}

	return strings.Replace(slug, ".", "-", -1), nil
}

func qualifySlug(product, version, base string) string {
	return fmt.Sprintf("%s-%s-%s", product, version, base)
}
