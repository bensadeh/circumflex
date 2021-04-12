package main

import (
	"clx/cli"
	"clx/config"
	"clx/controller"
	"os"

	constructor "clx/constructors"
)

func main() {
	useNativeTerminalSchemeColors()

	configuration := config.GetConfig()
	sc := constructor.NewScreenController()

	controller.SetAfterInitializationAndAfterResizeFunctions(sc.StoryHandler, sc.Application, sc.Articles, sc.MainView,
		sc.ApplicationState, configuration)
	controller.SetApplicationShortcuts(sc.StoryHandler, sc.VimRegister, sc.Application, sc.Articles, sc.MainView,
		sc.ApplicationState, configuration)

	cli.ClearScreen()
	sc.Application.SetRoot(sc.MainView.Grid, true)

	if err := sc.Application.Run(); err != nil {
		panic(err)
	}
}

func useNativeTerminalSchemeColors() {
	_ = os.Setenv("TCELL_TRUECOLOR", "disable")
}
