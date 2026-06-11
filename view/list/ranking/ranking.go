package ranking

import (
	"strconv"
	"strings"

	"github.com/bensadeh/circumflex/layout"

	"charm.land/lipgloss/v2"
)

const newParagraph = "\n\n\n"

var (
	rankStyle            = lipgloss.NewStyle().Width(layout.RankWidth).Align(lipgloss.Right)
	rankFaintStyle       = rankStyle.Faint(true)
	rankFaintItalicStyle = rankStyle.Faint(true).Italic(true)
)

func Rankings(itemsVisible, itemsTotal, currentPage, totalPages int, readStatuses []bool, faintAll bool) string {
	if itemsTotal == 0 {
		return ""
	}

	var rankings strings.Builder

	startingRank := itemsVisible*currentPage + 1

	onLastPage := currentPage+1 == totalPages
	if onLastPage {
		itemsVisible -= totalPages*itemsVisible - itemsTotal
	}

	endingRank := startingRank + itemsVisible

	for i := startingRank; i < endingRank; i++ {
		idx := i - startingRank

		s := rankStyle

		switch {
		case faintAll:
			s = rankFaintItalicStyle
		case idx < len(readStatuses) && readStatuses[idx]:
			s = rankFaintStyle
		}

		rank := s.Render(strconv.Itoa(i)+".") + " "
		rankings.WriteString(rank + newParagraph)
	}

	return strings.TrimSuffix(rankings.String(), "\n\n")
}
