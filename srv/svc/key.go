package svc

import (
	"strings"
)

func normalizeString(s string) string {
	return strings.TrimSpace(strings.TrimRight(s, "\r\n\x00"))
}
