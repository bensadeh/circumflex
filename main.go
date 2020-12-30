package main

import (
	"clx/cli"
	"clx/config"
	constructor "clx/constructors"
	"clx/controller"
)

func main() {
	config := config.GetConfig()

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
		sc.Settings,
		sc.Submissions,
		sc.MainView,
		sc.ApplicationState,
		config)

	cli.ClearScreen()

	sc.Application.SetRoot(sc.MainView.Grid, true)
	if err := sc.Application.Run(); err != nil {
		panic(err)
	}
}
