package reader

import (
	"clx/constants"
	"clx/meta"
	"clx/syntax"
	"os"
	"regexp"
	"strings"

	"github.com/muesli/reflow/wordwrap"
	"golang.org/x/term"

	"github.com/charmbracelet/glamour"

	termtext "github.com/MichaelMure/go-term-text"

	. "github.com/logrusorgru/aurora/v3"
)

const (
	indentLevel1 = "  "
	indentLevel2 = indentLevel1 + indentLevel1
	indentLevel3 = indentLevel2 + indentLevel1

	codeStart = "[CLX_CODE_START]"
	codeEnd   = "[CLX_CODE_END]"

	ansiItalic = "\u001B[3m"
)

var (
	reListToken           = regexp.MustCompile(`^\s*(-|\d+\.)`)
	reImageRefEOL         = regexp.MustCompile(`!\[(.*?)\]\(.*?\)$`)
	reImageRefSpace       = regexp.MustCompile(`!\[(.*?)\]\(.*?\)\s`)
	reImageRef            = regexp.MustCompile(`!\[(.*?)\]\(.*?\)`)
	reHref                = regexp.MustCompile(`<a href=.+>(.+)</a>`)
	reListItemNoSpace     = regexp.MustCompile(`(^\s*-)(\S)`)
	reCodeStartAfterNonWS = regexp.MustCompile(`([\S])(\[CLX_CODE_START\])`)
	reListPrefix2         = regexp.MustCompile(`^` + strings.Repeat(indentLevel1, 2) + "-")
	reListPrefix3         = regexp.MustCompile(`^` + strings.Repeat(indentLevel1, 3) + "-")
	reListPrefix4         = regexp.MustCompile(`^` + strings.Repeat(indentLevel1, 4) + "-")
	reListPrefix5         = regexp.MustCompile(`^` + strings.Repeat(indentLevel1, 5) + "-")
)

func createHeader(title string, domain string, lineWidth int) string {
	return meta.GetReaderModeMetaBlock(title, domain, lineWidth)
}

func convertToTerminalFormat(blocks []*block, lineWidth int, indentBlock string) string {
	var sb strings.Builder

	for _, b := range blocks {
		switch b.Kind {
		case blockText:
			sb.WriteString(renderText(b.Text, lineWidth))

		case blockImage:
			sb.WriteString(renderImage(b.Text, lineWidth))

		case blockCode:
			sb.WriteString(renderCode(b.Text))

		case blockQuote:
			sb.WriteString(renderQuote(b.Text, lineWidth, indentBlock))

		case blockTable:
			sb.WriteString(renderTable(b.Text))

		case blockList:
			sb.WriteString(renderList(b.Text, lineWidth))

		case blockDivider:
			sb.WriteString(renderDivider(lineWidth))

		case blockH1:
			sb.WriteString(h1(b.Text, lineWidth))

		case blockH2:
			sb.WriteString(h2(b.Text, lineWidth))

		case blockH3:
			sb.WriteString(h3(b.Text, lineWidth))

		case blockH4:
			sb.WriteString(h4(b.Text, lineWidth))

		case blockH5:
			sb.WriteString(h5(b.Text, lineWidth))

		case blockH6:
			sb.WriteString(h6(b.Text, lineWidth))

		default:
			sb.WriteString(renderText(b.Text, lineWidth))
		}

		sb.WriteString("\n\n")
	}

	return strings.TrimLeft(sb.String(), "\n")
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

	var sb strings.Builder
	lines := strings.SplitSeq(text, "\n")

	for line := range lines {
		listToken := reListToken.FindString(line)
		listText := strings.TrimPrefix(line, listToken)

		paddingBuffer := strings.Repeat(" ", len(listToken))
		padding := indentLevel1 + paddingBuffer + " "

		wrappedIndentedItem, _ := termtext.WrapWithPadIndent(listToken+listText, lineWidth, indentLevel1, padding)
		wrappedIndentedItem = insertSpaceAfterItemListSeparator(wrappedIndentedItem)

		sb.WriteString(wrappedIndentedItem)
		sb.WriteByte('\n')
	}

	output := replaceListPrefixes(sb.String())
	output = trimLeadingZero(output)

	return strings.TrimRight(output, "\n")
}

func renderImage(text string, lineWidth int) string {
	red := "\u001B[31m"
	italic := ansiItalic
	faint := "\u001B[2m"
	normal := "\u001B[0m"
	imageLabel := normal + Red(constants.Circle).Faint().String() + Yellow(constants.Circle).Faint().String() +
		Blue(constants.Circle).Faint().String() + normal + red + faint + italic + " Image " + normal + faint + italic

	text = reImageRefEOL.ReplaceAllString(text, imageLabel+`$1`)
	text = reImageRefSpace.ReplaceAllString(text, imageLabel+`$1`)
	text = reImageRef.ReplaceAllString(text, imageLabel+`$1`)

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
	screenWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))

	text = strings.TrimSuffix(text, "\n")
	text = strings.TrimPrefix(text, "\n")

	text = Faint(text).String()
	text = removeHrefs(text)

	padding := termtext.WrapPad(indentLevel1)
	text, _ = termtext.Wrap(text, screenWidth, padding)

	return text
}

func renderQuote(text string, lineWidth int, indentSymbol string) string {
	text = Italic(text).Faint().String()
	text = unescapeCharacters(text)
	text = removeHrefs(text)

	indentBlock := " " + indentSymbol
	text = itReversed(text)
	text = bld(text)

	padding := termtext.WrapPad(indentLevel1 + Faint(indentBlock).String())
	text, _ = termtext.Wrap(text, lineWidth, padding)

	return text
}

