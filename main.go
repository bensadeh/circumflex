package main

import (
	"clx/cli"
	"clx/config"
	"clx/controller"
	"clx/favorites"
	"os"

	constructor "clx/constructors"
)

func main() {
	// Use the terminal's native color scheme
	_ = os.Setenv("TCELL_TRUECOLOR", "disable")

	configuration := config.GetConfig()

	fav := favorites.Initialize()

	sc := constructor.NewScreenController()
	controller.SetAfterInitializationAndAfterResizeFunctions(
		fav,
		sc.Application,
		sc.Articles,
		sc.Submissions,
		sc.MainView,
		sc.ApplicationState,
		configuration)

	controller.SetApplicationShortcuts(
		fav,
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
