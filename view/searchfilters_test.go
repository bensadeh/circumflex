package view

import (
	"testing"
	"time"

	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/hn"

	"github.com/stretchr/testify/assert"
)

func TestSearchFilters_DefaultsMirrorTheSite(t *testing.T) {
	var f searchFilters

	groups := f.headerGroups()
	assert.Equal(t, []header.OptionGroup{
		{Options: []string{"popularity", "date"}, Active: 0},
		{Options: []string{"all", "24h", "week", "month", "year"}, Active: 0},
	}, groups)

	req := f.request("gpu", 30)
	assert.Equal(t, hn.SearchRequest{Query: "gpu", ItemsToFetch: 30}, req)
}

func TestSearchFilters_CyclesWrapAround(t *testing.T) {
	var f searchFilters

	f.cycleSort()
	assert.True(t, f.request("q", 1).SortByDate)

	f.cycleSort()
	assert.False(t, f.request("q", 1).SortByDate)

	for range searchAges {
		f.cycleAge()
	}

	assert.Equal(t, time.Duration(0), f.request("q", 1).MaxAge, "a full cycle returns to all time")
}

func TestSearchFilters_GroupsTrackState(t *testing.T) {
	var f searchFilters

	f.cycleSort()
	f.cycleAge()

	groups := f.headerGroups()
	assert.Equal(t, 1, groups[0].Active, "sort highlight moved to date")
	assert.Equal(t, 1, groups[1].Active, "age highlight moved to 24h")
	assert.Equal(t, 24*time.Hour, f.request("q", 1).MaxAge)
}
