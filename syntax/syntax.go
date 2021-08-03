package syntax

import (
	"regexp"
	"strings"

	. "github.com/logrusorgru/aurora/v3"
)

const (
	newParagraph = "\n\n"
	reset        = "\033[0m"
	italic       = "\033[3m"
	magenta      = "\033[35m"
)

func HighlightYCStartups(comment string) string {
	expression := regexp.MustCompile(`\((YC [SW]\d{2})\)`)

	orange := uint8(214)
	black := uint8(232)
	highlightedStartup := Index(black, ` $1 `).BgIndex(orange).String()

	return expression.ReplaceAllString(comment, highlightedStartup)
}

func HighlightHackerNewsHeadlines(title string) string {
	askHN := "Ask HN:"
	showHN := "Show HN:"
	tellHN := "Tell HN:"
	launchHN := "Launch HN:"

	title = strings.ReplaceAll(title, askHN, Blue(askHN).String())
	title = strings.ReplaceAll(title, showHN, Red(showHN).String())
	title = strings.ReplaceAll(title, tellHN, Magenta(tellHN).String())
	title = strings.ReplaceAll(title, launchHN, Green(launchHN).String())

	return title
}

func HighlightSpecialContent(title string) string {
	title = strings.ReplaceAll(title, "[audio]", Yellow("audio").String())
	title = strings.ReplaceAll(title, "[video]", Yellow("video").String())
	title = strings.ReplaceAll(title, "[pdf]", Yellow("pdf").String())
	title = strings.ReplaceAll(title, "[PDF]", Yellow("PDF").String())
	title = strings.ReplaceAll(title, "[flagged]", Red("flagged").String())

	return title
}

func ConvertSmileys(text string) string {
	text = replaceBetweenWhitespace(text, `:)`, "😊")
	text = replaceBetweenWhitespace(text, `(:`, "😊")
	text = replaceBetweenWhitespace(text, `:-)`, "😊")
	text = replaceBetweenWhitespace(text, `:D`, "😄")
	text = replaceBetweenWhitespace(text, `=)`, "😃")
	text = replaceBetweenWhitespace(text, `=D`, "😃")
	text = replaceBetweenWhitespace(text, `;)`, "😉")
	text = replaceBetweenWhitespace(text, `;-)`, "😉")
	text = replaceBetweenWhitespace(text, `:P`, "😜")
	text = replaceBetweenWhitespace(text, `;P`, "😜")
	text = replaceBetweenWhitespace(text, `:o`, "😮")
	text = replaceBetweenWhitespace(text, `:O`, "😮")
	text = replaceBetweenWhitespace(text, `:(`, "😔")
	text = replaceBetweenWhitespace(text, `:-(`, "😔")
	text = replaceBetweenWhitespace(text, `:/`, "😕")
	text = replaceBetweenWhitespace(text, `:-/`, "😕")
	text = replaceBetweenWhitespace(text, `-_-`, "😑")
	text = replaceBetweenWhitespace(text, `:|`, "😐")

	return text
}

func replaceBetweenWhitespace(text string, target string, replacement string) string {
	if text == target {
		return replacement
	}

	return strings.ReplaceAll(text, " "+target, " "+replacement)
}

func RemoveUnwantedNewLines(text string) string {
	exp := regexp.MustCompile(`([\w\W[:cntrl:]])(\n)([a-zA-Z" <[:cntrl:]…])`)

	return exp.ReplaceAllString(text, `$1`+" "+`$3`)
}

func RemoveUnwantedWhitespace(text string) string {
	singleSpace := " "
	doubleSpace := "  "
	tripleSpace := "   "

	text = strings.ReplaceAll(text, tripleSpace, singleSpace)
	text = strings.ReplaceAll(text, doubleSpace, singleSpace)

	return text
}

func HighlightReferences(input string) string {
	input = strings.ReplaceAll(input, "[0]", "["+White("0").String()+"]")
	input = strings.ReplaceAll(input, "[1]", "["+Red("1").String()+"]")
	input = strings.ReplaceAll(input, "[2]", "["+Yellow("2").String()+"]")
	input = strings.ReplaceAll(input, "[3]", "["+Green("3").String()+"]")
	input = strings.ReplaceAll(input, "[4]", "["+Blue("4").String()+"]")
	input = strings.ReplaceAll(input, "[5]", "["+Cyan("5").String()+"]")
	input = strings.ReplaceAll(input, "[6]", "["+Magenta("6").String()+"]")
	input = strings.ReplaceAll(input, "[7]", "["+BrightWhite("7").String()+"]")
	input = strings.ReplaceAll(input, "[8]", "["+BrightRed("8").String()+"]")
	input = strings.ReplaceAll(input, "[9]", "["+BrightYellow("9").String()+"]")
	input = strings.ReplaceAll(input, "[10]", "["+BrightGreen("10").String()+"]")

	return input
}

func ColorizeIndentSymbol(indentSymbol string, level int) string {
	switch level {
	case 0:
		indentSymbol = ""
	case 1:
		indentSymbol = Red(indentSymbol).String()
	case 2:
		indentSymbol = Yellow(indentSymbol).String()
	case 3:
		indentSymbol = Green(indentSymbol).String()
	case 4:
		indentSymbol = Cyan(indentSymbol).String()
	case 5:
		indentSymbol = Blue(indentSymbol).String()
	case 6:
		indentSymbol = Magenta(indentSymbol).String()
	case 7:
		indentSymbol = BrightRed(indentSymbol).String()
	case 8:
		indentSymbol = BrightYellow(indentSymbol).String()
	case 9:
		indentSymbol = BrightGreen(indentSymbol).String()
	case 10:
		indentSymbol = BrightCyan(indentSymbol).String()
	case 11:
		indentSymbol = BrightBlue(indentSymbol).String()
	case 12:
		indentSymbol = BrightMagenta(indentSymbol).String()
	}

	return reset + indentSymbol
}

