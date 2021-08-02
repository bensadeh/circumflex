package syntax

import (
	"clx/colors"
	"regexp"
	"strings"

	"github.com/logrusorgru/aurora/v3"
)

const (
	askHN       = "Ask HN:"
	showHN      = "Show HN:"
	tellHN      = "Tell HN:"
	launchHN    = "Launch HN:"
	singleSpace = " "
	doubleSpace = "  "
	tripleSpace = "   "
)

func HighlightYCStartups(comment string) string {
	expression := regexp.MustCompile(`\((YC [SW]\d{2})\)`)

	firstHighlightGroup := `$1`
	highlightedStartup := colors.OrangeBackground + colors.NearBlack + " " + firstHighlightGroup + " " + colors.Normal

	return expression.ReplaceAllString(comment, highlightedStartup)
}

func HighlightHackerNewsHeadlines(title string) string {
	title = strings.ReplaceAll(title, askHN, aurora.Blue(askHN).String())
	title = strings.ReplaceAll(title, showHN, aurora.Red(showHN).String())
	title = strings.ReplaceAll(title, tellHN, aurora.Magenta(tellHN).String())
	title = strings.ReplaceAll(title, launchHN, aurora.Green(launchHN).String())

	return title
}

func HighlightSpecialContent(title string) string {
	title = strings.ReplaceAll(title, "[audio]", aurora.Yellow("audio").String())
	title = strings.ReplaceAll(title, "[video]", aurora.Yellow("video").String())
	title = strings.ReplaceAll(title, "[pdf]", aurora.Yellow("pdf").String())
	title = strings.ReplaceAll(title, "[PDF]", aurora.Yellow("PDF").String())
	title = strings.ReplaceAll(title, "[flagged]", aurora.Red("flagged").String())

	return title
}

func HighlightWhoIsHiring(title string, author string) string {
	if author != "whoishiring" {
		return title
	}

	title = strings.ReplaceAll(title, " (", colors.Normal+" (")

	if strings.Contains(title, "Who is hiring?") {
		title = aurora.Index(232, title).String()

		return aurora.BgBlue(title).String()
	}

	if strings.Contains(title, "Freelancer?") {
		title = aurora.Index(232, title).String()

		return aurora.BgRed(title).String()
	}

	if strings.Contains(title, "Who wants to be hired?") {
		title = aurora.Index(232, title).String()

		return aurora.BgYellow(title).String()
	}

	return title
}

func ConvertSmileys(text string) string {
	text = replaceWhitespaceSeparatedToken(text, `\:\)`, "😊")
	text = replaceWhitespaceSeparatedToken(text, `\(\:`, "😊")
	text = replaceWhitespaceSeparatedToken(text, `\:\-\)`, "😊")
	text = replaceWhitespaceSeparatedToken(text, `\:D`, "😄")
	text = replaceWhitespaceSeparatedToken(text, `\=\)`, "😃")
	text = replaceWhitespaceSeparatedToken(text, `\=D`, "😃")
	text = replaceWhitespaceSeparatedToken(text, `\;\)`, "😉")
	text = replaceWhitespaceSeparatedToken(text, `\;\-\)`, "😉")
	text = replaceWhitespaceSeparatedToken(text, `\:P`, "😜")
	text = replaceWhitespaceSeparatedToken(text, `\;P`, "😜")
	text = replaceWhitespaceSeparatedToken(text, `\:o`, "😮")
	text = replaceWhitespaceSeparatedToken(text, `\:O`, "😮")
	text = replaceWhitespaceSeparatedToken(text, `\:\(`, "😔")
	text = replaceWhitespaceSeparatedToken(text, `\:\-\(`, "😔")
	text = replaceWhitespaceSeparatedToken(text, `\:\/`, "😕")
	text = replaceWhitespaceSeparatedToken(text, `\:\-\/`, "😕")
	text = replaceWhitespaceSeparatedToken(text, `-_-`, "😑")
	text = replaceWhitespaceSeparatedToken(text, `:\|`, "😐")

	return text
}

