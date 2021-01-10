package main

import (
	"clx/cli"
	"clx/config"
	constructor "clx/constructors"
	"clx/controller"
	"os"
)

func main() {
	_ = os.Setenv("TCELL_TRUECOLOR", "disable")

	configuration := config.GetConfig()

	sc := constructor.NewScreenController()
	controller.SetAfterInitializationAndAfterResizeFunctions(
		sc.Application,
		sc.Articles,
		sc.Submissions,
		sc.MainView,
		sc.ApplicationState,
		configuration)

	controller.SetApplicationShortcuts(
		sc.Application,
		sc.Articles,
		sc.Submissions,
		sc.MainView,
		sc.ApplicationState,
		configuration)

	cli.ClearScreen()

	sc.Application.SetRoot(sc.MainView.Grid, true)
	if err := sc.Application.Run(); err != nil {
		panic(err)
	}
}
