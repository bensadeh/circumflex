package clx

import (
	"clx/cli"
	constructor "clx/constructors"
	"clx/controller"
	"clx/core"
)

func Run(config *core.Config) {
	sc := constructor.NewScreenController(config.MarkAsRead)

	controller.SetAfterInitializationAndAfterResizeFunctions(sc.StoryHandler, sc.Application, sc.Articles, sc.MainView,
		sc.ApplicationState, config, sc.Service)
	controller.SetApplicationShortcuts(sc.StoryHandler, sc.VimRegister, sc.Application, sc.Articles, sc.MainView,
		sc.ApplicationState, config, sc.Service)

	cli.ClearScreen()
	sc.Application.SetRoot(sc.MainView.Grid, true)

	if err := sc.Application.Run(); err != nil {
		panic(err)
	}
}
