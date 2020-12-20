package constructor

import (
	"gitlab.com/tslocum/cview"
)

func NewSettingsList() *cview.List {
	settings := NewList()
	settings.SetBorder(false)

	li := cview.NewListItem("")
	li.SetSecondaryText("Submissions")
	settings.AddItem(li)

	li = cview.NewListItem("")
	li.SetSecondaryText("")
	settings.AddItem(li)

	li = cview.NewListItem("[::d]Change")
	li.SetSecondaryText("Comment width: 80")
	settings.AddItem(li)

	li = cview.NewListItem(" ")
	li.SetSecondaryText(" ")
	settings.AddItem(li)

	li = cview.NewListItem("[::d]Change")
	li.SetSecondaryText("Indent size: 4")
	settings.AddItem(li)

	li = cview.NewListItem(" ")
	li.SetSecondaryText(" ")
	settings.AddItem(li)

	li = cview.NewListItem("")
	li.SetSecondaryText("Comment section")
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

	return settings
}