func ConvertFractions(text string) string {
	text = replaceWhitespaceSeparatedToken(text, "1/2", "½")
	text = replaceWhitespaceSeparatedToken(text, "1/3", "⅓")
	text = replaceWhitespaceSeparatedToken(text, "2/3", "⅔")
	text = replaceWhitespaceSeparatedToken(text, "1/4", "¼")
	text = replaceWhitespaceSeparatedToken(text, "3/4", "¾")
	text = replaceWhitespaceSeparatedToken(text, "1/5", "⅕")
	text = replaceWhitespaceSeparatedToken(text, "2/5", "⅖")
	text = replaceWhitespaceSeparatedToken(text, "3/5", "⅗")
	text = replaceWhitespaceSeparatedToken(text, "4/5", "⅘")
	text = replaceWhitespaceSeparatedToken(text, "1/6", "⅙")
	text = replaceWhitespaceSeparatedToken(text, "1/10", "⅒ ")

	text = strings.ReplaceAll(text, "1/5th", "⅕th")
	text = strings.ReplaceAll(text, "1/6th", "⅙th")
	text = strings.ReplaceAll(text, "1/10th", "⅒ th")

	return text
}

func replaceWhitespaceSeparatedToken(text, targetToken, replacementToken string) string {
	exp := regexp.MustCompile(`((?:^| ))(` + targetToken + `)((?:$| |\.|\,)|\))`)

	return exp.ReplaceAllString(text, `$1`+replacementToken+`$3`)
}

func RemoveUnwantedNewLines(text string) string {
	exp := regexp.MustCompile(`([\w\W[:cntrl:]])(\n)([a-zA-Z" <[:cntrl:]…])`)

	return exp.ReplaceAllString(text, `$1`+" "+`$3`)
}

func RemoveUnwantedWhitespace(text string) string {
	text = strings.ReplaceAll(text, tripleSpace, singleSpace)
	text = strings.ReplaceAll(text, doubleSpace, singleSpace)

	return text
}

func HighlightReferences(input string) string {
	input = strings.ReplaceAll(input, "[0]", "["+aurora.White("0").String()+"]")
	input = strings.ReplaceAll(input, "[1]", "["+aurora.Red("1").String()+"]")
	input = strings.ReplaceAll(input, "[2]", "["+aurora.Yellow("2").String()+"]")
	input = strings.ReplaceAll(input, "[3]", "["+aurora.Green("3").String()+"]")
	input = strings.ReplaceAll(input, "[4]", "["+aurora.Blue("4").String()+"]")
	input = strings.ReplaceAll(input, "[5]", "["+aurora.Cyan("5").String()+"]")
	input = strings.ReplaceAll(input, "[6]", "["+aurora.Magenta("6").String()+"]")
	input = strings.ReplaceAll(input, "[7]", "["+aurora.BrightWhite("7").String()+"]")
	input = strings.ReplaceAll(input, "[8]", "["+aurora.BrightRed("8").String()+"]")
	input = strings.ReplaceAll(input, "[9]", "["+aurora.BrightYellow("9").String()+"]")
	input = strings.ReplaceAll(input, "[10]", "["+aurora.BrightGreen("10").String()+"]")

	return input
}

func ColorizeIndentSymbol(indentSymbol string, level int) string {
	if level == 0 {
		return ""
	}

	switch level {
	case 1:
		indentSymbol = aurora.Red(indentSymbol).String()
	case 2:
		indentSymbol = aurora.Yellow(indentSymbol).String()
	case 3:
		indentSymbol = aurora.Green(indentSymbol).String()
	case 4:
		indentSymbol = aurora.Cyan(indentSymbol).String()
	case 5:
		indentSymbol = aurora.Blue(indentSymbol).String()
	case 6:
		indentSymbol = aurora.Magenta(indentSymbol).String()
	case 7:
		indentSymbol = aurora.BrightRed(indentSymbol).String()
	case 8:
		indentSymbol = aurora.BrightYellow(indentSymbol).String()
	case 9:
		indentSymbol = aurora.BrightGreen(indentSymbol).String()
	case 10:
		indentSymbol = aurora.BrightCyan(indentSymbol).String()
	case 11:
		indentSymbol = aurora.BrightBlue(indentSymbol).String()
	case 12:
		indentSymbol = aurora.BrightMagenta(indentSymbol).String()
	}

	resetColor := "\033[0m"

	return resetColor + indentSymbol
}
