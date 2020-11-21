package view

import (
	"clx/primitives"
	"clx/types"
	text "github.com/MichaelMure/go-term-text"
	"strconv"
)

func SetHackerNewsHeader(m *primitives.MainView, screenWidth int, category int) {
	switch category {
	case types.NoCategory:
		base := "[black:orange:]   [Y[] [::b]Hacker News[::-]  new | ask | show"
		offset := -26
		header := appendWhitespace(base, offset, screenWidth)
		m.Header.SetText(header)
	case types.New:
		base := "[black:orange:]   [Y[] [::b]Hacker News[::-]  [white]new[black::] | ask | show"
		offset := -42
		header := appendWhitespace(base, offset, screenWidth)
		m.Header.SetText(header)
	case types.Ask:
		base := "[black:orange:]   [Y[] [::b]Hacker News[::-]  new | [white]ask[black::] | show"
		offset := -42
		header := appendWhitespace(base, offset, screenWidth)
		m.Header.SetText(header)
	case types.Show:
		base := "[black:orange:]   [Y[] [::b]Hacker News[::-]  new | ask | [white]show[black::]"
		offset := -42
		header := appendWhitespace(base, offset, screenWidth)
		m.Header.SetText(header)
	default:
		return
	}
}

func appendWhitespace(base string, offset int, screenWidth int) string {
	whitespace := ""
	for i := 0; i < screenWidth-text.Len(base)-offset; i++ {
		whitespace += " "
	}
	return base + whitespace
}

func SetKeymapsHeader(m *primitives.MainView, screenWidth int) {
	base := "[white:rebeccapurple:]   [^] [::b]Keymaps"
	offset := -27
	whitespace := ""
	for i := 0; i < screenWidth-text.Len(base)-offset; i++ {
		whitespace += " "
	}
	m.Header.SetText(base + whitespace)
}

func SetPanelCategory(m *primitives.MainView, category int) {
	c := strconv.Itoa(category)
	m.Panels.SetCurrentPanel(c)
}

func SetPanelToHelpScreen(m *primitives.MainView) {
	m.Panels.SetCurrentPanel("help")
}

func SetLeftMarginRanks(m *primitives.MainView, currentPage int, viewableStoriesOnSinglePage int) {
	marginText := ""
	indentationFromRight := " "
	startingRank := viewableStoriesOnSinglePage*currentPage + 1
	for i := startingRank; i < startingRank+viewableStoriesOnSinglePage; i++ {
		marginText += strconv.Itoa(i) + "." + indentationFromRight + "\n\n"
	}
	m.LeftMargin.SetText(marginText)
}

func HideLeftMarginRanks(m *primitives.MainView) {
	m.LeftMargin.SetText("")
}

func HideFooterText(m *primitives.MainView) {
	m.Footer.SetText("")
}

func SetFooterText(m *primitives.MainView, currentPage int, screenWidth int, maxPages int) {
	if maxPages == 2 {
		footerText := getFooterTextForThreePages(currentPage, screenWidth)
		m.Footer.SetText(footerText)
	} else if maxPages == 1 {
		footerText := getFooterTextForTwoPages(currentPage, screenWidth)
		m.Footer.SetText(footerText)
	}
}

func getFooterTextForThreePages(currentPage int, screenWidth int) string {
	orangeDot := "[orange]" + "•" + "[-:-]"
	footerText := ""

	switch currentPage {
	case 0:
		footerText = "" + orangeDot + "◦◦"
	case 1:
		footerText = "◦" + orangeDot + "◦"
	case 2:
		footerText = "◦◦" + orangeDot + ""
	default:
		footerText = ""
	}
	return padWithWhitespaceFromTheLeft(footerText, screenWidth)
}

func getFooterTextForTwoPages(currentPage int, screenWidth int) string {
	orangeDot := "[orange]" + "•" + "[-:-]"
	footerText := ""

	switch currentPage {
	case 0:
		footerText = "" + orangeDot + "◦ "
	case 1:
		footerText = "◦" + orangeDot + " "
	default:
		footerText = ""
	}
	return padWithWhitespaceFromTheLeft(footerText, screenWidth)
}

func padWithWhitespaceFromTheLeft(s string, screenWidth int) string {
	offset := +10
	whitespace := ""
	for i := 0; i < screenWidth-text.Len(s)+offset; i++ {
		whitespace += " "
	}
	return whitespace + s
}
