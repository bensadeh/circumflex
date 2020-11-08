package main

import (
	"clx/cli"
	"clx/cmd"
	subController "clx/controller"
)

func main() {
	cmd.Execute()

	sc := subController.NewScreenController()

	cli.Clear()

	sc.Application.SetRoot(sc.MainView.Grid, true)

	if err := sc.Application.Run(); err != nil {
		panic(err)
	}

}