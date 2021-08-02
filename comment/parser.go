package comment

import (
	"clx/colors"
	"clx/core"
	"clx/indent"
	"clx/syntax"
	"regexp"
	"strings"

	text "github.com/MichaelMure/go-term-text"
)

type Comment struct {
	Sections []*section
}

type section struct {
	IsCodeBlock bool
	IsQuote     bool
	Text        string
}

func ParseComment(c string, config *core.Config, availableCommentWidth int, availableScreenWidth int) string {
	if c == "[deleted]" {
		return colors.Dimmed + "[deleted]" + colors.Normal
	}

	c = strings.Replace(c, "<p>", "", 1)
	c = strings.ReplaceAll(c, "\n</code></pre>\n", "<p>")
	paragraphs := strings.Split(c, "<p>")

	comment := new(Comment)
	comment.Sections = make([]*section, len(paragraphs))

	for i, paragraph := range paragraphs {
		s := new(section)
		s.Text = replaceCharacters(paragraph)

		if strings.Contains(s.Text, "<pre><code>") {
			s.IsCodeBlock = true
		}

		if isQuote(s.Text) {
			s.IsQuote = true
		}

		comment.Sections[i] = s
	}

	output := ""

	for i, s := range comment.Sections {
		paragraph := s.Text

		switch {
		case s.IsQuote:
			paragraph = strings.ReplaceAll(paragraph, "<i>", "")
			paragraph = strings.ReplaceAll(paragraph, "</i>", "")
			paragraph = strings.ReplaceAll(paragraph, "</a>", colors.Normal+colors.Dimmed+colors.Italic)
			paragraph = replaceSymbols(paragraph)
			paragraph = replaceSmileys(paragraph, config.EmojiSmileys)

			paragraph = strings.Replace(paragraph, ">>", "", 1)
			paragraph = strings.Replace(paragraph, ">", "", 1)
			paragraph = strings.TrimLeft(paragraph, " ")
			paragraph = trimURLs(paragraph, false)
			paragraph = syntax.RemoveUnwantedNewLines(paragraph)
			paragraph = syntax.RemoveUnwantedWhitespace(paragraph)

			paragraph = colors.Italic + colors.Dimmed + paragraph + colors.Normal

			quoteIndent := " " + indent.GetIndentSymbol(false, config.AltIndentBlock)
			padding := text.WrapPad(colors.Dimmed + quoteIndent)
			wrappedAndPaddedComment, _ := text.Wrap(paragraph, availableCommentWidth, padding)
			paragraph = wrappedAndPaddedComment

		case s.IsCodeBlock:
			paragraph = replaceHTML(paragraph)

			paddingWithBlock := text.WrapPad("")
			wrappedAndPaddedComment, _ := text.Wrap(paragraph, availableScreenWidth, paddingWithBlock)

			codeLines := strings.Split(wrappedAndPaddedComment, newLine)
			formattedCodeLines := ""

			for j, codeLine := range codeLines {
				isOnLastLine := j == len(codeLines)-1

				if isOnLastLine {
					formattedCodeLines += colors.Dimmed + codeLine + colors.Normal

					break
				}

				formattedCodeLines += colors.Dimmed + codeLine + colors.Normal + newLine
			}

			paragraph = formattedCodeLines

		default:
			paragraph = replaceSymbols(paragraph)
			paragraph = replaceSmileys(paragraph, config.EmojiSmileys)

			paragraph = replaceHTML(paragraph)
			paragraph = strings.TrimLeft(paragraph, " ")
			paragraph = highlightCommentSyntax(paragraph, config.HighlightComments)

			paragraph = trimURLs(paragraph, config.HighlightComments)
			paragraph = syntax.RemoveUnwantedNewLines(paragraph)
			paragraph = syntax.RemoveUnwantedWhitespace(paragraph)

			padding := text.WrapPad("")
			wrappedAndPaddedComment, _ := text.Wrap(paragraph, availableCommentWidth, padding)
			paragraph = wrappedAndPaddedComment
		}

		separator := getParagraphSeparator(i, len(comment.Sections))
		output += paragraph + separator
	}

	return output
}

func replaceSymbols(paragraph string) string {
	paragraph = strings.ReplaceAll(paragraph, "...", "…")
	paragraph = replaceDoubleDashes(paragraph)
	paragraph = strings.ReplaceAll(paragraph, "CO2", "CO₂")
	paragraph = syntax.ConvertFractions(paragraph)

	return paragraph
}

