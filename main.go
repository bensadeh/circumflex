package main

import (
	"clx/cli"
	"clx/cmd"
	subController "clx/controller"
)

func main() {
	cmd.Execute()
	cli.Clear()

	sc := subController.NewScreenController()

	if err := sc.Application.SetRoot(sc.MainView.Grid, true).Run(); err != nil {
		panic(err)
	}

}