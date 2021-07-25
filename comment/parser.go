package comment

import (
	"clx/colors"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/exp/utf8string"

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

func ParseComment(c string, commentWidth int, availableScreenWidth int, commentHighlighting bool,
	useAlternateIndent bool, emojiSmiley bool) string {
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
			paragraph = replaceSmileys(paragraph, emojiSmiley)

			paragraph = strings.Replace(paragraph, ">>", "", 1)
			paragraph = strings.Replace(paragraph, ">", "", 1)
			paragraph = strings.TrimLeft(paragraph, " ")
			paragraph = trimURLs(paragraph)

			paragraph = colors.Italic + colors.Dimmed + paragraph + colors.Normal

			indentBlock := " " + getIndentationSymbol(useAlternateIndent)
			padding := text.WrapPad(colors.Dimmed + indentBlock)
			wrappedAndPaddedComment, _ := text.Wrap(paragraph, commentWidth, padding)
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
			paragraph = replaceSmileys(paragraph, emojiSmiley)

			paragraph = replaceHTML(paragraph)
			paragraph = strings.TrimLeft(paragraph, " ")
			paragraph = highlightCommentSyntax(paragraph, commentHighlighting)

			paragraph = trimURLs(paragraph)

			padding := text.WrapPad("")
			wrappedAndPaddedComment, _ := text.Wrap(paragraph, commentWidth, padding)
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
	paragraph = strings.ReplaceAll(paragraph, " -- ", " — ")
	paragraph = strings.ReplaceAll(paragraph, " 1/2", " ½ ")
	paragraph = strings.ReplaceAll(paragraph, " 1/3", " ⅓ ")
	paragraph = strings.ReplaceAll(paragraph, " 2/3", " ⅔ ")
	paragraph = strings.ReplaceAll(paragraph, " 1/4", " ¼ ")
	paragraph = strings.ReplaceAll(paragraph, " 3/4", " ¾ ")
	paragraph = strings.ReplaceAll(paragraph, " 1/5", " ⅕ ")
	paragraph = strings.ReplaceAll(paragraph, " 2/5", " ⅖ ")
	paragraph = strings.ReplaceAll(paragraph, " 3/5", " ⅗ ")
	paragraph = strings.ReplaceAll(paragraph, " 4/5", " ⅘ ")
	paragraph = strings.ReplaceAll(paragraph, " 1/6", " ⅙ ")
	paragraph = strings.ReplaceAll(paragraph, " 1/10", " ⅒ ")

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
	input = strings.ReplaceAll(input, "</a>", colors.Normal)
	input = strings.ReplaceAll(input, "<pre><code>", "")
	input = strings.ReplaceAll(input, "</code></pre>", "")

	return input
}

func highlightCommentSyntax(input string, commentHighlighting bool) string {
	if !commentHighlighting {
		return input
	}

	input = highlightUrls(input)
	input = highlightBackticks(input)
	input = highlightMentions(input)
	input = highlightVariables(input)
	input = highlightAbbreviations(input)
	input = highlightReferences(input)

	return input
}

func highlightUrls(input string) string {
	input = strings.ReplaceAll(input, "https://", colors.Blue)
	input = strings.ReplaceAll(input, "http://", colors.Blue)

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

func trimURLs(comment string) string {
	expression := regexp.MustCompile(`<a href=".*?" rel="nofollow">`)

	return expression.ReplaceAllString(comment, "")
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
	numberOfMentions := strings.Count(input, "@")

	if numberOfMentions == 0 {
		return input
	}

	output := ""
	words := strings.Split(input, " ")

	for _, word := range words {
		isScreenResolution := strings.HasPrefix(word, "@60Hz") ||
			strings.HasPrefix(word, "@144Hz") ||
			strings.HasPrefix(word, "@120Hz")

		wordIsSingleAtSign := word == "@"

		switch {
		case isScreenResolution || wordIsSingleAtSign:
			output += word + " "

		case strings.HasPrefix(word, "@dang"):
			mention := colors.Green + word + colors.Normal + " "
			mention = strings.ReplaceAll(mention, ",", colors.Normal+",")
			mention = strings.ReplaceAll(mention, ".", colors.Normal+".")

			output += mention

		case strings.HasPrefix(word, "@"):
			mention := colors.Yellow + word + colors.Normal + " "
			mention = strings.ReplaceAll(mention, ",", colors.Normal+",")
			mention = strings.ReplaceAll(mention, ".", colors.Normal+".")

			output += mention

		default:
			output += word + " "
		}
	}

	return output
}

func highlightVariables(input string) string {
	backtick := "`"
	numberOfBackticks := strings.Count(input, backtick)

	// Highlighting variables inside commands marked with backticks
	// messes with the formatting. If there are both backticks and variables
	// in the comment, we give priority to the backticks.
	if numberOfBackticks != 0 {
		return input
	}

	numberOfDollarSigns := strings.Count(input, "$")

	if numberOfDollarSigns == 0 {
		return input
	}

	output := ""
	words := strings.Split(input, " ")

	for _, word := range words {
		currentWord := utf8string.NewString(word)
		wordHasOnlyOneCharacter := currentWord.RuneCount() == 1

		if word == "$" || word == "" || wordHasOnlyOneCharacter {
			output += word + " "

			continue
		}

		s := utf8string.NewString(word)
		secondRune := s.At(1)

		switch {
		case strings.HasPrefix(word, "$") && unicode.IsLetter(secondRune):
			variable := colors.Cyan + word + colors.Normal + " "
			variable = strings.ReplaceAll(variable, "\"", colors.Normal+"\"")
			variable = strings.ReplaceAll(variable, "'", colors.Normal+"'")
			variable = strings.ReplaceAll(variable, "”", colors.Normal+"”")

			output += variable

		default:
			output += word + " "
		}
	}

	return output
}

func highlightAbbreviations(input string) string {
	iAmNotALawyer := "IANAL"
	iAmALawyer := "IAAL"

	input = strings.ReplaceAll(input, iAmNotALawyer, colors.Red+iAmNotALawyer+colors.Normal)
	input = strings.ReplaceAll(input, iAmALawyer, colors.Green+iAmALawyer+colors.Normal)

	return input
}
