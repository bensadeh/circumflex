package main

import (
	"clx/cli"
	"clx/constructors"
	"clx/controller"
)

func main() {
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
		sc.ApplicationState)

	cli.ClearScreen()

	sc.Application.SetRoot(sc.MainView.Grid, true)
	if err := sc.Application.Run(); err != nil {
		panic(err)
	}
}
