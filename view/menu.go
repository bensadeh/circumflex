package view

import (
	"clx/types"
	text "github.com/MichaelMure/go-term-text"
	"gitlab.com/tslocum/cview"
	"strconv"
)

func SetHackerNewsHeader(m *types.MainView, screenWidth int, category int) {
	switch category {
	case types.FrontPage:
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

func SetKeymapsHeader(m *types.MainView, screenWidth int) {
	base := "[#292D3E:#82aaff:b]       Keymaps"
	offset := -19
	whitespace := ""
	for i := 0; i < screenWidth-text.Len(base)-offset; i++ {
		whitespace += " "
	}
	m.Header.SetText(base + whitespace)
}

func SetPanelCategory(m *types.MainView, category int) {
	c := strconv.Itoa(category)
	m.Panels.SetCurrentPanel(c)
}

func SetPanelToHelpScreen(m *types.MainView) {
	m.Panels.SetCurrentPanel("help")
}

func SetLeftMarginRanks(m *types.MainView, currentPage int, viewableStoriesOnSinglePage int) {
	marginText := ""
	indentationFromRight := " "
	startingRank := viewableStoriesOnSinglePage*currentPage + 1
	for i := startingRank; i < startingRank+viewableStoriesOnSinglePage; i++ {
		marginText += strconv.Itoa(i) + "." + indentationFromRight + "\n\n"
	}
	m.LeftMargin.SetText(marginText)
}

func HideLeftMarginRanks(m *types.MainView) {
	m.LeftMargin.SetText("")
}

func HideFooterText(m *types.MainView) {
	m.PageIndicator.SetText("")
}

func SetPageCounter(m *types.MainView, currentPage int, maxPages int) {
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

func SelectFirstElementInList(main *types.MainView) {
	list := getListFromFrontPanel(main.Panels)
	list.SetCurrentItem(0)

}

func SelectLastElementInList(main *types.MainView, appState *types.ApplicationState) {
	list := getListFromFrontPanel(main.Panels)
	list.SetCurrentItem(appState.SubmissionsToShow)
}

func SelectElementInList(main *types.MainView, index int) {
	list := getListFromFrontPanel(main.Panels)
	list.SetCurrentItem(index)
}

func getListFromFrontPanel(pages *cview.Panels) *cview.List {
	_, primitive := pages.GetFrontPanel()
	list, _ := primitive.(*cview.List)
	return list
}
