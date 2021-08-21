package renderer

import (
	"clx/constants/messages"
	"clx/indent"
	"clx/markdown"
	"clx/syntax"
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour"

	terminal "github.com/wayneashleyberry/terminal-dimensions"

	termtext "github.com/MichaelMure/go-term-text"
	. "github.com/logrusorgru/aurora/v3"
)

const (
	indentLevel1 = "  "
	indentLevel2 = "    "
	newLine      = "\n"
	newParagraph = "\n\n"
)

func CreateHeader(title string, domain string, lineWidth int) string {
	wrappedTitle, _ := termtext.Wrap(title, lineWidth)
	truncatedDomain := termtext.TruncateMax(domain, lineWidth)

	wrappedTitle += newParagraph
	wrappedTitle += Faint(messages.LessScreenInfo).String() + newLine
	wrappedTitle += Faint(truncatedDomain).String() + newLine
	separator := messages.GetSeparator(lineWidth)
	wrappedTitle += separator + newParagraph

	return wrappedTitle
}

func ToString(blocks []*markdown.Block, lineWidth int, altIndentBlock bool) string {
	output := ""

	for _, block := range blocks {
		switch block.Kind {
		case markdown.Text:
			output += renderText(block.Text, lineWidth, indentLevel1) + "\n\n"

		case markdown.Image:
			output += renderImage(block.Text, lineWidth) + "\n\n"

		case markdown.Code:
			output += renderCode(block.Text) + "\n\n"

		case markdown.Quote:
			output += renderQuote(block.Text, lineWidth, altIndentBlock) + "\n\n"

		case markdown.Table:
			output += renderTable(block.Text, indentLevel2) + "\n\n"

		case markdown.List:
			output += renderList(block.Text, lineWidth, indentLevel2) + "\n\n"

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
			output += renderText(block.Text, lineWidth, indentLevel1) + "\n\n"
		}
	}

	return output
}

func renderDivider(lineWidth int) string {
	divider := strings.Repeat("-", lineWidth-len(indentLevel2)*2)

	return Faint(indentLevel2 + divider).String()
}

func renderText(text string, lineWidth int, indentLevel string) string {
	text = it(text)
	text = bld(text)
	text = removeHrefs(text)
	text = unescapeCharacters(text)

	text = syntax.RemoveUnwantedNewLines(text)
	text = syntax.HighlightBackticks(text)
	text = syntax.HighlightMentions(text)

	text = strings.ReplaceAll(text, `\_`, "_")

	padding := termtext.WrapPad(indentLevel)
	text, _ = termtext.Wrap(text, lineWidth, padding)

	return text
}

func renderList(text string, lineWidth int, indentLevel string) string {
	text = it(text)
	text = bld(text)
	text = removeImageReference(text)
	text = removeHrefs(text)
	text = unescapeCharacters(text)

	// Remove unwanted newlines
	exp := regexp.MustCompile(`([\w\W[:cntrl:]])(\n)([a-zA-Z])`)
	text = exp.ReplaceAllString(text, `$1`+" "+`$3`)

	text = syntax.HighlightBackticks(text)

	padding := termtext.WrapPad(indentLevel)
	text, _ = termtext.Wrap(text, lineWidth, padding)

	// Indent lists that span over two lines
	indentation := indentLevel + "  "
	expIndent := regexp.MustCompile(`\n` + indentLevel + `([a-zA-Z])`)
	text = expIndent.ReplaceAllString(text, "\n"+indentation+`$1`)

	text = strings.ReplaceAll(text, ` \- `, " - ")

	return text
}

func renderImage(text string, lineWidth int) string {
	magenta := "\u001B[35m"
	italic := "\u001B[3m"
	faint := "\u001B[2m"
	normal := "\u001B[0m"

	exp := regexp.MustCompile(`!\[(.*?)\]\(.*?\)`)
	image := magenta + faint + "Image: " + normal + faint + italic

	// imageLabel := image+italic+faint+`$1.`+"### "

	text = exp.ReplaceAllString(text, image+`$1. `)

	lines := strings.Split(text, image)
	output := ""

	for i, line := range lines {
		if i == 0 || line == "" {
			continue
		}

		output += image + line + "\n\n"
	}

	// Remove 'Image: .' for images without captions
	output = strings.ReplaceAll(output, image+". ", magenta+faint+"Image "+normal+faint+italic)

	// output = strings.ReplaceAll(output, "###", "")
	output = strings.ReplaceAll(output, "..", ".")
	output = strings.TrimSuffix(output, "\n\n")
	output += normal

	output = it(output)
	output = bld(output)

	padding := termtext.WrapPad(indentLevel2)
	output, _ = termtext.Wrap(output, lineWidth, padding)

	return output
}

