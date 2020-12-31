package tool

import (
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"strings"
	"unicode"
)

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r)
}

var transformer = transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)

// SearchLib generate a matching friendly string (removing accents, trailing spaces, ...)
func SearchLib(lib string) string {
	lib = strings.ToLower(strings.TrimSpace(lib))

	result, _, err := transform.String(transformer, lib)
	if err != nil {
		return lib
	}
	return result
}
