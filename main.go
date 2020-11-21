package main

import (
	"clx/cli"
	"clx/cmd"
	"clx/controller"
	builder "clx/initializers"
)

func main() {
	cmd.Execute()
	sc := builder.NewScreenController()
	controller.InitializeScreenController(sc)

	cli.Clear()

	sc.Application.SetRoot(sc.MainView.Grid, true)

	if err := sc.Application.Run(); err != nil {
		panic(err)
	}

}