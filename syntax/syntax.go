package syntax

import (
	"clx/constants"
	"clx/nerdfonts"
	"clx/style"
	"image/color"
	"regexp"
	"strings"

	"charm.land/lipgloss/v2"
)

const (
	newParagraph = "\n\n"
	ansiBlack    = 16 // ANSI 256-color black

	Unselected = iota
	HeadlineInCommentSection
	Selected
	MarkAsRead
	AddToFavorites
	RemoveFromFavorites
)

var (
	reYCWithSeason    = regexp.MustCompile(`\((YC ([SW]\d{2}))\)`)
	reYCWithoutSeason = regexp.MustCompile(`\((YC [SW]\d{2})\)`)
	reYear            = regexp.MustCompile(`\((\d{4})\)`)
	reUnwantedNewLine = regexp.MustCompile(`([\w\W[:cntrl:]])(\n)([a-zA-Z0-9" \-<[:cntrl:]…])`)
	reHTMLAnchor      = regexp.MustCompile(`<a href=".*?".*>`)
	reURL             = regexp.MustCompile(`https?://([^,"\) \n]+)`)
	reMention         = regexp.MustCompile(`((?:^| )\B@[\w.]+)`)
	reVariable        = regexp.MustCompile(`(\$+[a-zA-Z_\-]+)`)
	reDoubleDash      = regexp.MustCompile(`([a-zA-Z])--([a-zA-Z])`)
)

func HighlightYCStartupsInHeadlines(comment string, highlightType int, enableNerdFonts bool) string {
	if enableNerdFonts {
		highlightedStartup := style.Reset + getYCBarNerdFonts(nerdfonts.YCombinator+constants.NoBreakSpace+`$2`, highlightType) +
			getHighlight(highlightType)

		return reYCWithSeason.ReplaceAllString(comment, highlightedStartup)
	}

	highlightedStartup := style.Reset + getYCBar(`$1`, highlightType) +
		getHighlight(highlightType)

	return reYCWithoutSeason.ReplaceAllString(comment, highlightedStartup)
}

func getYCBar(text string, highlightType int) string {
	c := style.HeadlineYCLabelColor()

	switch highlightType {
	case Selected:
		return lipgloss.NewStyle().Foreground(c).Reverse(true).Render(text)

	case MarkAsRead:
		return lipgloss.NewStyle().Foreground(c).Faint(true).Render(text)

	default:
		return lipgloss.NewStyle().Foreground(c).Render(text)
	}
}

func getYCBarNerdFonts(text string, highlightType int) string {
	c := style.HeadlineYCLabelColor()
	black := lipgloss.ANSIColor(ansiBlack)

	if highlightType == Selected {
		return label(text, c, black, highlightType)
	}

	return label(text, black, c, highlightType)
}

func HighlightYear(comment string, highlightType int) string {
	content := getYear(`$1`, highlightType)

	return reYear.ReplaceAllString(comment, style.Reset+content+getHighlight(highlightType))
}

func getYear(text string, highlightType int) string {
	c := style.HeadlineYearColor()

	switch highlightType {
	case Selected:
		return lipgloss.NewStyle().Foreground(c).Reverse(true).Render(text)

	case MarkAsRead:
		return lipgloss.NewStyle().Foreground(c).Faint(true).Render(text)

	default:
		return lipgloss.NewStyle().Foreground(c).Render(text)
	}
}

func label(text string, fg color.Color, bg color.Color, highlightType int) string {
	content := lipgloss.NewStyle().
		Foreground(fg).
		Background(bg)

	if highlightType == MarkAsRead {
		content.
			Italic(true).
			Faint(true)
	}

	if highlightType == HeadlineInCommentSection {
		content.Bold(true)
	}

	return style.Reset +
		getLeftBorder(bg, highlightType) +
		content.Render(text) +
		getRightBorder(bg, highlightType)
}

func getLeftBorder(bg color.Color, highlightType int) string {
	return borderStyle(bg, highlightType).Render(nerdfonts.LeftSeparator)
}

func getRightBorder(bg color.Color, highlightType int) string {
	return borderStyle(bg, highlightType).Render(nerdfonts.RightSeparator)
}

func borderStyle(bg color.Color, highlightType int) lipgloss.Style {
	if highlightType == Selected {
		return lipgloss.NewStyle().
			Foreground(lipgloss.NoColor{}).
			Background(bg).
			Reverse(true)
	}

	return lipgloss.NewStyle().
		Foreground(bg)
}

