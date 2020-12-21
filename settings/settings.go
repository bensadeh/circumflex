package settings

import (
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
)

func GetUnselectableItems() []int {
	return []int{0, 1, 3, 5, 6, 7, 9}
}
func NewSettingsList() *cview.List {
	settings := newList()
	settings.SetBorder(false)

	li := cview.NewListItem("")
	li.SetSecondaryText("Front Page")
	settings.AddItem(li)

	li = cview.NewListItem("")
	li.SetSecondaryText("")
	settings.AddItem(li)

	li = cview.NewListItem("[::d]Change")
	li.SetSecondaryText("Comment width: [::b]80")
	settings.AddItem(li)

	li = cview.NewListItem(" ")
	li.SetSecondaryText(" ")
	settings.AddItem(li)

	li = cview.NewListItem("[::d]Change")
	li.SetSecondaryText("Indent size: [::b]4")
	settings.AddItem(li)

	li = cview.NewListItem(" ")
	li.SetSecondaryText(" ")
	settings.AddItem(li)

	li = cview.NewListItem("")
	li.SetSecondaryText("Comment Section")
	settings.AddItem(li)

	li = cview.NewListItem("")
	li.SetSecondaryText("")
	settings.AddItem(li)

	li = cview.NewListItem("[::d]Change")
	li.SetSecondaryText("Show colors: [black:#82aaff:]yes")
	settings.AddItem(li)

	li = cview.NewListItem(" ")
	li.SetSecondaryText("")
	settings.AddItem(li)

	li = cview.NewListItem("[::d]Change")
	li.SetSecondaryText("Show labels: [black:orange:]no")
	settings.AddItem(li)

	settings.SetCurrentItem(2)

	return settings
}

func NewDialogueBox() *cview.Modal {
	modal := cview.NewModal()
	modal.SetText("Do you want to quit the application? " +
		"Do you want to quit the application? Do you want to quit the application?")
	modal.AddButtons([]string{"Quit", "Cancel"})
	modal.SetBackgroundColor(tcell.ColorDefault)
	modal.SetTextColor(tcell.ColorDefault)
	
	return modal
}

func newList() *cview.List {
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
