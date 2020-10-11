package main

import (
	"clx/cli"
	"clx/cmd"
	subController "clx/submission-controller"
)

func main() {
	cmd.Execute()
	cli.Clear()

	sh := subController.NewSubmissionHandler()

	if err := sh.Application.SetRoot(sh.Pages, true).Run(); err != nil {
		panic(err)
	}

}