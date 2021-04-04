package main

import (
	"clx/cli"
	"clx/config"
	"clx/controller"
	"clx/favorites"
	"clx/retriever"
	"clx/utils/vim"
	"os"

	constructor "clx/constructors"
)

func main() {
	// Use the terminal's native color scheme
	_ = os.Setenv("TCELL_TRUECOLOR", "disable")

	configuration := config.GetConfig()

	fav := favorites.Initialize()
	ret := new(retriever.Retriever)
	ret.Init(fav)

	sc := constructor.NewScreenController()
	controller.SetAfterInitializationAndAfterResizeFunctions(
		ret,
		sc.Application,
		sc.Articles,
		sc.MainView,
		sc.ApplicationState,
		configuration)

	controller.SetApplicationShortcuts(
		ret,
		new(vim.Register),
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
