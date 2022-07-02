package preprocessor

import (
	"strings"

	"clx/markdown"
)

func ConvertItalicTags(text string) string {
	text = strings.ReplaceAll(text, "<i>", markdown.ItalicStart)
	text = strings.ReplaceAll(text, "</i>", markdown.ItalicStop)
	text = strings.ReplaceAll(text, "<em>", markdown.ItalicStart)
	text = strings.ReplaceAll(text, "</em>", markdown.ItalicStop)

	return text
}

func ConvertBoldTags(text string) string {
	text = strings.ReplaceAll(text, "<b>", markdown.BoldStart)
	text = strings.ReplaceAll(text, "</b>", markdown.BoldStop)

	text = strings.ReplaceAll(text, "<strong>", markdown.BoldStart)
	text = strings.ReplaceAll(text, "</strong>", markdown.BoldStop)

	return text
}