func HighlightHackerNewsHeadlines(title string, highlightType int) string {
	askHN := "Ask HN:"
	showHN := "Show HN:"
	tellHN := "Tell HN:"
	thankHN := "Thank HN:"
	launchHN := "Launch HN:"

	highlight := getHighlight(highlightType)

	title = strings.ReplaceAll(title, askHN, style.HeadlineAskHN(askHN)+highlight)
	title = strings.ReplaceAll(title, showHN, style.HeadlineShowHN(showHN)+highlight)
	title = strings.ReplaceAll(title, tellHN, style.HeadlineTellHN(tellHN)+highlight)
	title = strings.ReplaceAll(title, thankHN, style.HeadlineThankHN(thankHN)+highlight)
	title = strings.ReplaceAll(title, launchHN, style.HeadlineLaunchHN(launchHN)+highlight)

	return title
}

func getHighlight(highlightType int) string {
	switch highlightType {
	case HeadlineInCommentSection:
		return style.ANSIBold
	case Selected:
		return style.Reverse
	case MarkAsRead:
		return style.ANSIFaint + style.Italic
	case AddToFavorites:
		return "\033[32m" + style.Reverse
	case RemoveFromFavorites:
		return "\033[31m" + style.Reverse
	default:
		return ""
	}
}

func HighlightSpecialContent(title string, highlightType int, enableNerdFonts bool) string {
	highlight := getHighlight(highlightType)

	if enableNerdFonts {
		title = strings.ReplaceAll(title, "[audio]", style.HeadlineAudio(nerdfonts.Audio)+highlight)
		title = strings.ReplaceAll(title, "[video]", style.HeadlineVideo(nerdfonts.Video)+highlight)
		title = strings.ReplaceAll(title, "[pdf]", style.HeadlinePDF(nerdfonts.Document)+highlight)
		title = strings.ReplaceAll(title, "[PDF]", style.HeadlinePDF(nerdfonts.Document)+highlight)

		return title
	}

	title = strings.ReplaceAll(title, "[audio]", style.HeadlineAudio("audio")+highlight)
	title = strings.ReplaceAll(title, "[video]", style.HeadlineVideo("video")+highlight)
	title = strings.ReplaceAll(title, "[pdf]", style.HeadlinePDF("pdf")+highlight)
	title = strings.ReplaceAll(title, "[PDF]", style.HeadlinePDF("PDF")+highlight)

	return title
}

var smileys = []struct{ from, to string }{
	{`:)`, "😊"},
	{`(:`, "😊"},
	{`:-)`, "😊"},
	{`:D`, "😄"},
	{`=)`, "😃"},
	{`=D`, "😃"},
	{`;)`, "😉"},
	{`;-)`, "😉"},
	{`:P`, "😜"},
	{`;P`, "😜"},
	{`:o`, "😮"},
	{`:O`, "😮"},
	{`:(`, "😔"},
	{`:-(`, "😔"},
	{`:/`, "😕"},
	{`:-/`, "😕"},
	{`-_-`, "😑"},
	{`:|`, "😐"},
}

func ConvertSmileys(text string) string {
	for _, s := range smileys {
		text = replaceBetweenWhitespace(text, s.from, s.to)
	}

	return text
}

func replaceBetweenWhitespace(text string, target string, replacement string) string {
	if text == target {
		return replacement
	}

	return strings.ReplaceAll(text, " "+target, " "+replacement)
}

func RemoveUnwantedNewLines(text string) string {
	return reUnwantedNewLine.ReplaceAllString(text, `$1`+" "+`$3`)
}

func RemoveUnwantedWhitespace(text string) string {
	singleSpace := " "
	doubleSpace := "  "
	tripleSpace := "   "

	text = strings.ReplaceAll(text, tripleSpace, singleSpace)
	text = strings.ReplaceAll(text, doubleSpace, singleSpace)

	return text
}

func HighlightDomain(domain string) string {
	if domain == "" {
		return style.Reset
	}

	return style.Reset + style.Faint("("+domain+")")
}

func HighlightReferences(input string) string {
	input = strings.ReplaceAll(input, "[0]", "["+style.White("0")+"]")
	input = strings.ReplaceAll(input, "[1]", "["+style.Red("1")+"]")
	input = strings.ReplaceAll(input, "[2]", "["+style.Yellow("2")+"]")
	input = strings.ReplaceAll(input, "[3]", "["+style.Green("3")+"]")
	input = strings.ReplaceAll(input, "[4]", "["+style.Blue("4")+"]")
	input = strings.ReplaceAll(input, "[5]", "["+style.Cyan("5")+"]")
	input = strings.ReplaceAll(input, "[6]", "["+style.Magenta("6")+"]")
	input = strings.ReplaceAll(input, "[7]", "["+style.BrightWhite("7")+"]")
	input = strings.ReplaceAll(input, "[8]", "["+style.BrightRed("8")+"]")
	input = strings.ReplaceAll(input, "[9]", "["+style.BrightYellow("9")+"]")
	input = strings.ReplaceAll(input, "[10]", "["+style.BrightGreen("10")+"]")

	return input
}

