package terminal

import (
	"regexp"
	"strings"

	"github.com/muesli/reflow/wordwrap"

	"clx/reader/markdown"

	"clx/constants/unicode"
	"clx/meta"
	"clx/syntax"

	"github.com/charmbracelet/glamour"

	terminal "github.com/wayneashleyberry/terminal-dimensions"

	termtext "github.com/MichaelMure/go-term-text"

	. "github.com/logrusorgru/aurora/v3"
)

const (
	indentLevel1 = "  "
	indentLevel2 = indentLevel1 + indentLevel1
	indentLevel3 = indentLevel2 + indentLevel1

	codeStart = "[CLX_CODE_START]"
	codeEnd   = "[CLX_CODE_END]"
)

func CreateHeader(title string, domain string, lineWidth int) string {
	return meta.GetReaderModeMetaBlock(title, domain, lineWidth)
}

func ConvertToTerminalFormat(blocks []*markdown.Block, lineWidth int, indentBlock string) string {
	output := ""

	for _, block := range blocks {
		switch block.Kind {
		case markdown.Text:
			output += renderText(block.Text, lineWidth) + "\n\n"

		case markdown.Image:
			output += renderImage(block.Text, lineWidth) + "\n\n"

		case markdown.Code:
			output += renderCode(block.Text) + "\n\n"

		case markdown.Quote:
			output += renderQuote(block.Text, lineWidth, indentBlock) + "\n\n"

		case markdown.Table:
			output += renderTable(block.Text) + "\n\n"

		case markdown.List:
			output += renderList(block.Text, lineWidth) + "\n\n"

		case markdown.Divider:
			output += renderDivider(lineWidth) + "\n\n"

		case markdown.H1:
			output += h1(block.Text, lineWidth) + "\n\n"

		case markdown.H2:
			output += h2(block.Text, lineWidth) + "\n\n"

		case markdown.H3:
			output += h3(block.Text, lineWidth) + "\n\n"

		case markdown.H4:
			output += h4(block.Text, lineWidth) + "\n\n"

		case markdown.H5:
			output += h5(block.Text, lineWidth) + "\n\n"

		case markdown.H6:
			output += h6(block.Text, lineWidth) + "\n\n"

		default:
			output += renderText(block.Text, lineWidth) + "\n\n"
		}
	}

	output = strings.TrimLeft(output, "\n")

	return output
}

func renderDivider(lineWidth int) string {
	divider := strings.Repeat("-", lineWidth-len(indentLevel1)*2)

	return Faint(indentLevel1 + divider).String()
}

func renderText(text string, lineWidth int) string {
	text = it(text)
	text = bld(text)
	text = removeHrefs(text)
	text = unescapeCharacters(text)
	text = removeImageReference(text)

	text = syntax.RemoveUnwantedNewLines(text)
	text = highlightBackticks(text)
	text = syntax.HighlightMentions(text)
	text = syntax.TrimURLs(text, true)

	return wordwrap.String(text, lineWidth)
}

func renderList(text string, lineWidth int) string {
	text = it(text)
	text = bld(text)
	text = removeImageReference(text)
	text = removeHrefs(text)
	text = unescapeCharacters(text)
	text = highlightBackticks(text)

	output := ""
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		exp := regexp.MustCompile(`^\s*(-|\d+\.)`)

		listToken := exp.FindString(line)
		listText := strings.TrimLeft(line, listToken)

		paddingBuffer := strings.Repeat(" ", len(listToken))
		padding := indentLevel1 + paddingBuffer + " "

		wrappedIndentedItem, _ := termtext.WrapWithPadIndent(listToken+listText, lineWidth, indentLevel1, padding)
		wrappedIndentedItem = insertSpaceAfterItemListSeparator(wrappedIndentedItem)

		output += wrappedIndentedItem + "\n"
	}

	output = replaceListPrefixes(output)
	output = trimLeadingZero(output)

	return strings.TrimRight(output, "\n")
}

