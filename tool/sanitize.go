package tool

import (
	"regexp"
	"strings"
)

var illegalCharacters = regexp.MustCompile(`[/\\:;&*"?<>{}\[\]|%#@=\^]`)

func SanitizeFilename(rawfilename string) string {

	// Remove all illegal characters
	cleanFilename := illegalCharacters.ReplaceAllString(strings.Trim(rawfilename, " "), "_")

	return cleanFilename
}
