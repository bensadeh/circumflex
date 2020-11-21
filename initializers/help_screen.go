package builder

import (
	text "github.com/MichaelMure/go-term-text"
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
	"strings"
)

const (
	helpScreenText = `
      [-:-:b]Main Screen[-:-:-]

Enter:         read comments
o:             open submission link in browser
c:             open comments in browser
Tab:           change category
                                    
i, ?:          bring up this screen
q:             quit to prompt

      [-:-:b]Comment Section (less)[-:-:-]

d:             down one half-window
u:             up one half-window

/ + '::':      go to next top-level comment
n:             repeat last search
N:             repeat last search in reverse direction

h:             help screen
q:             quit to Main Screen
`
)

func GetHelpScreen() *cview.TextView {
	helpScreen := cview.NewTextView()
	helpScreen.SetBackgroundColor(tcell.ColorDefault)
	helpScreen.SetTextColor(tcell.ColorDefault)
	helpScreen.SetTextAlign(cview.AlignCenter)
	helpScreen.SetTitleColor(tcell.ColorDefault)
	helpScreen.SetBorderColor(tcell.ColorDefault)
	helpScreen.SetTextColor(tcell.ColorDefault)
	helpScreen.SetDynamicColors(true)
	helpScreen.Box.SetBorderPadding(2, 0, 2, 0)

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
