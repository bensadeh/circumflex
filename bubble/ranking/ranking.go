package ranking

import (
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/logrusorgru/aurora/v3"
)

const (
	newParagraph         = "\n\n\n"
	indentationFromRight = " "
)

func GetRankings(useRelativeNumbering bool, itemsVisible, itemsTotal, currentPosition, currentPage, totalPages int) string {
	if itemsTotal == 0 {
		return ""
	}

	if useRelativeNumbering {
		return relativeRankings(itemsVisible, itemsTotal, currentPosition, currentPage, totalPages)
	}

	return absoluteRankings(itemsVisible, itemsTotal, currentPage, totalPages)
}

func absoluteRankings(itemsVisible int, itemsTotal int, currentPage int, totalPages int) string {
	var rankings strings.Builder

	startingRank := itemsVisible*currentPage + 1
	endingRank := 0
	onLastPage := currentPage+1 == totalPages

	if onLastPage {
		itemsVisible = itemsVisible - (totalPages*itemsVisible - itemsTotal)
		endingRank = startingRank + itemsVisible
	} else {
		endingRank = startingRank + itemsVisible
	}

	for i := startingRank; i < endingRank; i++ {
		rank := lipgloss.NewStyle().Width(6).Align(lipgloss.Right).Render(strconv.Itoa(i)+".") + " "
		rankings.WriteString(rank + newParagraph)
	}

	return strings.TrimSuffix(rankings.String(), "\n\n")
}

func relativeRankings(itemsVisible int, itemsTotal int, currentPosition int, currentPage int, totalPages int) string {
	rankOfCurrentlySelectedItem := itemsVisible*currentPage + currentPosition + 1
	onLastPage := currentPage+1 == totalPages
	if onLastPage {
		itemsVisible = itemsVisible - (totalPages*itemsVisible - itemsTotal)
	}

	var rankings strings.Builder
	end := itemsVisible - currentPosition
	iterator := currentPosition

	for iterator != 0 {
		number := strconv.Itoa(iterator)
		rankings.WriteString(aurora.Faint(number).String() + indentationFromRight + newParagraph)
		iterator--
	}

	rankings.WriteString(strconv.Itoa(rankOfCurrentlySelectedItem) + " " + indentationFromRight + newParagraph)
	iterator++

	for iterator < end {
		number := strconv.Itoa(iterator)
		rankings.WriteString(aurora.Faint(number).String() + indentationFromRight + newParagraph)
		iterator++
	}

	return rankings.String()
}