func TrimURLs(comment string, highlightComment bool) string {
	expression := regexp.MustCompile(`<a href=".*?" rel="nofollow">`)

	if !highlightComment {
		return expression.ReplaceAllString(comment, "")
	}

	comment = expression.ReplaceAllString(comment, "")

	e := regexp.MustCompile(`https?://([^,"\) \n]+)`)
	comment = e.ReplaceAllString(comment, Blue(`$1`).String())

	comment = strings.ReplaceAll(comment, "."+reset+" ", reset+"."+" ")

	return comment
}

func HighlightBackticks(input string) string {
	backtick := "`"
	numberOfBackticks := strings.Count(input, backtick)
	numberOfBackticksIsOdd := numberOfBackticks%2 != 0

	if numberOfBackticks == 0 || numberOfBackticksIsOdd {
		return input
	}

	isOnFirstBacktick := true

	for i := 0; i < numberOfBackticks+1; i++ {
		if isOnFirstBacktick {
			input = strings.Replace(input, backtick, italic+magenta, 1)
		} else {
			input = strings.Replace(input, backtick, reset, 1)
		}

		isOnFirstBacktick = !isOnFirstBacktick
	}

	return input
}

func HighlightMentions(input string) string {
	exp := regexp.MustCompile(`((?:^| )\B@[\w.]+)`)
	input = exp.ReplaceAllString(input, Yellow(`$1`).String())

	input = strings.ReplaceAll(input, Yellow("@dang").String(),
		Green("@dang").String())
	input = strings.ReplaceAll(input, Yellow(" @dang").String(),
		Green(" @dang").String())

	return input
}

func HighlightVariables(input string) string {
	// Highlighting variables inside commands marked with backticks
	// messes with the formatting. If there are both backticks and variables
	// in the comment, we give priority to the backticks.
	numberOfBackticks := strings.Count(input, "`")
	if numberOfBackticks > 0 {
		return input
	}

	exp := regexp.MustCompile(`(\$+[a-zA-Z_\-]+)`)

	return exp.ReplaceAllString(input, Cyan(`$1`).String())
}

func HighlightAbbreviations(input string) string {
	iAmNotALawyer := "IANAL"
	iAmALawyer := "IAAL"

	input = strings.ReplaceAll(input, iAmNotALawyer, Red(iAmNotALawyer).String())
	input = strings.ReplaceAll(input, iAmALawyer, Green(iAmALawyer).String())

	return input
}

func ReplaceCharacters(input string) string {
	input = strings.ReplaceAll(input, "&#x27;", "'")
	input = strings.ReplaceAll(input, "&gt;", ">")
	input = strings.ReplaceAll(input, "&lt;", "<")
	input = strings.ReplaceAll(input, "&#x2F;", "/")
	input = strings.ReplaceAll(input, "&quot;", `"`)
	input = strings.ReplaceAll(input, "&amp;", "&")

	return input
}

func ReplaceHTML(input string) string {
	input = strings.Replace(input, "<p>", "", 1)

	input = strings.ReplaceAll(input, "<p>", newParagraph)
	input = strings.ReplaceAll(input, "<i>", italic)
	input = strings.ReplaceAll(input, "</i>", reset)
	input = strings.ReplaceAll(input, "</a>", "")
	input = strings.ReplaceAll(input, "<pre><code>", "")
	input = strings.ReplaceAll(input, "</code></pre>", "")

	return input
}

func ReplaceSymbols(paragraph string) string {
	paragraph = strings.ReplaceAll(paragraph, "...", "…")
	paragraph = strings.ReplaceAll(paragraph, "CO2", "CO₂")

	paragraph = replaceDoubleDashes(paragraph)
	paragraph = convertFractions(paragraph)

	return paragraph
}

func replaceDoubleDashes(paragraph string) string {
	paragraph = strings.ReplaceAll(paragraph, " -- ", " — ")

	exp := regexp.MustCompile(`([a-zA-Z])--([a-zA-Z])`)

	return exp.ReplaceAllString(paragraph, `$1`+"—"+`$2`)
}

func convertFractions(text string) string {
	text = strings.ReplaceAll(text, " 1/2", " ½")
	text = strings.ReplaceAll(text, " 1/3", " ⅓")
	text = strings.ReplaceAll(text, " 2/3", " ⅔")
	text = strings.ReplaceAll(text, " 1/4", " ¼")
	text = strings.ReplaceAll(text, " 3/4", " ¾")
	text = strings.ReplaceAll(text, " 1/5", " ⅕")
	text = strings.ReplaceAll(text, " 2/5", " ⅖")
	text = strings.ReplaceAll(text, " 3/5", " ⅗")
	text = strings.ReplaceAll(text, " 4/5", " ⅘")
	text = strings.ReplaceAll(text, " 1/6", " ⅙")
	text = strings.ReplaceAll(text, " 1/10", " ⅒ ")

	text = strings.ReplaceAll(text, "1/5th", "⅕th")
	text = strings.ReplaceAll(text, "1/6th", "⅙th")
	text = strings.ReplaceAll(text, "1/10th", "⅒ th")

	return text
}