func ColorizeIndentSymbol(indentSymbol string, level int) string {
	if level == 0 {
		return style.Reset
	}

	cycle := style.IndentCycle()
	idx := (level - 1) % len(cycle)

	return style.Reset + cycle[idx](indentSymbol)
}

func TrimURLs(comment string, disableCommentHighlighting bool) string {
	if disableCommentHighlighting {
		return reHTMLAnchor.ReplaceAllString(comment, "")
	}

	comment = reHTMLAnchor.ReplaceAllString(comment, "")

	comment = reURL.ReplaceAllString(comment, style.CommentURL(`$1`))

	comment = strings.ReplaceAll(comment, "."+style.Reset+" ", style.Reset+". ")

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

	for range numberOfBackticks + 1 {
		if isOnFirstBacktick {
			input = strings.Replace(input, backtick, style.Italic+style.CommentBacktickColor(), 1)
		} else {
			input = strings.Replace(input, backtick, style.Reset, 1)
		}

		isOnFirstBacktick = !isOnFirstBacktick
	}

	return input
}

func HighlightMentions(input string) string {
	input = reMention.ReplaceAllString(input, style.CommentMention(`$1`))

	input = strings.ReplaceAll(input, style.CommentMention("@dang"),
		style.CommentMod("@dang"))
	input = strings.ReplaceAll(input, style.CommentMention(" @dang"),
		style.CommentMod(" @dang"))

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

	return reVariable.ReplaceAllString(input, style.CommentVariable(`$1`))
}

func HighlightAbbreviations(input string) string {
	iAmNotALawyer := "IANAL"
	iAmALawyer := "IAAL"

	input = strings.ReplaceAll(input, iAmNotALawyer, style.Red(iAmNotALawyer))
	input = strings.ReplaceAll(input, iAmALawyer, style.Green(iAmALawyer))

	return input
}

func ReplaceCharacters(input string) string {
	input = strings.ReplaceAll(input, "&#x27;", "'")
	input = strings.ReplaceAll(input, "&gt;", ">")
	input = strings.ReplaceAll(input, "&lt;", "<")
	input = strings.ReplaceAll(input, "&#x2F;", "/")
	input = strings.ReplaceAll(input, "&quot;", `"`)
	input = strings.ReplaceAll(input, "&#34;", `"`)
	input = strings.ReplaceAll(input, "&amp;", "&")

	return input
}

func ReplaceHTML(input string) string {
	input = strings.TrimPrefix(input, "<p>")

	input = strings.ReplaceAll(input, "<p>", newParagraph)
	input = strings.ReplaceAll(input, "<i>", style.Italic)
	input = strings.ReplaceAll(input, "</i>", style.Reset)
	input = strings.ReplaceAll(input, "</a>", "")
	input = strings.ReplaceAll(input, "<pre><code>", "")
	input = strings.ReplaceAll(input, "</code></pre>", "")

	return input
}

func ReplaceSymbols(paragraph string) string {
	paragraph = strings.ReplaceAll(paragraph, "...", "…")
	paragraph = strings.ReplaceAll(paragraph, "CO2", "CO₂")

	paragraph = replaceDoubleDashesWithEmDash(paragraph)
	paragraph = convertFractions(paragraph)

	return paragraph
}

func replaceDoubleDashesWithEmDash(paragraph string) string {
	paragraph = strings.ReplaceAll(paragraph, " -- ", " — ")

	return reDoubleDash.ReplaceAllString(paragraph, `$1`+"—"+`$2`)
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

	text = strings.ReplaceAll(text, "1/2 ", "½ ")
	text = strings.ReplaceAll(text, "1/3 ", "⅓ ")
	text = strings.ReplaceAll(text, "2/3 ", "⅔ ")
	text = strings.ReplaceAll(text, "1/4 ", "¼ ")
	text = strings.ReplaceAll(text, "3/4 ", "¾ ")
	text = strings.ReplaceAll(text, "1/5 ", "⅕ ")
	text = strings.ReplaceAll(text, "2/5 ", "⅖ ")
	text = strings.ReplaceAll(text, "3/5 ", "⅗ ")
	text = strings.ReplaceAll(text, "4/5 ", "⅘ ")
	text = strings.ReplaceAll(text, "1/6 ", "⅙ ")
	text = strings.ReplaceAll(text, "1/10 ", "⅒  ")

	text = strings.ReplaceAll(text, "1/5th", "⅕th")
	text = strings.ReplaceAll(text, "1/6th", "⅙th")
	text = strings.ReplaceAll(text, "1/10th", "⅒ th")

	return text
}
