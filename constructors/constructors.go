package constructor

import (
	"clx/constants/margins"
	"clx/constants/panels"
	"clx/core"
	"clx/screen"

	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
)

const (
	maximumStoriesToDisplay = 30
)

func NewScreenController() *core.ScreenController {
	sc := new(core.ScreenController)
	sc.Application = cview.NewApplication()

	sc.ApplicationState = new(core.ApplicationState)
	sc.ApplicationState.ScreenWidth = screen.GetTerminalWidth()
	sc.ApplicationState.ScreenHeight = screen.GetTerminalHeight()
	sc.ApplicationState.StoriesToShow = screen.GetSubmissionsToShow(
		sc.ApplicationState.ScreenHeight,
		maximumStoriesToDisplay)

	sc.Articles = NewList()

	sc.MainView = NewMainView()
	sc.MainView.Panels.AddPanel(panels.StoriesPanel, sc.Articles, true, true)

	return sc
}

func NewList() *cview.List {
	list := cview.NewList()
	list.SetBackgroundTransparent(false)
	list.SetBackgroundColor(tcell.ColorDefault)
	list.SetMainTextColor(tcell.ColorDefault)
	list.SetSecondaryTextColor(tcell.ColorDefault)
	list.SetSelectedTextAttributes(tcell.AttrReverse)
	list.SetSelectedTextColor(tcell.ColorDefault)
	list.SetSelectedBackgroundColor(tcell.ColorDefault)
	list.SetScrollBarVisibility(cview.ScrollBarNever)

	return list
}

func NewMainView() *core.MainView {
	main := new(core.MainView)
	main.Panels = cview.NewPanels()
	main.Grid = cview.NewGrid()
	main.LeftMargin = newTextViewPrimitive()
	main.LeftMargin.SetTextAlign(cview.AlignRight)
	main.Header = newTextViewPrimitive()
	main.PageCounter = newTextViewPrimitive()
	main.StatusBar = newTextViewPrimitive()
	main.StatusBar.SetTextAlign(cview.AlignCenter)
	main.StatusBar.SetPadding(0, 0, -4, 0)
	main.InfoScreen = newTextViewPrimitive()

	flex, inputField := newFavoritesFlex()

	main.CustomFavorite = inputField

	main.Grid.SetBorder(false)
	main.Grid.SetRows(2, 0, 1)
	main.Grid.SetColumns(margins.LeftMargin, 0, margins.RightMarginPageCounter)
	main.Grid.SetBackgroundColor(tcell.ColorDefault)
	main.Grid.AddItem(main.Header, 0, 0, 1, 3, 0, 0, true)
	main.Grid.AddItem(main.LeftMargin, 1, 0, 1, 1, 0, 0, false)
	main.Grid.AddItem(main.Panels, 1, 1, 1, 2, 0, 0, false)
	main.Grid.AddItem(main.StatusBar, 2, 1, 1, 1, 0, 0, false)
	main.Grid.AddItem(main.PageCounter, 2, 2, 1, 1, 0, 0, false)

	main.Panels.AddPanel(panels.InfoPanel, main.InfoScreen, true, false)
	main.Panels.AddPanel(panels.AddCustomFavoritePanel, flex, true, false)

	return main
}

func newTextViewPrimitive() *cview.TextView {
	tv := cview.NewTextView()
	tv.SetTextAlign(cview.AlignLeft)
	tv.SetBorder(false)
	tv.SetBackgroundColor(tcell.ColorDefault)
	tv.SetTextColor(tcell.ColorDefault)
	tv.SetDynamicColors(true)
	tv.SetScrollBarVisibility(cview.ScrollBarNever)

	return tv
}

func newFavoritesFlex() (*cview.Flex, *cview.InputField) {
	inputField := cview.NewInputField()
	inputField.SetTitle("Add to favorites by ID")
	inputField.SetTitleColor(tcell.ColorDefault)
	inputField.SetTitleAlign(cview.AlignCenter)

	inputField.SetBorder(true)
	inputField.SetBorderColor(tcell.ColorDefault)

	inputField.SetBackgroundColor(tcell.ColorDefault)

	inputField.SetLabel("ID: ")
	inputField.SetLabelColor(tcell.ColorDefault)

	inputField.SetFieldTextColor(tcell.ColorDefault)
	inputField.SetFieldBackgroundColor(tcell.ColorDefault)

	inputField.SetFieldWidth(20)
	inputField.SetPadding(1, 0, 5, 5)

	inputField.SetAcceptanceFunc(cview.InputFieldInteger)
	inputField.SetDoneFunc(func(key tcell.Key) {
		inputField.GetText()
	})

	subFlex := cview.NewFlex()
	subFlex.SetDirection(cview.FlexRow)
	subFlex.AddItem(demoBox(), 0, 1, false)
	subFlex.AddItem(inputField, 5, 0, false)
	subFlex.AddItem(demoBox(), 0, 1, false)

	flex := cview.NewFlex()
	flex.AddItem(demoBox(), 0, 1, false)
	flex.AddItem(subFlex, 30, 2, false)
	flex.AddItem(demoBox(), 0, 1, false)
	flex.AddItem(demoBox(), margins.LeftMargin, 0, false)

	return flex, inputField
}

func demoBox() *cview.Box {
	b := cview.NewBox()
	b.SetBorder(false)
	b.SetBackgroundColor(tcell.ColorDefault)

	return b
}
