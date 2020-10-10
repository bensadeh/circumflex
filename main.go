package main

import (
	"clx/cmd"
	subController "clx/submission-controller"
	"github.com/gdamore/tcell"
	"gitlab.com/tslocum/cview"
	"os"
	"os/exec"
)

func main() {
	cmd.Execute()
	clearScreen()

	sh := subController.NewSubmissionHandler()
	app := sh.Application
	setShortcuts(app)

	if err := app.SetRoot(sh.Pages, true).EnableMouse(false).Run(); err != nil {
		panic(err)
	}

}

func setShortcuts(app *cview.Application) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlN {
			//
		} else if event.Key() == tcell.KeyCtrlP {
			//
		} else if event.Rune() == 'q' {
			app.Stop()
		}
		return event
	})
}

func clearScreen() {
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	_ = c.Run()
}