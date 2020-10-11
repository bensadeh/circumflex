package submission_controller

import (
	text "github.com/MichaelMure/go-term-text"
	"github.com/gdamore/tcell"
	"gitlab.com/tslocum/cview"
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
	t += padString("j, ↓:          down")
	t += padString("h, ↑:          up")
	t += padString("")
	t += padString("Enter:         read comments" )
	t += padString("o:             open in browser" )
	t += padString("q:             quit" )
	t += padString("h:             bring up this screen" )
	t += padString("")
	t += padString("Ctrl + n:      next page" )
	t += padString("Ctrl + p:      previous page" )

	helpScreen.SetText(t)

	return helpScreen
}

func padString(s string) string {
	maxWidth := 40

	spaces := ""
	for i := 0; i < maxWidth - text.Len(s); i++ {
		spaces += " "
	}

	return s + spaces + "\n"
}