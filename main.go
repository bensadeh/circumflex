package main

import (
	"clx/cli"
	"clx/config"
	constructor "clx/constructors"
	"clx/controller"
)

func main() {
	configuration := config.GetConfig()

	sc := constructor.NewScreenController()
	controller.SetAfterInitializationAndAfterResizeFunctions(
		sc.Application,
		sc.Articles,
		sc.Submissions,
		sc.MainView,
		sc.ApplicationState)

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