func replaceDoubleDashes(paragraph string) string {
	paragraph = strings.ReplaceAll(paragraph, " -- ", " — ")

	exp := regexp.MustCompile(`([a-zA-Z])--([a-zA-Z])`)

	return exp.ReplaceAllString(paragraph, `$1`+"—"+`$2`)
}

func replaceSmileys(paragraph string, emojiSmiley bool) string {
	if !emojiSmiley {
		return paragraph
	}

	paragraph = syntax.ConvertSmileys(paragraph)

	return paragraph
}

func isQuote(text string) bool {
	quoteMark := ">"

	return strings.HasPrefix(text, quoteMark) ||
		strings.HasPrefix(text, " "+quoteMark) ||
		strings.HasPrefix(text, "<i>"+quoteMark) ||
		strings.HasPrefix(text, "<i> "+quoteMark)
}

func getParagraphSeparator(index int, sliceLength int) string {
	isAtLastParagraph := index == sliceLength-1

	if isAtLastParagraph {
		return ""
	}

	return newParagraph
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

	input = strings.ReplaceAll(input, "<p>", newParagraph)
	input = strings.ReplaceAll(input, "<i>", colors.Italic)
	input = strings.ReplaceAll(input, "</i>", colors.Normal)
	input = strings.ReplaceAll(input, "</a>", "")
	input = strings.ReplaceAll(input, "<pre><code>", "")
	input = strings.ReplaceAll(input, "</code></pre>", "")

	return input
}

func highlightCommentSyntax(input string, commentHighlighting bool) string {
	if !commentHighlighting {
		return input
	}

	input = highlightBackticks(input)
	input = highlightMentions(input)
	input = highlightVariables(input)
	input = highlightAbbreviations(input)
	input = syntax.HighlightReferences(input)
	input = syntax.HighlightYCStartups(input)

	return input
}

func trimURLs(comment string, highlightComment bool) string {
	expression := regexp.MustCompile(`<a href=".*?" rel="nofollow">`)

	if !highlightComment {
		return expression.ReplaceAllString(comment, "")
	}

	comment = expression.ReplaceAllString(comment, "")

	e := regexp.MustCompile(`https?://([^,"\) \n]+)`)
	comment = e.ReplaceAllString(comment, colors.Blue+`$1`+colors.Normal)

	comment = strings.ReplaceAll(comment, "."+colors.Normal+" ", colors.Normal+"."+" ")

	return comment
}

func highlightBackticks(input string) string {
	backtick := "`"
	numberOfBackticks := strings.Count(input, backtick)
	numberOfBackticksIsOdd := numberOfBackticks%2 != 0

	if numberOfBackticks == 0 || numberOfBackticksIsOdd {
		return input
	}

	isOnFirstBacktick := true

	for i := 0; i < numberOfBackticks+1; i++ {
		if isOnFirstBacktick {
			input = strings.Replace(input, backtick, colors.Italic+colors.Magenta, 1)
		} else {
			input = strings.Replace(input, backtick, colors.Normal, 1)
		}

		isOnFirstBacktick = !isOnFirstBacktick
	}

	return input
}

func highlightMentions(input string) string {
	exp := regexp.MustCompile(`((?:^| )\B@[\w.]+)`)
	input = exp.ReplaceAllString(input, colors.Yellow+`$1`+colors.Normal)

	input = strings.ReplaceAll(input, colors.Yellow+"@dang", colors.Green+"@dang")
	input = strings.ReplaceAll(input, colors.Yellow+" @dang", colors.Green+" @dang")

	return input
}

func highlightVariables(input string) string {
	backtick := "`"
	numberOfBackticks := strings.Count(input, backtick)

	// Highlighting variables inside commands marked with backticks
	// messes with the formatting. If there are both backticks and variables
	// in the comment, we give priority to the backticks.
	if numberOfBackticks > 0 {
		return input
	}

	exp := regexp.MustCompile(`(\$+[a-zA-Z_\-]+)`)

	return exp.ReplaceAllString(input, colors.Cyan+`$1`+colors.Normal)
}

func highlightAbbreviations(input string) string {
	iAmNotALawyer := "IANAL"
	iAmALawyer := "IAAL"

	input = strings.ReplaceAll(input, iAmNotALawyer, colors.Red+iAmNotALawyer+colors.Normal)
	input = strings.ReplaceAll(input, iAmALawyer, colors.Green+iAmALawyer+colors.Normal)

	return input
}