func renderTable(text string) string {
	screenWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))
	text = strings.ReplaceAll(text, italicStart, "")
	text = strings.ReplaceAll(text, italicStop, "")

	text = strings.ReplaceAll(text, boldStart, "")
	text = strings.ReplaceAll(text, boldStop, "")

	text = unescapeCharacters(text)
	text = removeImageReference(text)

	r, err := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(screenWidth))
	if err != nil {
		return text
	}

	out, _ := r.Render(text)

	out = strings.ReplaceAll(out, " --- ", "     ")
	out = strings.TrimPrefix(out, "\n")
	out = strings.TrimLeft(out, " ")
	out = strings.TrimPrefix(out, "\n")
	out = strings.TrimSuffix(out, "\n\n")

	return out
}

func removeImageReference(text string) string {
	return reImageRef.ReplaceAllString(text, `$1`)
}

func it(text string) string {
	italic := ansiItalic
	noItalic := "\u001B[23m"

	text = strings.ReplaceAll(text, italicStart, italic)
	text = strings.ReplaceAll(text, italicStop, noItalic)

	return text
}

func itReversed(text string) string {
	italic := ansiItalic
	noItalic := "\u001B[23m"

	text = strings.ReplaceAll(text, italicStart, noItalic)
	text = strings.ReplaceAll(text, italicStop, italic)

	return text
}

func bld(text string) string {
	text = strings.ReplaceAll(text, boldStart, "")
	text = strings.ReplaceAll(text, boldStop, "")

	return text
}

func h1(text string, lineWidth int) string {
	text = preFormatHeader(text)
	text = White(constants.Block+" ").String() + Bold(text).String()

	text, _ = termtext.Wrap(text, lineWidth)

	return constants.InvisibleCharacterForTopLevelComments + text
}

func h2(text string, lineWidth int) string {
	text = preFormatHeader(text)
	text = Blue(constants.Block+" ").String() + Bold(text).String()

	text, _ = termtext.Wrap(text, lineWidth)

	return constants.InvisibleCharacterForTopLevelComments + text
}

func h3(text string, lineWidth int) string {
	text = preFormatHeader(text)
	text = Red(constants.Block+" ").String() + Bold(text).String()

	text, _ = termtext.Wrap(text, lineWidth)

	return constants.InvisibleCharacterForTopLevelComments + text
}

func h4(text string, lineWidth int) string {
	text = preFormatHeader(text)
	text = Magenta(constants.Block+" ").String() + Bold(text).String()

	text, _ = termtext.Wrap(text, lineWidth)

	return constants.InvisibleCharacterForTopLevelComments + text
}

func h5(text string, lineWidth int) string {
	text = preFormatHeader(text)
	text = Yellow(constants.Block+" ").String() + Bold(text).String()

	text, _ = termtext.Wrap(text, lineWidth)

	return constants.InvisibleCharacterForTopLevelComments + text
}

func h6(text string, lineWidth int) string {
	text = preFormatHeader(text)
	text = Green(constants.Block+" ").String() + Bold(text).String()

	text, _ = termtext.Wrap(text, lineWidth)

	return constants.InvisibleCharacterForTopLevelComments + text
}

func removeHrefs(text string) string {
	return reHref.ReplaceAllString(text, `$1`)
}

func insertSpaceAfterItemListSeparator(text string) string {
	return reListItemNoSpace.ReplaceAllString(text, `$1 $2`)
}

func preFormatHeader(text string) string {
	text = removeImageReference(text)
	text = strings.TrimLeft(text, "#")
	text = strings.TrimPrefix(text, " ")
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
	text = strings.ReplaceAll(text, "  ", " ")

	return text
}

func removeBoldAndItalicTags(text string) string {
	text = strings.ReplaceAll(text, boldStart, "")
	text = strings.ReplaceAll(text, boldStop, "")

	text = strings.ReplaceAll(text, italicStart, "")
	text = strings.ReplaceAll(text, italicStop, "")

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
	italic := ansiItalic
	normal := "\u001B[0m"

	backtick := "`"
	numberOfBackticks := strings.Count(text, backtick)
	numberOfBackticksIsOdd := numberOfBackticks%2 != 0

	if numberOfBackticks == 0 || numberOfBackticksIsOdd {
		return text
	}

	isOnFirstBacktick := true

	for range numberOfBackticks + 1 {
		if isOnFirstBacktick {
			text = strings.Replace(text, backtick, codeStart, 1)
		} else {
			text = strings.Replace(text, backtick, codeEnd, 1)
		}

		isOnFirstBacktick = !isOnFirstBacktick
	}

	text = reCodeStartAfterNonWS.ReplaceAllString(text, `$1 $2`)

	text = strings.ReplaceAll(text, "( "+codeStart, "("+codeStart)

	text = strings.ReplaceAll(text, codeStart, normal+magenta+italic)
	text = strings.ReplaceAll(text, codeEnd, normal)

	return text
}

func replaceListPrefixes(text string) string {
	lines := strings.Split(text, "\n")
	var output strings.Builder

	for _, line := range lines {
		line = reListPrefix2.ReplaceAllString(line, strings.Repeat(indentLevel1, 2)+"•")
		line = reListPrefix3.ReplaceAllString(line, strings.Repeat(indentLevel1, 3)+"◦")
		line = reListPrefix4.ReplaceAllString(line, strings.Repeat(indentLevel1, 4)+"▪")
		line = reListPrefix5.ReplaceAllString(line, strings.Repeat(indentLevel1, 5)+"▫")

		output.WriteString(line + "\n")
	}

	return output.String()
}
