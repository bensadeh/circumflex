package primitives

import (
	"gitlab.com/tslocum/cview"
)

type ListController struct {
	News []*cview.List
	Ask  []*cview.List
	Show []*cview.List
}

func NewListController(maxNumberOfPages int) *ListController {
	lc := new(ListController)

	lc.News = createList(maxNumberOfPages)
	lc.Ask = createList(maxNumberOfPages)
	lc.Show = createList(maxNumberOfPages)

	return lc
}

func createList(maxNumberOfPages int) []*cview.List {
	var list []*cview.List

	for i := 0; i < maxNumberOfPages; i++ {
		list = append(list, cview.NewList())
	}

	return list
}
