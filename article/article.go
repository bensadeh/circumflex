package article

import (
	"clx/constants/messages"
	"strconv"
	"strings"

	text "github.com/MichaelMure/go-term-text"

	. "github.com/logrusorgru/aurora/v3"
)

const (
	newLine     = "\n"
	indentBlock = "   "
	link1       = "\033]8;;"
	link2       = "\a"
	link3       = "\033]8;;\a"
	normal      = "\033[0m"
	screenWidth = 73
)

func Parse(title, domain, article, references string) string {
	wrappedTitle, _ := text.Wrap(title, screenWidth)
	truncatedDomain := text.TruncateMax(domain, screenWidth)

	wrappedTitle += newLine
	wrappedTitle += Faint(truncatedDomain).String() + newLine
	wrappedTitle += Faint(messages.LessScreenInfo).String() + newLine
	separator := messages.GetSeparator(screenWidth)
	wrappedTitle += separator + newLine + newLine

	if strings.Contains(domain, "https://en.wikipedia.org") {
		article = preFormatWikipediaArticle(article)
	}

	lines := strings.Split(article, newLine)
	formattedArticle := ""

	for i, line := range lines {
		if line == "References" || line == "     *" {
			break
		}

		if isQuote(line) {
			formattedArticle += normal + Faint(line).Italic().String() + newLine
			formattedArticle = highlightReferencesInsideQuotes(formattedArticle)

			continue
		}

		if isHeader(lines, i, line) {
			formattedArticle += normal + Bold(line).String() + newLine

			continue
		}

		formattedArticle += normal + line + newLine
		formattedArticle = highlightReferences(formattedArticle)
	}

	formattedReferences := formatReferences(references)

	return wrappedTitle + formattedArticle + formattedReferences
}

func isHeader(lines []string, i int, line string) bool {
	currentLineIsIndented := strings.HasPrefix(line, " ")

	if currentLineIsIndented {
		return false
	}

	var previousLine string
	var nextLine string

	isOnFirstLine := i == 0
	isOnLastLine := len(lines) == i-1

	if isOnFirstLine {
		previousLine = ""
	} else {
		previousLine = lines[i-1]
	}

	if isOnLastLine {
		nextLine = ""
	} else {
		nextLine = lines[i+1]
	}

	previousLineIsEmpty := len(previousLine) == 0
	nextLineLineIsEmpty := len(nextLine) == 0

	lineIsHeader := previousLineIsEmpty && nextLineLineIsEmpty

	return lineIsHeader
}

func isQuote(input string) bool {
	indentCutOff := 5

	for i, c := range input {
		if i < indentCutOff && c != ' ' {
			return false
		}

		if i == indentCutOff && c != ' ' && c != '*' {
			return true
		}
	}

	return false
}

func highlightReferences(input string) string {
	input = strings.ReplaceAll(input, "[1]", "["+Red("1").String()+"]")
	input = strings.ReplaceAll(input, "[2]", "["+Yellow("2").String()+"]")
	input = strings.ReplaceAll(input, "[3]", "["+Green("3").String()+"]")
	input = strings.ReplaceAll(input, "[4]", "["+Cyan("4").String()+"]")
	input = strings.ReplaceAll(input, "[5]", "["+Blue("5").String()+"]")
	input = strings.ReplaceAll(input, "[6]", "["+Magenta("6").String()+"]")
	input = strings.ReplaceAll(input, "[7]", "["+White("7").String()+"]")
	input = strings.ReplaceAll(input, "[8]", "["+BrightRed("8").String()+"]")
	input = strings.ReplaceAll(input, "[9]", "["+BrightYellow("9").String()+"]")
	input = strings.ReplaceAll(input, "[10]", "["+BrightGreen("10").String()+"]")
	input = strings.ReplaceAll(input, "[11]", "["+BrightCyan("11").String()+"]")
	input = strings.ReplaceAll(input, "[12]", "["+BrightBlue("12").String()+"]")
	input = strings.ReplaceAll(input, "[13]", "["+BrightMagenta("13").String()+"]")
	input = strings.ReplaceAll(input, "[14]", "["+White("14").String()+"]")
	input = strings.ReplaceAll(input, "[15]", "["+Red("15").String()+"]")
	input = strings.ReplaceAll(input, "[16]", "["+Yellow("16").String()+"]")

	return input
}

func highlightReferencesInsideQuotes(input string) string {
	faintItalic := "\033[2m" + "\033[3m"

	input = strings.ReplaceAll(input, "[1]", "["+Red("1").String()+faintItalic+"]")
	input = strings.ReplaceAll(input, "[2]", "["+Yellow("2").String()+faintItalic+"]")
	input = strings.ReplaceAll(input, "[3]", "["+Green("3").String()+faintItalic+"]")
	input = strings.ReplaceAll(input, "[4]", "["+Cyan("4").String()+faintItalic+"]")
	input = strings.ReplaceAll(input, "[5]", "["+Blue("5").String()+faintItalic+"]")
	input = strings.ReplaceAll(input, "[6]", "["+Magenta("6").String()+faintItalic+"]")
	input = strings.ReplaceAll(input, "[7]", "["+White("7").String()+faintItalic+"]")
	input = strings.ReplaceAll(input, "[8]", "["+BrightRed("8").String()+faintItalic+"]")
	input = strings.ReplaceAll(input, "[9]", "["+BrightYellow("9").String()+faintItalic+"]")
	input = strings.ReplaceAll(input, "[10]", "["+BrightGreen("10").String()+faintItalic+"]")
	input = strings.ReplaceAll(input, "[11]", "["+BrightCyan("11").String()+faintItalic+"]")
	input = strings.ReplaceAll(input, "[12]", "["+BrightBlue("12").String()+faintItalic+"]")
	input = strings.ReplaceAll(input, "[13]", "["+BrightMagenta("13").String()+faintItalic+"]")
	input = strings.ReplaceAll(input, "[14]", "["+White("14").String()+faintItalic+"]")
	input = strings.ReplaceAll(input, "[15]", "["+Red("15").String()+faintItalic+"]")
	input = strings.ReplaceAll(input, "[16]", "["+Yellow("16").String()+faintItalic+"]")

	return input
}

func formatReferences(references string) string {
	lines := strings.Split(references, newLine)
	formattedReferences := Faint("References").String() + newLine + newLine

	if len(lines) == 1 {
		return ""
	}

	for i, line := range lines {
		isOnLastLine := i == len(lines)-1

		if isOnLastLine {
			formattedReferences += newLine

			break
		}

		if i == 16 {
			formattedReferences += newLine

			break
		}

		number := strconv.Itoa(i + 1)

		formattedReferences += indentBlock + "[" + number + "] " + formatURL(line, 65) + newLine
	}

	formattedReferences = highlightReferences(formattedReferences)

	return formattedReferences
}

func formatURL(url string, maxURLLength int) string {
	if len(url) < maxURLLength {
		return getHyperlinkText(url, url)
	}

	truncatedURL := text.TruncateMax(url, maxURLLength)

	return getHyperlinkText(url, truncatedURL)
}

func getHyperlinkText(url string, text string) string {
	return link1 + url + link2 + text + link3
}
