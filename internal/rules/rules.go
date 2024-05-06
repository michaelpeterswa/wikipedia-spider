package rules

import "strings"

// IsValidLink checks if a link is valid for the spider to visit
//
// Most of these are covered by robots.txt, but this is a good preventative measure
func IsValidLink(link string) bool {
	hasPrefixes := []string{
		"/wiki/",
	}

	notHasPrefixes := []string{
		"/wiki/Category:",
		"/wiki/Special:",
		"/wiki/Help:",
		"/wiki/Wikipedia:",
		"/wiki/Template:",
		"/wiki/Template_talk:",
		"/wiki/File:",
		"#cite",
	}

	hasPrefix := false
	for _, prefix := range hasPrefixes {
		if strings.HasPrefix(link, prefix) {
			hasPrefix = true
		}
	}

	notHasPrefix := true
	for _, prefix := range notHasPrefixes {
		if strings.HasPrefix(link, prefix) {
			notHasPrefix = false
		}
	}

	if hasPrefix && notHasPrefix {
		return true
	}
	return false
}
