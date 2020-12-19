package view

import (
	"clx/constants"
	"clx/structs"
	text "github.com/MichaelMure/go-term-text"
	"gitlab.com/tslocum/cview"
	"strconv"
	"time"
)

func SetHackerNewsHeader(m *structs.MainView, screenWidth int, category int) {
	switch category {
	case constants.FrontPage:
		base := "[black:orange:]   [Y[] [::b]Hacker News[::-]  new | ask | show"
		offset := -26
		header := appendWhitespace(base, offset, screenWidth)
		m.Header.SetText(header)
	case constants.New:
		base := "[black:orange:]   [Y[] [::b]Hacker News[::-]  [white]new[black::] | ask | show"
		offset := -42
		header := appendWhitespace(base, offset, screenWidth)
		m.Header.SetText(header)
	case constants.Ask:
		base := "[black:orange:]   [Y[] [::b]Hacker News[::-]  new | [white]ask[black::] | show"
		offset := -42
		header := appendWhitespace(base, offset, screenWidth)
		m.Header.SetText(header)
	case constants.Show:
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

func SetHelpScreenHeader(m *structs.MainView, screenWidth int, category int) {
	switch category {
	case constants.Info:
		base := "[black:#82aaff:]   [^] [::b]circumflex[::-]   keymaps | settings"
		offset := -26
		header := appendWhitespace(base, offset, screenWidth)
		m.Header.SetText(header)
	case constants.Keymaps:
		base := "[black:#82aaff:]   [^] [::b]circumflex[::-]   [white]keymaps[black::] | settings"
		offset := -42
		header := appendWhitespace(base, offset, screenWidth)
		m.Header.SetText(header)
	case constants.Environment:
		base := "[black:#82aaff:]   [^] [::b]circumflex[::-]   keymaps | [white]settings[black::]"
		offset := -42
		header := appendWhitespace(base, offset, screenWidth)
		m.Header.SetText(header)
	default:
		return
	}
}

func SetPanelToSubmissions(m *structs.MainView) {
	m.Panels.SetCurrentPanel(constants.SubmissionsPanel)
}

func SetHelpScreenPanel(m *structs.MainView, category int) {
	switch category {
	case constants.Info:
		m.Panels.SetCurrentPanel(constants.InfoPanel)
	case constants.Keymaps:
		m.Panels.SetCurrentPanel(constants.KeymapsPanel)
	case constants.Environment:
		m.Panels.SetCurrentPanel(constants.EnvironmentPanel)
	default:
		return
	}
}

func SetPermanentStatusBar(m *structs.MainView, text string) {
	m.StatusBar.SetText(text)
}

func SetTemporaryStatusBar(app *cview.Application, m *structs.MainView, text string, duration time.Duration) {
	go setAndClearStatusBar(app, m, text, duration)
}

func setAndClearStatusBar(app *cview.Application, m *structs.MainView, text string, duration time.Duration) {
	m.StatusBar.SetText(text)
	time.Sleep(duration)
	m.StatusBar.SetText("")
	app.Draw()
}

func SetLeftMarginRanks(m *structs.MainView, currentPage int, viewableStoriesOnSinglePage int) {
	marginText := ""
	indentationFromRight := " "
	startingRank := viewableStoriesOnSinglePage*currentPage + 1
	for i := startingRank; i < startingRank+viewableStoriesOnSinglePage; i++ {
		marginText += strconv.Itoa(i) + "." + indentationFromRight + "\n\n"
	}
	m.LeftMargin.SetText(marginText)
}

func HideLeftMarginRanks(m *structs.MainView) {
	m.LeftMargin.SetText("")
}

func HideFooterText(m *structs.MainView) {
	m.PageIndicator.SetText("")
}

func SetPageCounter(m *structs.MainView, currentPage int, maxPages int) {
	pageCounter := ""

	if maxPages == 2 {
		pageCounter = getPageCounterForThreePages(currentPage)
	} else if maxPages == 1 {
		pageCounter = getPageCounterForTwoPages(currentPage)
	}

	m.PageIndicator.SetText(pageCounter)
}

func getPageCounterForThreePages(currentPage int) string {
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
	return footerText
}

func getPageCounterForTwoPages(currentPage int) string {
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
	return footerText
}

func SelectFirstElementInList(list *cview.List) {
	firstElement := 0
	list.SetCurrentItem(firstElement)
}

func SelectLastElementInList(list *cview.List) {
	lastElement := -1
	list.SetCurrentItem(lastElement)
}

func SelectElementInList(list *cview.List, index int) {
	list.SetCurrentItem(index)
}