func renderCode(text string) string {
	screenWidth, _ := terminal.Width()

	text = strings.TrimSuffix(text, "\n")
	text = strings.TrimPrefix(text, "\n")

	text = Faint(text).String()
	text = removeHrefs(text)

	padding := termtext.WrapPad(indentLevel2)
	text, _ = termtext.Wrap(text, int(screenWidth), padding)

	return text
}

func renderQuote(text string, lineWidth int, altIndentBlock bool) string {
	text = Italic(text).Faint().String()
	text = unescapeCharacters(text)

	text = removeUnwantedNewLines(text)

	indentSymbol := " " + indent.GetIndentSymbol(false, altIndentBlock)
	text = itReversed(text)
	text = bldInQuote(text)

	padding := termtext.WrapPad(indentLevel2 + Faint(indentSymbol).String())
	text, _ = termtext.Wrap(text, lineWidth, padding)

	// text = strings.TrimSuffix(text, "\n")
	// text = strings.TrimPrefix(text, "\n")

	return text
}

func removeUnwantedNewLines(text string) string {
	paragraphSeparator := "\n\n"
	paragraphs := strings.Split(text, paragraphSeparator)
	output := ""

	for _, paragraph := range paragraphs {
		paragraph = syntax.RemoveUnwantedNewLines(paragraph)

		output += paragraph + paragraphSeparator
	}

	output = strings.TrimSuffix(output, paragraphSeparator)

	return output
}

func renderTable(text string, indentLevel string) string {
	text = strings.ReplaceAll(text, markdown.ItalicStart, "")
	text = strings.ReplaceAll(text, markdown.ItalicStop, "")

	text = strings.ReplaceAll(text, markdown.BoldStart, "")
	text = strings.ReplaceAll(text, markdown.BoldStop, "")

	text = strings.ReplaceAll(text, `\*`, "*")

	r, _ := glamour.NewTermRenderer(glamour.WithStyles(glamour.NoTTYStyleConfig), glamour.WithWordWrap(80))

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
	text = Bold(text).String()

	padding := termtext.WrapPad("")
	text, _ = termtext.Wrap(text, lineWidth, padding)

	return text
}

func h2(text string, lineWidth int) string {
	text = preFormatHeader(text)
	text = Bold(text).String()

	padding := termtext.WrapPad("")
	text, _ = termtext.Wrap(text, lineWidth, padding)

	return text
}

func h3(text string, lineWidth int) string {
	text = preFormatHeader(text)
	text = Bold(text).Underline().Blue().String()

	padding := termtext.WrapPad(indentLevel1)
	text, _ = termtext.Wrap(text, lineWidth, padding)

	return text
}

func h4(text string, lineWidth int) string {
	text = preFormatHeader(text)
	text = Bold(text).Underline().Yellow().String()

	padding := termtext.WrapPad(indentLevel1)
	text, _ = termtext.Wrap(text, lineWidth, padding)

	return text
}

func h5(text string, lineWidth int) string {
	text = preFormatHeader(text)
	text = Bold(text).Underline().Red().String()

	padding := termtext.WrapPad(indentLevel1)
	text, _ = termtext.Wrap(text, lineWidth, padding)

	return text
}

func h6(text string, lineWidth int) string {
	text = preFormatHeader(text)
	text = Bold(text).Underline().Cyan().String()

	padding := termtext.WrapPad(indentLevel1)
	text, _ = termtext.Wrap(text, lineWidth, padding)

	return text
}

func removeHrefs(text string) string {
	exp := regexp.MustCompile(`<a href=.+>(.+)</a>`)
	text = exp.ReplaceAllString(text, `$1`)

	return text
}

func preFormatHeader(text string) string {
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

	return text
}

func removeBoldAndItalicTags(text string) string {
	text = strings.ReplaceAll(text, markdown.BoldStart, "")
	text = strings.ReplaceAll(text, markdown.BoldStop, "")

	text = strings.ReplaceAll(text, markdown.ItalicStart, "")
	text = strings.ReplaceAll(text, markdown.ItalicStop, "")

	return text
}
