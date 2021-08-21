package preprocessor

import (
	"clx/markdown"
	"strings"
)

func ConvertItalicTags(text string) string {
	text = strings.ReplaceAll(text, "<i>", markdown.ItalicStart)
	text = strings.ReplaceAll(text, "</i>", markdown.ItalicStop)
	text = strings.ReplaceAll(text, "<em>", markdown.ItalicStart)
	text = strings.ReplaceAll(text, "</em>", markdown.ItalicStop)

	return text
}
