package article

import (
	"strings"

	"github.com/bensadeh/circumflex/layout"
)

func processArticle(text string, url string, width int) string {
	text = filterSite(text, url)
	text = indent(text, width)

	return text
}

func indent(commentSection string, _ int) string {
	indentBlock := strings.Repeat(" ", layout.ReaderViewLeftMargin)

	lines := strings.Split(commentSection, "\n")
	for i, line := range lines {
		lines[i] = indentBlock + line
	}

	return strings.Join(lines, "\n")
}
