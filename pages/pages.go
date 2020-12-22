package pages

import "strings"

func GetPageCounter(currentPage int, maxPages int, color string) string {
	coloredDot := "[" + color + "]•[-::]"
	uncoloredDot := "◦"

	dotsOnTheLeft := strings.Repeat(uncoloredDot, currentPage)
	dotsOnTheRight := strings.Repeat(uncoloredDot, maxPages-currentPage)

	return dotsOnTheLeft + coloredDot + dotsOnTheRight
}
