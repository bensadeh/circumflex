package main

import (
	"clx/cli"
	"clx/cmd"
	subController "clx/submission-controller"
)

func main() {
	cmd.Execute()
	cli.Clear()

	sc := subController.NewScreenController()

	if err := sc.Application.SetRoot(sc.Grid, true).Run(); err != nil {
		panic(err)
	}

}