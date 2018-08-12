package main

import (
	"strings"
)

func normalizePageLink(link string, withRootPrefix bool) string {
	normalizedLink := strings.TrimLeft(strings.Replace(link, " ", "-", -1), "/")
	if withRootPrefix {
		return "/" + normalizedLink
	}
	return normalizedLink
}