func renderImage(text string, lineWidth int) string {
	red := "\u001B[31m"
	italic := "\u001B[3m"
	faint := "\u001B[2m"
	normal := "\u001B[0m"
	imageLabel := normal + Red(unicode.Block).Faint().String() + Yellow(unicode.Block).Faint().String() +
		Blue(unicode.Block).Faint().String() + normal + red + faint + italic + " Image " + normal + faint + italic

	text = regexp.MustCompile(`!\[(.*?)\]\(.*?\)$`).
		ReplaceAllString(text, imageLabel+`$1`)

	text = regexp.MustCompile(`!\[(.*?)\]\(.*?\)\s`).
		ReplaceAllString(text, imageLabel+`$1`)

	text = regexp.MustCompile(`!\[(.*?)\]\(.*?\)`).
		ReplaceAllString(text, imageLabel+`$1`)

	if text == imageLabel {
		return indentLevel2 + text + normal
	}

	lines := strings.Split(text, imageLabel)
	output := ""

	for _, line := range lines {
		if len(lines) == 1 || len(lines) == 0 {
			output += imageLabel + line + "\n\n"

			break
		}

		if line == "" {
			continue
		}

		output += imageLabel + line + "\n\n"
	}

	output = strings.TrimSuffix(output, "\n\n")
	output += normal

	output = it(output)
	output = bld(output)
	output = removeDoubleWhitespace(output)

	padding := termtext.WrapPad(indentLevel1)
	output, _ = termtext.Wrap(output, lineWidth, padding)

	return output
}

func renderCode(text string) string {
	screenWidth, _ := terminal.Width()

	text = strings.TrimSuffix(text, "\n")
	text = strings.TrimPrefix(text, "\n")

	text = Faint(text).String()
	text = removeHrefs(text)

	padding := termtext.WrapPad(indentLevel1)
	text, _ = termtext.Wrap(text, int(screenWidth), padding)

	return text
}

func renderQuote(text string, lineWidth int, indentSymbol string) string {
	text = Italic(text).Faint().String()
	text = unescapeCharacters(text)
	text = removeHrefs(text)

	indentBlock := " " + indentSymbol
	text = itReversed(text)
	text = bldInQuote(text)

	padding := termtext.WrapPad(indentLevel1 + Faint(indentBlock).String())
	text, _ = termtext.Wrap(text, lineWidth, padding)

	return text
}

func renderTable(text string) string {
	screenWidth, _ := terminal.Width()
	text = strings.ReplaceAll(text, markdown.ItalicStart, "")
	text = strings.ReplaceAll(text, markdown.ItalicStop, "")

	text = strings.ReplaceAll(text, markdown.BoldStart, "")
	text = strings.ReplaceAll(text, markdown.BoldStop, "")

	text = unescapeCharacters(text)
	text = removeImageReference(text)

	r, _ := glamour.NewTermRenderer(glamour.WithStyles(glamour.NoTTYStyleConfig),
		glamour.WithWordWrap(int(screenWidth)))

	out, _ := r.Render(text)

	out = strings.ReplaceAll(out, " --- ", "     ")
	out = strings.TrimPrefix(out, "\n")
	out = strings.TrimLeft(out, " ")
	out = strings.TrimPrefix(out, "\n")
	out = strings.TrimSuffix(out, "\n\n")

	return out
}

func removeImageReference(text string) string {
	exp := regexp.MustCompile(`!\[(.*?)\]\(.*?\)`)

	return exp.ReplaceAllString(text, `$1`)
}

func it(text string) string {
	italic := "\u001B[3m"
	noItalic := "\u001B[23m"

	text = strings.ReplaceAll(text, markdown.ItalicStart, italic)
	text = strings.ReplaceAll(text, markdown.ItalicStop, noItalic)

	return text
}

