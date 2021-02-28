package pages

import "strings"

func GetPageCounter(currentPage int, maxPages int) string {
	coloredDot := "[::]•[-::]"
	uncoloredDot := "◦"

	dotsOnTheLeft := strings.Repeat(uncoloredDot, currentPage)
	dotsOnTheRight := strings.Repeat(uncoloredDot, maxPages-currentPage)

	return dotsOnTheLeft + coloredDot + dotsOnTheRight
}
