package main

import (
	"strings"
	"path"
)

func normalizePageLink(link string, withRootPrefix bool) string {
	// remove file extension
	normalizedLink := strings.TrimSuffix(link, path.Ext(link))

	// dashify
	normalizedLink = strings.Replace(normalizedLink, " ", "-", -1)

	// make sure at maximum one root slash is there
	normalizedLink = strings.TrimLeft(normalizedLink, "/")
	if withRootPrefix {
		return "/" + normalizedLink
	}
	return normalizedLink
}