func itReversed(text string) string {
	italic := "\u001B[3m"
	noItalic := "\u001B[23m"

	text = strings.ReplaceAll(text, markdown.ItalicStart, noItalic)
	text = strings.ReplaceAll(text, markdown.ItalicStop, italic)

	return text
}

func bld(text string) string {
	// bold := "\033[31m"
	// noBold := "\033[0m"

	text = strings.ReplaceAll(text, markdown.BoldStart, "")
	text = strings.ReplaceAll(text, markdown.BoldStop, "")

	return text
}

func bldInQuote(text string) string {
	// bold := "\033[31m"
	// noBold := "\033[0m"

	text = strings.ReplaceAll(text, markdown.BoldStart, "")
	text = strings.ReplaceAll(text, markdown.BoldStop, "")

	return text
}

func h1(text string, lineWidth int) string {
	text = preFormatHeader(text)
	text = White(unicode.Block+" ").String() + Bold(text).String()

	text, _ = termtext.Wrap(text, lineWidth)

	return unicode.InvisibleCharacterForTopLevelComments + text
}

func h2(text string, lineWidth int) string {
	text = preFormatHeader(text)
	text = Blue(unicode.Block+" ").String() + Bold(text).String()

	text, _ = termtext.Wrap(text, lineWidth)

	return unicode.InvisibleCharacterForTopLevelComments + text
}

func h3(text string, lineWidth int) string {
	text = preFormatHeader(text)
	block := strings.Repeat(unicode.Block, 2)
	text = Red(block).String() + " " + Bold(text).String()

	text, _ = termtext.Wrap(text, lineWidth)

	return unicode.InvisibleCharacterForTopLevelComments + text
}

func h4(text string, lineWidth int) string {
	text = preFormatHeader(text)
	block := strings.Repeat(unicode.Block, 3)
	text = Magenta(block).String() + " " + Bold(text).String()

	text, _ = termtext.Wrap(text, lineWidth)

	return unicode.InvisibleCharacterForTopLevelComments + text
}

func h5(text string, lineWidth int) string {
	text = preFormatHeader(text)
	block := strings.Repeat(unicode.Block, 4)
	text = Yellow(block).String() + " " + Bold(text).String()

	text, _ = termtext.Wrap(text, lineWidth)

	return unicode.InvisibleCharacterForTopLevelComments + text
}

func h6(text string, lineWidth int) string {
	text = preFormatHeader(text)
	block := strings.Repeat(unicode.Block, 5)
	text = Green(block).String() + " " + Bold(text).String()

	text, _ = termtext.Wrap(text, lineWidth)

	return unicode.InvisibleCharacterForTopLevelComments + text
}

func removeHrefs(text string) string {
	exp := regexp.MustCompile(`<a href=.+>(.+)</a>`)
	text = exp.ReplaceAllString(text, `$1`)

	return text
}

func insertSpaceAfterItemListSeparator(text string) string {
	exp := regexp.MustCompile(`(^\s*-)(\S)`)

	return exp.ReplaceAllString(text, `$1 $2`)
}

func preFormatHeader(text string) string {
	text = removeImageReference(text)
	text = strings.TrimLeft(text, "# ")
	text = removeBoldAndItalicTags(text)
	text = unescapeCharacters(text)
	text = it(text)

	return text
}

