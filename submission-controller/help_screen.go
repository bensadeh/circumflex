package submission_controller

import (
	text "github.com/MichaelMure/go-term-text"
	"github.com/gdamore/tcell"
	"gitlab.com/tslocum/cview"
	"strings"
)

const (
	helpScreenText = `
j, ↓:          down
h, ↑:          up

Enter:         read comments
o:             open in browser
q:             quit
h:             bring up this screen

Ctrl + n:      next page
Ctrl + p:      previous page
`
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

	helpScreen.SetText(padLines(helpScreenText))

	return helpScreen
}

func padLines(s string) string {
	newLine := "\n"
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
