package ranking

import (
	"strconv"
	"strings"

	"github.com/bensadeh/circumflex/layout"

	"charm.land/lipgloss/v2"
)

const (
	newParagraph         = "\n\n\n"
	indentationFromRight = " "
)

func Rankings(useRelativeNumbering bool, itemsVisible, itemsTotal, currentPosition, currentPage, totalPages int, readStatuses []bool, faintAll bool) string {
	if itemsTotal == 0 {
		return ""
	}

	if useRelativeNumbering {
		return relativeRankings(itemsVisible, itemsTotal, currentPosition, currentPage, totalPages, faintAll)
	}

	return absoluteRankings(itemsVisible, itemsTotal, currentPage, totalPages, readStatuses, faintAll)
}

var (
	rankStyle            = lipgloss.NewStyle().Width(layout.RankWidth).Align(lipgloss.Right)
	rankFaintStyle       = rankStyle.Faint(true)
	rankFaintItalicStyle = rankStyle.Faint(true).Italic(true)
	faintStyle           = lipgloss.NewStyle().Faint(true)
	faintItalicStyle     = lipgloss.NewStyle().Faint(true).Italic(true)
)

func absoluteRankings(itemsVisible int, itemsTotal int, currentPage int, totalPages int, readStatuses []bool, faintAll bool) string {
	var rankings strings.Builder

	startingRank := itemsVisible*currentPage + 1

	var endingRank int

	onLastPage := currentPage+1 == totalPages

	if onLastPage {
		itemsVisible -= totalPages*itemsVisible - itemsTotal
	}

	endingRank = startingRank + itemsVisible

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

func relativeRankings(itemsVisible int, itemsTotal int, currentPosition int, currentPage int, totalPages int, faintAll bool) string {
	rankOfCurrentlySelectedItem := itemsVisible*currentPage + currentPosition + 1

	onLastPage := currentPage+1 == totalPages
	if onLastPage {
		itemsVisible -= totalPages*itemsVisible - itemsTotal
	}

	var rankings strings.Builder

	end := itemsVisible - currentPosition
	iterator := currentPosition

	for iterator != 0 {
		number := strconv.Itoa(iterator)
		rankings.WriteString(faintStyle.Render(number) + indentationFromRight + newParagraph)

		iterator--
	}

	if faintAll {
		rankings.WriteString(faintItalicStyle.Render(strconv.Itoa(rankOfCurrentlySelectedItem)) + " " + indentationFromRight + newParagraph)
	} else {
		rankings.WriteString(strconv.Itoa(rankOfCurrentlySelectedItem) + " " + indentationFromRight + newParagraph)
	}

	iterator++

	for iterator < end {
		number := strconv.Itoa(iterator)
		rankings.WriteString(faintStyle.Render(number) + indentationFromRight + newParagraph)

		iterator++
	}

	return rankings.String()
}
