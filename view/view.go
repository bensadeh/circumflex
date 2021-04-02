package view

import (
	"clx/constants/help"
	"clx/constants/panels"
	"clx/core"
	"clx/header"
	"clx/pages"
	"time"

	constructor "clx/constructors"

	"gitlab.com/tslocum/cview"
)

func SetHackerNewsHeader(m *core.MainView, header string) {
	m.Header.SetText(header)
}

func SetHelpScreenHeader(m *core.MainView, category int) {
	h := header.GetCircumflexHeader(category)
	m.Header.SetText(h)
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

func ClearStatusBar(m *core.MainView) {
	SetPermanentStatusBar(m, "", cview.AlignCenter)
}

func UpdateSettingsScreen(m *core.MainView) {
	m.Settings.SetText(constructor.GetSettingsText())
}

func UpdateInfoScreen(m *core.MainView) {
	m.InfoScreen.SetText(constructor.GetInfoText())
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

func SetLeftMarginText(m *core.MainView, text string) {
	m.LeftMargin.SetText(text)
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

func SetPageCounter(m *core.MainView, currentPage int, maxPages int) {
	pageCounter := pages.GetPageCounter(currentPage, maxPages)
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

func ClearList(list *cview.List) {
	list.Clear()
}

func SelectItem(list *cview.List, index int) {
	list.SetCurrentItem(index)
}

func ShowItems(list *cview.List, listItems []*cview.ListItem) {
	list.Clear()

	for _, item := range listItems {
		list.AddItem(item)
	}
}
