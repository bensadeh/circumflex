package constructor

import (
	"gitlab.com/tslocum/cview"
)

func NewSettingsList() *cview.List {
	settings := NewList()
	settings.SetBorder(false)

	settings = NewList()
	li := cview.NewListItem("Comment Width")
	li.SetSecondaryText("[::d]Currently set to 80")
	settings.AddItem(li)

	li = cview.NewListItem(" ")
	li.SetSecondaryText(" ")
	settings.AddItem(li)

	li = cview.NewListItem("Indentation size")
	li.SetSecondaryText("[::d]Currently set to 4")
	settings.AddItem(li)

	li = cview.NewListItem("")
	li.SetSecondaryText("")
	settings.AddItem(li)

	li = cview.NewListItem("Indentation size")
	li.SetSecondaryText("[::d]Currently set to 4")
	settings.AddItem(li)

	li = cview.NewListItem(" ")
	li.SetSecondaryText("")
	settings.AddItem(li)

	li = cview.NewListItem("Enable colors")
	li.SetSecondaryText("[::d]Currently set to true")
	settings.AddItem(li)

	return settings
}
