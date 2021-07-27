package comment

import (
	"clx/colors"
	"clx/core"
	s "clx/syntax"
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

const (
	singleSpace = " "
	doubleSpace = "  "
	tripleSpace = "   "
)

func ParseComment(c string, config *core.Config, availableCommentWidth int, availableScreenWidth int) string {
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

			paragraph = colors.Italic + colors.Dimmed + paragraph + colors.Normal

			indentBlock := " " + getIndentationSymbol(config.AltIndentBlock)
			padding := text.WrapPad(colors.Dimmed + indentBlock)
			wrappedAndPaddedComment, _ := text.Wrap(paragraph, availableCommentWidth, padding)
			paragraph = wrappedAndPaddedComment

		case s.IsCodeBlock:
			paragraph = replaceHTML(paragraph)

			paddingWithBlock := text.WrapPad("")
			wrappedAndPaddedComment, _ := text.Wrap(paragraph, availableScreenWidth, paddingWithBlock)

			codeLines := strings.Split(wrappedAndPaddedComment, colors.NewLine)
			formattedCodeLines := ""

			for j, codeLine := range codeLines {
				isOnLastLine := j == len(codeLines)-1

				if isOnLastLine {
					formattedCodeLines += colors.Dimmed + codeLine + colors.Normal

					break
				}

				formattedCodeLines += colors.Dimmed + codeLine + colors.Normal + colors.NewLine
			}

			paragraph = formattedCodeLines

		default:
			paragraph = replaceSymbols(paragraph)
			paragraph = replaceSmileys(paragraph, config.EmojiSmileys)

			paragraph = replaceHTML(paragraph)
			paragraph = strings.TrimLeft(paragraph, " ")
			paragraph = highlightCommentSyntax(paragraph, config.HighlightComments)

			paragraph = trimURLs(paragraph, config.HighlightComments)

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
	paragraph = strings.ReplaceAll(paragraph, tripleSpace, singleSpace)
	paragraph = strings.ReplaceAll(paragraph, doubleSpace, singleSpace)
	paragraph = strings.ReplaceAll(paragraph, "...", "…")
	paragraph = replaceDoubleDashes(paragraph)
	paragraph = strings.ReplaceAll(paragraph, "CO2", "CO₂")
	paragraph = replaceFractions(paragraph)

	return paragraph
}

func replaceDoubleDashes(paragraph string) string {
	paragraph = strings.ReplaceAll(paragraph, " -- ", " — ")

	exp := regexp.MustCompile(`([a-zA-Z])--([a-zA-Z])`)

	return exp.ReplaceAllString(paragraph, `$1`+"—"+`$2`)
}

func replaceFractions(paragraph string) string {
	paragraph = strings.ReplaceAll(paragraph, " 1/2", " ½")
	paragraph = strings.ReplaceAll(paragraph, " 1/3", " ⅓")
	paragraph = strings.ReplaceAll(paragraph, " 2/3", " ⅔")
	paragraph = strings.ReplaceAll(paragraph, " 1/4", " ¼")
	paragraph = strings.ReplaceAll(paragraph, " 3/4", " ¾")
	paragraph = strings.ReplaceAll(paragraph, " 1/5", " ⅕")
	paragraph = strings.ReplaceAll(paragraph, " 2/5", " ⅖")
	paragraph = strings.ReplaceAll(paragraph, " 3/5", " ⅗")
	paragraph = strings.ReplaceAll(paragraph, " 4/5", " ⅘")
	paragraph = strings.ReplaceAll(paragraph, " 1/6", " ⅙")
	paragraph = strings.ReplaceAll(paragraph, " 1/10", " ⅒ ")

	paragraph = strings.ReplaceAll(paragraph, "1/5th", "⅕ th")
	paragraph = strings.ReplaceAll(paragraph, "1/6th", "⅙ th")
	paragraph = strings.ReplaceAll(paragraph, "1/10th", "⅒ th")

	return paragraph
}

func replaceSmileys(paragraph string, emojiSmiley bool) string {
	if !emojiSmiley {
		return paragraph
	}

	paragraph = strings.ReplaceAll(paragraph, " :)", " 😊")
	paragraph = strings.ReplaceAll(paragraph, " (:", " 😊")
	paragraph = strings.ReplaceAll(paragraph, " :-)", " 😊")
	paragraph = strings.ReplaceAll(paragraph, " :D", " 😄")
	paragraph = strings.ReplaceAll(paragraph, " =)", " 😃")
	paragraph = strings.ReplaceAll(paragraph, " =D", " 😃")
	paragraph = strings.ReplaceAll(paragraph, " ;)", " 😉")
	paragraph = strings.ReplaceAll(paragraph, " ;-)", " 😉")
	paragraph = strings.ReplaceAll(paragraph, " :P", " 😜")
	paragraph = strings.ReplaceAll(paragraph, " ;P", " 😜")
	paragraph = strings.ReplaceAll(paragraph, " :o", " 😮")
	paragraph = strings.ReplaceAll(paragraph, " :O", " 😮")
	paragraph = strings.ReplaceAll(paragraph, " :(", " 😔")
	paragraph = strings.ReplaceAll(paragraph, " :-(", " 😔")
	paragraph = strings.ReplaceAll(paragraph, " :/", " 😕")
	paragraph = strings.ReplaceAll(paragraph, " :-/", " 😕")

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

	return colors.NewParagraph
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

	input = strings.ReplaceAll(input, "<p>", colors.NewParagraph)
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
	input = highlightReferences(input)
	input = s.HighlightYCStartups(input)

	return input
}

func highlightReferences(input string) string {
	input = strings.ReplaceAll(input, "[0]", "["+colors.ToWhite("0")+"]")
	input = strings.ReplaceAll(input, "[1]", "["+colors.ToRed("1")+"]")
	input = strings.ReplaceAll(input, "[2]", "["+colors.ToYellow("2")+"]")
	input = strings.ReplaceAll(input, "[3]", "["+colors.ToGreen("3")+"]")
	input = strings.ReplaceAll(input, "[4]", "["+colors.ToBlue("4")+"]")
	input = strings.ReplaceAll(input, "[5]", "["+colors.ToCyan("5")+"]")
	input = strings.ReplaceAll(input, "[6]", "["+colors.ToMagenta("6")+"]")
	input = strings.ReplaceAll(input, "[7]", "["+colors.ToBrightWhite("7")+"]")
	input = strings.ReplaceAll(input, "[8]", "["+colors.ToBrightRed("8")+"]")
	input = strings.ReplaceAll(input, "[9]", "["+colors.ToBrightYellow("9")+"]")
	input = strings.ReplaceAll(input, "[10]", "["+colors.ToBrightGreen("10")+"]")

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

	exp := regexp.MustCompile(`(\$+[a-zA-Z]+)`)

	return exp.ReplaceAllString(input, colors.Cyan+`$1`+colors.Normal)
}

func highlightAbbreviations(input string) string {
	iAmNotALawyer := "IANAL"
	iAmALawyer := "IAAL"

	input = strings.ReplaceAll(input, iAmNotALawyer, colors.Red+iAmNotALawyer+colors.Normal)
	input = strings.ReplaceAll(input, iAmALawyer, colors.Green+iAmALawyer+colors.Normal)

	return input
}
