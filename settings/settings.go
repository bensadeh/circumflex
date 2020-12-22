package settings

import (
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
)

const (
	Submissions    = 0
	CommentSection = 1
)

func GetUnselectableItems() []int {
	return []int{0, 1, 3, 5, 6, 7, 9}
}

func SetSettingsList(list *cview.List, page int){
	if page == 0 {
		SetToSubmissionsSettings(list)
	}
	if page == 1 {
		SetToCommentSectionSettings(list)
	}
}

func SetToSubmissionsSettings(list *cview.List) {
	list.Clear()

	li := cview.NewListItem("[::d]Change")
	li.SetSecondaryText("Comment width: [::b]80")
	list.AddItem(li)

	li = cview.NewListItem(" ")
	li.SetSecondaryText(" ")
	list.AddItem(li)

	li = cview.NewListItem("[::d]Change")
	li.SetSecondaryText("Indent size: [::b]4")
	list.AddItem(li)

	li = cview.NewListItem(" ")
	li.SetSecondaryText(" ")
	list.AddItem(li)

	li = cview.NewListItem("[::d]Change")
	li.SetSecondaryText("Show colors: [black:#82aaff:]yes")
	list.AddItem(li)

	li = cview.NewListItem(" ")
	li.SetSecondaryText("")
	list.AddItem(li)

	li = cview.NewListItem("[::d]Change")
	li.SetSecondaryText("Show labels: [black:orange:]no")
	list.AddItem(li)

	list.SetCurrentItem(2)
}

func SetToCommentSectionSettings(list *cview.List) {
	list.Clear()

	li := cview.NewListItem("[::d]Change")
	li.SetSecondaryText("Comment width: [::b]80")
	list.AddItem(li)

	li = cview.NewListItem(" ")
	li.SetSecondaryText(" ")
	list.AddItem(li)

	li = cview.NewListItem("[::d]Change")
	li.SetSecondaryText("Indent size: [::b]4")
	list.AddItem(li)

	list.SetCurrentItem(2)
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

func GetHeader(page int) string {
	if page == 0 {
		return "SettingsCategory"
	}
	if page == 1 {
		return "SettingsCategory"
	}

	return ""
}
