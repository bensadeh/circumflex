package constructor

import (
	"clx/constants/margins"
	"strings"

	text "github.com/MichaelMure/go-term-text"
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
)

const (
	helpScreenText = `
           [-:-:b]Main Screen[-:-:-]

Read comments ............................... Enter         
Change category ............................. Tab
Open submission link in browser ............. o
Open comments in browser .................... c
Refresh ..................................... r

Add to favorites ............................ f
Add to favorites by ID ...................... F
Delete from favorites ....................... x
                                    
Bring up this screen ........................ i, ?
Quit to prompt .............................. q             

           [-:-:b]Comment Section (less)[-:-:-]

Down one half-window ........................ d
Up one half-window .......................... u

Go to next top-level comment ................ / + '::'
Repeat last search .......................... n
Repeat last search in reverse direction ..... N

Help screen ................................. h             
Quit to Main Screen ......................... q             
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
	helpScreen.SetPadding(0, 0, -margins.LeftMargin, 0)

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
