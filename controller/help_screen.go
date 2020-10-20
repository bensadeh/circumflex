package controller

import (
	text "github.com/MichaelMure/go-term-text"
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
	"strings"
)

const (
	helpScreenText = `
j, ↓:         down
h, ↑:         up
n, →:         next page
p, ←:         previous page

Enter:        read comments
o:            open in browser
i, ?:         bring up this screen

q:            quit
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
