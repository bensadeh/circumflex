package comment

import (
	"regexp"
	"strings"
)

type Comment struct {
	Sections []Section
}

type Section struct {
	IsCodeBlock bool
	IsQuote     bool
	Text        string
}

const (
	doubleSpace = "  "
	singleSpace = " "
)

func ParseComment(text string) (string, []string) {
	var URLs []string

	text = strings.Replace(text, "<p>", "", 1)
	paragraphs := strings.Split(text, "<p>")
	c := new(Comment)

	for _, paragraph := range paragraphs {
		section := new(Section)
		section.Text = replaceCharacters(paragraph)

		if strings.Contains(section.Text, "<pre><code>") {
			section.IsCodeBlock = true
		}

		if isQuote(section.Text) {
			section.IsQuote = true
		}

		c.Sections = append(c.Sections, *section)
	}

	output := ""

	for i, section := range c.Sections {
		switch {
		case section.IsQuote:
			section.Text = strings.ReplaceAll(section.Text, "<i>", "")
			section.Text = strings.ReplaceAll(section.Text, "</i>", "")
			section.Text = strings.ReplaceAll(section.Text, doubleSpace, singleSpace)
			section.Text = strings.ReplaceAll(section.Text, ">", "")
			section.Text = strings.ReplaceAll(section.Text, ">>", "")
			section.Text = strings.TrimLeft(section.Text, " ")

			section.Text = Italic + Dimmed + section.Text + Normal

		case section.IsCodeBlock:
			section.Text = replaceHTML(section.Text)

		default:
			section.Text = strings.ReplaceAll(section.Text, doubleSpace, singleSpace)
			section.Text = highlightReferences(section.Text)
			section.Text = replaceHTML(section.Text)
			URLs = append(URLs, extractURLs(section.Text)...)
			section.Text = trimURLs(section.Text)
		}

		separator := getSeparator(i, len(c.Sections))
		output += section.Text + separator

		//if section.IsQuote {
		//
		//}
		//
		//if !section.IsCodeBlock {
		//	section.Text = highlightReferences(section.Text)
		//	section.Text = strings.ReplaceAll(section.Text, doubleSpace, singleSpace)
		//}
		//
		//separator := getSeparator(i, len(c.Sections))
		//
		//section.Text = replaceHTML(section.Text)
		//URLs = append(URLs, extractURLs(section.Text)...)
		//section.Text = trimURLs(section.Text)
		//
		//output += section.Text + separator
	}

	return output, URLs
}

func isQuote(text string) bool {
	quoteMark := ">"

	return strings.HasPrefix(text, ""+quoteMark) ||
		strings.HasPrefix(text, " "+quoteMark) ||
		strings.HasPrefix(text, "<i>"+quoteMark) ||
		strings.HasPrefix(text, "<i> "+quoteMark)
}

func getSeparator(index int, sliceLength int) string {
	isAtLastParagraph := index == sliceLength-1

	if isAtLastParagraph {
		return ""
	}

	return NewParagraph
}

func replaceCharacters(input string) string {
	input = strings.ReplaceAll(input, "&#x27;", "'")
	input = strings.ReplaceAll(input, "&gt;", ">")
	input = strings.ReplaceAll(input, "&lt;", "<")
	input = strings.ReplaceAll(input, "&#x2F;", "/")
	input = strings.ReplaceAll(input, "&quot;", `"`)
	input = strings.ReplaceAll(input, "&amp;", "&")

	return input
}

func replaceHTML(input string) string {
	input = strings.Replace(input, "<p>", "", 1)

	input = strings.ReplaceAll(input, "<p>", NewParagraph)
	input = strings.ReplaceAll(input, "<i>", Italic)
	input = strings.ReplaceAll(input, "</i>", Normal)
	input = strings.ReplaceAll(input, "</a>", "")
	input = strings.ReplaceAll(input, "<pre><code>", Dimmed)
	input = strings.ReplaceAll(input, "</code></pre>", Normal)

	return input
}

func highlightReferences(input string) string {
	input = strings.ReplaceAll(input, "[0]", "["+white("0")+"]")
	input = strings.ReplaceAll(input, "[1]", "["+red("1")+"]")
	input = strings.ReplaceAll(input, "[2]", "["+yellow("2")+"]")
	input = strings.ReplaceAll(input, "[3]", "["+green("3")+"]")
	input = strings.ReplaceAll(input, "[4]", "["+blue("4")+"]")
	input = strings.ReplaceAll(input, "[5]", "["+cyan("5")+"]")
	input = strings.ReplaceAll(input, "[6]", "["+magenta("6")+"]")
	input = strings.ReplaceAll(input, "[7]", "["+altWhite("7")+"]")
	input = strings.ReplaceAll(input, "[8]", "["+altRed("8")+"]")
	input = strings.ReplaceAll(input, "[9]", "["+altYellow("9")+"]")
	input = strings.ReplaceAll(input, "[10]", "["+altGreen("10")+"]")

	return input
}

func extractURLs(input string) []string {
	expForFirstTag := regexp.MustCompile(`<a href=".*?" rel="nofollow">`)
	URLs := expForFirstTag.FindAllString(input, 10)

	for i := range URLs {
		URLs[i] = strings.ReplaceAll(URLs[i], `<a href="`, "")
		URLs[i] = strings.ReplaceAll(URLs[i], `" rel="nofollow">`, "")
	}

	return URLs
}

func trimURLs(comment string) string {
	expression := regexp.MustCompile(`<a href=".*?" rel="nofollow">`)

	return expression.ReplaceAllString(comment, "")
}
