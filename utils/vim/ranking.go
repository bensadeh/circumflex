package vim

import (
	"clx/utils/format"
	"strconv"
)

const (
	newParagraph         = "\n\n"
	indentationFromRight = " "
)

func AbsoluteRankings(viewableStories int, maxItems int, currentPage int) string {
	rankings := ""

	startingRank := viewableStories*currentPage + 1
	for i := startingRank; i < startingRank+maxItems; i++ {
		rankings += strconv.Itoa(i) + "." + indentationFromRight + newParagraph
	}

	return rankings
}

func RelativeRankings(maxItems int, currentPosition int, currentPage int) string {
	rankings := ""
	end := maxItems - currentPosition
	iterator := currentPosition

	for iterator != 0 {
		number := strconv.Itoa(iterator)
		rankings += format.Dim(number) + indentationFromRight + newParagraph
		iterator--
	}

	rankOfCurrentlySelectedItem := maxItems*currentPage + currentPosition + 1
	rankings += strconv.Itoa(rankOfCurrentlySelectedItem) + " " + indentationFromRight + newParagraph
	iterator++

	for iterator < end {
		number := strconv.Itoa(iterator)
		rankings += format.Dim(number) + indentationFromRight + newParagraph
		iterator++
	}

	return rankings
}
