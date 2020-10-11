package submission_controller

import (
	text "github.com/MichaelMure/go-term-text"
	"github.com/gdamore/tcell"
	"gitlab.com/tslocum/cview"
	"strings"
)

const (
	newLine = "\n"
)

func getHelpScreen() *cview.TextView {
	helpScreen := cview.NewTextView()
	helpScreen.SetBackgroundColor(tcell.ColorDefault)
	helpScreen.SetTextColor(tcell.ColorDefault)
	helpScreen.SetTextAlign(cview.AlignCenter)
	helpScreen.SetTitle("circumflex")
	helpScreen.SetTitleColor(tcell.ColorDefault)
	helpScreen.SetBorderColor(tcell.ColorDefault)
	helpScreen.SetTextColor(tcell.ColorDefault)
	helpScreen.Box.SetBorderPadding(10, 10, 10, 10)
	helpScreen.Box.SetBorder(true)
	helpScreen.Box.SetBorderAttributes(tcell.AttrDim)

	t := ""

	t += "j, ↓:          down" + newLine
	t += "h, ↑:          up" + newLine
	t += "" + newLine
	t += "Enter:         read comments" + newLine
	t += "o:             open in browser" + newLine
	t += "q:             quit" + newLine
	t += "h:             bring up this screen" + newLine
	t += "" + newLine
	t += "Ctrl + n:      next page" + newLine
	t += "Ctrl + p:      previous page" + newLine

	helpScreen.SetText(padLines(t))

	return helpScreen
}

func padLines(s string) string {
	maxWidth := text.MaxLineLen(s)
	lines := strings.Split(s, newLine)
	paddedLines := ""

	for _, line := range lines {
		paddedLines += padString(line, maxWidth) + newLine
	}

	return paddedLines
}

func padString(s string, maxWidth int) string {
	paddedString := s

	for i := 0; i < maxWidth-text.Len(s); i++ {
		paddedString += " "
	}
	return paddedString
}
