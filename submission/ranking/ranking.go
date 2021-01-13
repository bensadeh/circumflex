package ranking

import "strconv"

const (
	newParagraph         = "\n\n"
	indentationFromRight = " "
)

func AbsoluteRankings(viewableStoriesOnSinglePage int, currentPage int) string {
	rankings := ""

	startingRank := viewableStoriesOnSinglePage*currentPage + 1
	for i := startingRank; i < startingRank+viewableStoriesOnSinglePage; i++ {
		rankings += strconv.Itoa(i) + "." + indentationFromRight + newParagraph
	}

	return rankings
}

func RelativeRankings(viewableStoriesOnSinglePage int, currentPosition int, currentPage int) string {
	rankings := ""
	end := viewableStoriesOnSinglePage - currentPosition
	iterator := currentPosition

	for iterator != 0 {
		number := strconv.Itoa(iterator)
		rankings += dim(number) + indentationFromRight + newParagraph
		iterator--
	}

	rankOfCurrentlySelectedItem := viewableStoriesOnSinglePage*currentPage + currentPosition + 1
	rankings += strconv.Itoa(rankOfCurrentlySelectedItem) + " " + indentationFromRight + newParagraph
	iterator++

	for iterator < end {
		number := strconv.Itoa(iterator)
		rankings += dim(number) + indentationFromRight + newParagraph
		iterator++
	}

	return rankings
}

func dim(text string) string {
	return "[::d]" + text + "[::-]"
}
