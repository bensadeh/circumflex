package main

import (
	"clx/cmd"
	subController "clx/submission-controller"
	"os"
	"os/exec"
)

func main() {
	cmd.Execute()
	clearScreen()

	sh := subController.NewSubmissionHandler()

	if err := sh.Application.SetRoot(sh.Pages, true).Run(); err != nil {
		panic(err)
	}

}

func clearScreen() {
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	_ = c.Run()
}