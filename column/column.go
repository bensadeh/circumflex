package column

import (
	text "github.com/MichaelMure/go-term-text"
	"strings"
)

const (
	newLine = "\n"
)

func PutInColumns(leftCol string, rightCol string, colWidth int, spaceWidth int) string {
	space := text.LeftPadMaxLine("", spaceWidth, 0)
	leftCol, _ = text.Wrap(leftCol, colWidth)
	rightCol, _ = text.Wrap(rightCol, colWidth)

	linesA := strings.Split(leftCol, newLine)
	linesB := strings.Split(rightCol, newLine)

	length := max(len(linesA), len(linesB))

	output := ""
	for i := 0; i < length; i++ {
		if i >= len(linesA) {
			output += text.LeftPadMaxLine("", colWidth, 0) + space +
				text.LeftPadMaxLine(linesB[i], colWidth, 0) + newLine
		} else if i >= len(linesB) {
			output += text.LeftPadMaxLine(linesA[i], colWidth, 0) + space +
				text.LeftPadMaxLine("", colWidth, 0) + newLine
		} else {
			output += text.LeftPadMaxLine(linesA[i], colWidth, 0) + space +
				text.LeftPadMaxLine(linesB[i], colWidth, 0) + newLine
		}

	}

	return output
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}
