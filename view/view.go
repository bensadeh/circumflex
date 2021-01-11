package view

import (
	"clx/constants/help"
	"clx/constants/panels"
	"clx/constants/submissions"
	constructor "clx/constructors"
	"clx/core"
	"clx/pages"
	text "github.com/MichaelMure/go-term-text"
	"gitlab.com/tslocum/cview"
	"strconv"
	"time"
)

const (
	black = "#0c0c0c"
)

func SetHackerNewsHeader(m *core.MainView, screenWidth int, category int) {
	switch category {
	case submissions.FrontPage:
		base := "[" + black + ":orange:]   [Y[] [::b]Hacker News[::-]  new | ask | show"
		offset := -28
		header := appendWhitespace(base, offset, screenWidth)
		m.Header.SetText(header)
	case submissions.New:
		base := "[" + black + ":orange:]   [Y[] [::b]Hacker News[::-]  [white]new[" + black + "::] | ask | show"
		offset := -46
		header := appendWhitespace(base, offset, screenWidth)
		m.Header.SetText(header)
	case submissions.Ask:
		base := "[" + black + ":orange:]   [Y[] [::b]Hacker News[::-]  new | [white]ask[" + black + "::] | show"
		offset := -46
		header := appendWhitespace(base, offset, screenWidth)
		m.Header.SetText(header)
	case submissions.Show:
		base := "[" + black + ":orange:]   [Y[] [::b]Hacker News[::-]  new | ask | [white]show[" + black + "::]"
		offset := -46
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

func SetHelpScreenHeader(m *core.MainView, screenWidth int, category int) {
	switch category {
	case help.Info:
		base := "[" + black + ":#82aaff:]   [^] [::b]circumflex[::-]   keymaps | settings"
		offset := -28
		header := appendWhitespace(base, offset, screenWidth)
		m.Header.SetText(header)
	case help.Keymaps:
		base := "[" + black + ":#82aaff:]   [^] [::b]circumflex[::-]   [white]keymaps[" + black + "::] | settings"
		offset := -46
		header := appendWhitespace(base, offset, screenWidth)
		m.Header.SetText(header)
	case help.Settings:
		base := "[" + black + ":#82aaff:]   [^] [::b]circumflex[::-]   keymaps | [white]settings[" + black + "::]"
		offset := -46
		header := appendWhitespace(base, offset, screenWidth)
		m.Header.SetText(header)
	default:
		return
	}
}

func SetPanelToSubmissions(m *core.MainView) {
	m.Panels.SetCurrentPanel(panels.SubmissionsPanel)
}

func SetHelpScreenPanel(m *core.MainView, category int) {
	switch category {
	case help.Info:
		m.Panels.SetCurrentPanel(panels.InfoPanel)
	case help.Keymaps:
		m.Panels.SetCurrentPanel(panels.KeymapsPanel)
	case help.Settings:
		m.Panels.SetCurrentPanel(panels.SettingsPanel)
	default:
		return
	}
}

func HideStatusBar(m *core.MainView) {
	SetPermanentStatusBar(m, "", cview.AlignCenter)
}

func UpdateSettingsScreen(m *core.MainView) {
	m.Settings.SetText(constructor.GetSettingsText())
}

func SetPermanentStatusBar(m *core.MainView, text string, align int) {
	m.StatusBar.SetTextAlign(align)
	m.StatusBar.SetText(text)
}

func SetTemporaryStatusBar(app *cview.Application, m *core.MainView, text string, duration time.Duration) {
	go setAndClearStatusBar(app, m, text, duration)
}

func setAndClearStatusBar(app *cview.Application, m *core.MainView, text string, duration time.Duration) {
	m.StatusBar.SetText(text)
	time.Sleep(duration)
	m.StatusBar.SetText("")
	app.Draw()
}

func SetLeftMarginRanks(m *core.MainView, currentPage int, viewableStoriesOnSinglePage int) {
	marginText := ""
	indentationFromRight := " "
	startingRank := viewableStoriesOnSinglePage*currentPage + 1
	for i := startingRank; i < startingRank+viewableStoriesOnSinglePage; i++ {
		marginText += strconv.Itoa(i) + "." + indentationFromRight + "\n\n"
	}
	m.LeftMargin.SetText(marginText)
}

func HideLeftMarginRanks(m *core.MainView) {
	m.LeftMargin.SetText("")
}

func HidePageCounter(m *core.MainView) {
	m.PageCounter.SetText("")
}

func ScrollSettingsOneLineUp(m *core.MainView) {
	row, col := m.Settings.GetScrollOffset()
	m.Settings.ScrollTo(row-1, col)
}

func ScrollSettingsOneLineDown(m *core.MainView) {
	row, col := m.Settings.GetScrollOffset()
	m.Settings.ScrollTo(row+1, col)
}

func ScrollSettingsByAmount(m *core.MainView, amount int) {
	row, col := m.Settings.GetScrollOffset()
	m.Settings.ScrollTo(row+amount, col)
}

func ScrollSettingsToBeginning(m *core.MainView) {
	m.Settings.ScrollToBeginning()
}

func ScrollSettingsToEnd(m *core.MainView) {
	m.Settings.ScrollToEnd()
}

func SetPageCounter(m *core.MainView, currentPage int, maxPages int, color string) {
	pageCounter := pages.GetPageCounter(currentPage, maxPages, color)
	m.PageCounter.SetText(pageCounter)
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
