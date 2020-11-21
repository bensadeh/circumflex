package builder

import (
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
)

func NewList() *cview.List {
	list := cview.NewList()
	list.SetBackgroundTransparent(false)
	list.SetBackgroundColor(tcell.ColorDefault)
	list.SetMainTextColor(tcell.ColorDefault)
	list.SetSecondaryTextColor(tcell.ColorDefault)
	list.ShowSecondaryText(true)

	return list
}