func unescapeCharacters(text string) string {
	text = strings.ReplaceAll(text, `\|`, "|")
	text = strings.ReplaceAll(text, `\-`, "-")
	text = strings.ReplaceAll(text, `\_`, "_")
	text = strings.ReplaceAll(text, `\*`, "*")
	text = strings.ReplaceAll(text, `\\`, `\`)
	text = strings.ReplaceAll(text, `\#`, "#")
	text = strings.ReplaceAll(text, `\.`, ".")
	text = strings.ReplaceAll(text, `\>`, ">")
	text = strings.ReplaceAll(text, `\<`, "<")
	text = strings.ReplaceAll(text, "\\`", "`")
	text = strings.ReplaceAll(text, "...", "…")
	text = strings.ReplaceAll(text, `\(`, "(")
	text = strings.ReplaceAll(text, `\)`, ")")
	text = strings.ReplaceAll(text, `\[`, "[")
	text = strings.ReplaceAll(text, `\]`, "]")

	return text
}

func removeDoubleWhitespace(text string) string {
	text = strings.ReplaceAll(text, "  ", " ")

	return text
}

func removeBoldAndItalicTags(text string) string {
	text = strings.ReplaceAll(text, markdown.BoldStart, "")
	text = strings.ReplaceAll(text, markdown.BoldStop, "")

	text = strings.ReplaceAll(text, markdown.ItalicStart, "")
	text = strings.ReplaceAll(text, markdown.ItalicStop, "")

	return text
}

func trimLeadingZero(text string) string {
	text = strings.ReplaceAll(text, indentLevel2+"01", indentLevel2+" 1")
	text = strings.ReplaceAll(text, indentLevel2+"02", indentLevel2+" 2")
	text = strings.ReplaceAll(text, indentLevel2+"03", indentLevel2+" 3")
	text = strings.ReplaceAll(text, indentLevel2+"04", indentLevel2+" 4")
	text = strings.ReplaceAll(text, indentLevel2+"05", indentLevel2+" 5")
	text = strings.ReplaceAll(text, indentLevel2+"06", indentLevel2+" 6")
	text = strings.ReplaceAll(text, indentLevel2+"07", indentLevel2+" 7")
	text = strings.ReplaceAll(text, indentLevel2+"08", indentLevel2+" 8")
	text = strings.ReplaceAll(text, indentLevel2+"09", indentLevel2+" 9")

	return text
}

func highlightBackticks(text string) string {
	magenta := "\u001B[35m"
	italic := "\u001B[3m"
	normal := "\u001B[0m"

	backtick := "`"
	numberOfBackticks := strings.Count(text, backtick)
	numberOfBackticksIsOdd := numberOfBackticks%2 != 0

	if numberOfBackticks == 0 || numberOfBackticksIsOdd {
		return text
	}

	isOnFirstBacktick := true

	for i := 0; i < numberOfBackticks+1; i++ {
		if isOnFirstBacktick {
			text = strings.Replace(text, backtick, codeStart, 1)
		} else {
			text = strings.Replace(text, backtick, codeEnd, 1)
		}

		isOnFirstBacktick = !isOnFirstBacktick
	}

	exp := regexp.MustCompile(`([\S])(\[CLX_CODE_START\])`)
	text = exp.ReplaceAllString(text, `$1 $2`)

	text = strings.ReplaceAll(text, "( "+codeStart, "("+codeStart)

	text = strings.ReplaceAll(text, codeStart, normal+magenta+italic)
	text = strings.ReplaceAll(text, codeEnd, normal)

	return text
}

func replaceListPrefixes(text string) string {
	lines := strings.Split(text, "\n")
	output := ""

	for _, line := range lines {
		line = regexp.MustCompile(`^`+strings.Repeat(indentLevel1, 2)+"-").
			ReplaceAllString(line, strings.Repeat(indentLevel1, 2)+"•")

		line = regexp.MustCompile(`^`+strings.Repeat(indentLevel1, 3)+"-").
			ReplaceAllString(line, strings.Repeat(indentLevel1, 3)+"◦")

		line = regexp.MustCompile(`^`+strings.Repeat(indentLevel1, 4)+"-").
			ReplaceAllString(line, strings.Repeat(indentLevel1, 4)+"▪")

		line = regexp.MustCompile(`^`+strings.Repeat(indentLevel1, 5)+"-").
			ReplaceAllString(line, strings.Repeat(indentLevel1, 5)+"▫")

		output += line + "\n"
	}

	return output
}
