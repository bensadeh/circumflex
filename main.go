package main

import (
	"clx/cli"
	"clx/config"
	"clx/controller"
	"clx/favorites"
	"clx/retriever"
	"os"

	constructor "clx/constructors"
)

func main() {
	// Use the terminal's native color scheme
	_ = os.Setenv("TCELL_TRUECOLOR", "disable")

	configuration := config.GetConfig()

	fav := favorites.Initialize()
	ret := new(retriever.Retriever)

	sc := constructor.NewScreenController()
	controller.SetAfterInitializationAndAfterResizeFunctions(
		fav,
		ret,
		sc.Application,
		sc.Articles,
		sc.MainView,
		sc.ApplicationState,
		configuration)

	controller.SetApplicationShortcuts(
		fav,
		ret,
		sc.Application,
		sc.Articles,
		sc.MainView,
		sc.ApplicationState,
		configuration)

	cli.ClearScreen()

	sc.Application.SetRoot(sc.MainView.Grid, true)

	if err := sc.Application.Run(); err != nil {
		panic(err)
	}
}
