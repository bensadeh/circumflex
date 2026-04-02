package article

import (
	"regexp"
	"strings"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/style"
	"github.com/bensadeh/circumflex/syntax"

	"github.com/muesli/reflow/wordwrap"

	"charm.land/glamour/v2"
	"charm.land/glamour/v2/styles"

	termtext "github.com/MichaelMure/go-term-text"

	"charm.land/lipgloss/v2"
)

const (
	sectionMarker = "■"
	circle        = "●"
)

const (
	indentLevel1 = "  "
	indentLevel2 = indentLevel1 + indentLevel1
	indentLevel3 = indentLevel2 + indentLevel1

	codeStart = "[CLX_CODE_START]"
	codeEnd   = "[CLX_CODE_END]"
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

func createHeader(url string, lineWidth int) string {
	s := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		PaddingLeft(1).
		PaddingRight(1).
		Width(lineWidth)

	contentWidth := lineWidth - s.GetHorizontalBorderSize() - s.GetHorizontalPadding()

	formattedURL := style.MetaURL(termtext.TruncateMax(url, contentWidth))
	info := "\n\n" + style.MetaReaderMode("Reader Mode")

	return s.Render(formattedURL+info) + "\n\n"
}

func convertToTerminalFormat(blocks []*block, lineWidth int) string {
	var sb strings.Builder

	for _, b := range blocks {
		switch b.Kind {
		case blockText:
			sb.WriteString(renderText(b.Text, lineWidth))

		case blockImage:
			sb.WriteString(renderImage(b.Text, lineWidth))

		case blockCode:
			sb.WriteString(renderCode(b.Text, lineWidth))

		case blockQuote:
			sb.WriteString(renderQuote(b.Text, lineWidth))

		case blockTable:
			sb.WriteString(renderTable(b.Text, lineWidth))

		case blockList:
			sb.WriteString(renderList(b.Text, lineWidth))

		case blockDivider:
			sb.WriteString(renderDivider(lineWidth))

		case blockH1, blockH2, blockH3, blockH4, blockH5, blockH6:
			sb.WriteString(renderHeader(b.Kind, b.Text, lineWidth))

		default:
			sb.WriteString(renderText(b.Text, lineWidth))
		}

		sb.WriteString("\n\n")
	}

	return strings.Trim(sb.String(), "\n")
}

func renderDivider(lineWidth int) string {
	divider := strings.Repeat("-", lineWidth-len(indentLevel1)*2)

	return style.Faint(indentLevel1 + divider)
}

func renderText(text string, lineWidth int) string {
	text = it(text)
	text = removeHrefs(text)
	text = unescapeCharacters(text)
	text = removeImageReference(text)

	text = syntax.RemoveUnwantedNewLines(text)
	text = highlightBackticks(text)
	text = syntax.HighlightMentions(text)
	text = syntax.TrimURLs(text, false)

	return wordwrap.String(text, lineWidth)
}

func renderList(text string, lineWidth int) string {
	text = it(text)
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
	normal := ansi.Reset
	imageColor := style.ReaderImageColor()
	imageLabel := normal +
		lipgloss.NewStyle().Foreground(style.HeaderC()).Faint(true).Render(circle) +
		lipgloss.NewStyle().Foreground(style.HeaderL()).Faint(true).Render(circle) +
		lipgloss.NewStyle().Foreground(style.HeaderX()).Faint(true).Render(circle) +
		normal + lipgloss.NewStyle().Foreground(imageColor).Faint(true).Italic(true).Render(" Image ") + ansi.Faint + ansi.Italic

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
	output = removeDoubleWhitespace(output)

	padding := termtext.WrapPad(indentLevel1)
	output, _ = termtext.Wrap(output, lineWidth, padding)

	return output
}

func renderCode(text string, lineWidth int) string {
	text = strings.TrimSuffix(text, "\n")
	text = strings.TrimPrefix(text, "\n")

	text = style.Faint(text)
	text = removeHrefs(text)

	padding := termtext.WrapPad(indentLevel1)
	text, _ = termtext.Wrap(text, lineWidth, padding)

	return text
}

func renderQuote(text string, lineWidth int) string {
	text = lipgloss.NewStyle().Italic(true).Faint(true).Render(text)
	text = unescapeCharacters(text)
	text = removeHrefs(text)

	indentBlock := " " + style.IndentSymbol
	text = itReversed(text)

	padding := termtext.WrapPad(indentLevel1 + style.Faint(indentBlock))
	text, _ = termtext.Wrap(text, lineWidth, padding)

	return text
}

func renderTable(text string, lineWidth int) string {
	text = strings.ReplaceAll(text, italicStart, "")
	text = strings.ReplaceAll(text, italicStop, "")

	text = unescapeCharacters(text)
	text = removeImageReference(text)

	r, err := glamour.NewTermRenderer(glamour.WithStyles(styles.ASCIIStyleConfig), glamour.WithWordWrap(lineWidth))
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
	text = strings.ReplaceAll(text, italicStart, ansi.Italic)
	text = strings.ReplaceAll(text, italicStop, ansi.ItalicOff)

	return text
}

func itReversed(text string) string {
	text = strings.ReplaceAll(text, italicStart, ansi.ItalicOff)
	text = strings.ReplaceAll(text, italicStop, ansi.Italic)

	return text
}

var headerIndent = [...]int{0, 0, 0, 2, 4, 6, 8, 10}

var headerStyleFuncs = [...]func(string) string{
	0: nil, 1: nil,
	blockH1: style.ReaderH1,
	blockH2: style.ReaderH2,
	blockH3: style.ReaderH3,
	blockH4: style.ReaderH4,
	blockH5: style.ReaderH5,
	blockH6: style.ReaderH6,
}

func renderHeader(kind blockKind, text string, lineWidth int) string {
	text = preFormatHeader(text)
	indent := headerIndent[kind]
	styleFn := headerStyleFuncs[kind]
	text = styleFn(sectionMarker+" ") + style.Bold(text)

	text, _ = termtext.Wrap(text, lineWidth-indent)

	if indent > 0 {
		padding := strings.Repeat(" ", indent)
		text = strings.ReplaceAll(text, "\n", "\n"+padding)
		text = padding + text
	}

	return text
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
	text = removeItalicTags(text)
	text = unescapeCharacters(text)
	text = it(text)

	return text
}

var unescaper = strings.NewReplacer(
	`\|`, "|",
	`\-`, "-",
	`\_`, "_",
	`\*`, "*",
	`\\`, `\`,
	`\#`, "#",
	`\.`, ".",
	`\>`, ">",
	`\<`, "<",
	"\\`", "`",
	"...", "…",
	`\(`, "(",
	`\)`, ")",
	`\[`, "[",
	`\]`, "]",
)

func unescapeCharacters(text string) string {
	return unescaper.Replace(text)
}

func removeDoubleWhitespace(text string) string {
	text = strings.ReplaceAll(text, "  ", " ")

	return text
}

func removeItalicTags(text string) string {
	text = strings.ReplaceAll(text, italicStart, "")
	text = strings.ReplaceAll(text, italicStop, "")

	return text
}

var leadingZeroTrimmer = strings.NewReplacer(
	indentLevel2+"01", indentLevel2+" 1",
	indentLevel2+"02", indentLevel2+" 2",
	indentLevel2+"03", indentLevel2+" 3",
	indentLevel2+"04", indentLevel2+" 4",
	indentLevel2+"05", indentLevel2+" 5",
	indentLevel2+"06", indentLevel2+" 6",
	indentLevel2+"07", indentLevel2+" 7",
	indentLevel2+"08", indentLevel2+" 8",
	indentLevel2+"09", indentLevel2+" 9",
)

func trimLeadingZero(text string) string {
	return leadingZeroTrimmer.Replace(text)
}

func highlightBackticks(text string) string {
	numberOfBackticks := strings.Count(text, "`")
	if numberOfBackticks == 0 || numberOfBackticks%2 != 0 {
		return text
	}

	isOnFirstBacktick := true

	for range numberOfBackticks + 1 {
		if isOnFirstBacktick {
			text = strings.Replace(text, "`", codeStart, 1)
		} else {
			text = strings.Replace(text, "`", codeEnd, 1)
		}

		isOnFirstBacktick = !isOnFirstBacktick
	}

	text = reCodeStartAfterNonWS.ReplaceAllString(text, `$1 $2`)
	text = strings.ReplaceAll(text, "( "+codeStart, "("+codeStart)

	for {
		start := strings.Index(text, codeStart)
		if start == -1 {
			break
		}

		end := strings.Index(text[start:], codeEnd)
		if end == -1 {
			break
		}

		end += start
		content := text[start+len(codeStart) : end]
		text = text[:start] + ansi.Reset + style.CommentBacktick(content) + text[end+len(codeEnd):]
	}

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
