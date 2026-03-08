package reader

import "regexp"

var reWikipediaRef = regexp.MustCompile(`\[\d+\]`)

func removeWikipediaReferences(input string) string {
	return reWikipediaRef.ReplaceAllString(input, "")
}
