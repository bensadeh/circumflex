package view

import (
	"time"

	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/hn"
)

// searchFilters is the search mode's filter state, mirroring the site's
// dropdowns: how to rank and how far back. The zero value is the site's
// default — by popularity, all time.
type searchFilters struct {
	byDate bool
	ageIdx int
}

var searchSorts = []string{"popularity", "date"}

var searchAges = []struct {
	label  string
	maxAge time.Duration
}{
	{"all", 0},
	{"24h", 24 * time.Hour},
	{"week", 7 * 24 * time.Hour},
	{"month", 30 * 24 * time.Hour},
	{"year", 365 * 24 * time.Hour},
}

func (f *searchFilters) cycleSort() { f.byDate = !f.byDate }

func (f *searchFilters) cycleAge() { f.ageIdx = (f.ageIdx + 1) % len(searchAges) }

// headerGroups lays the filters out for the search header's segmented
// controls, every option visible with the active one highlighted.
func (f *searchFilters) headerGroups() []header.OptionGroup {
	sortIdx := 0
	if f.byDate {
		sortIdx = 1
	}

	ages := make([]string, len(searchAges))
	for i, age := range searchAges {
		ages[i] = age.label
	}

	return []header.OptionGroup{
		{Options: searchSorts, Active: sortIdx},
		{Options: ages, Active: f.ageIdx},
	}
}

// request assembles the search the filters currently describe.
func (f *searchFilters) request(query string, itemsToFetch int) hn.SearchRequest {
	return hn.SearchRequest{
		Query:        query,
		SortByDate:   f.byDate,
		MaxAge:       searchAges[f.ageIdx].maxAge,
		ItemsToFetch: itemsToFetch,
	}
}
