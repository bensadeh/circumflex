package model

import (
	"clx/browser"
	"clx/cli"
	"clx/comment"
	"clx/constants/categories"
	"clx/constants/messages"
	"clx/constants/state"
	"clx/core"
	"clx/handler"
	"clx/hn"
	"clx/info"
	"clx/item"
	"clx/markdown/parser"
	"clx/markdown/postprocessor"
	"clx/markdown/renderer"
	"clx/reader"
	"clx/screen"
	"clx/utils/message"
	"clx/utils/ranking"
	"clx/utils/vim"
	"clx/validator"
	"clx/view"
	"strconv"
	"time"

	text "github.com/MichaelMure/go-term-text"

	"code.rocketnine.space/tslocum/cview"
	"github.com/gdamore/tcell/v2"
)

func SetAfterInitializationAndAfterResizeFunctions(app *cview.Application, list *cview.List,
	main *core.MainView, appState *core.ApplicationState, config *core.Config,
	ret *handler.StoryHandler, service hn.Service) {
	app.SetAfterResizeFunc(func(_ int, _ int) {
		app.SetRoot(main.Grid, true)

		resetStates(appState, ret, service)
		initializeView(appState, main, ret, config)

		listItems, err := ret.GetStories(appState.CurrentCategory, appState.CurrentPage,
			appState.StoriesToShow, config.HighlightHeadlines, service)
		if err != nil {
			setToErrorState(appState, main, list, app)

			return
		}

		marginText := ranking.GetRankings(config.RelativeNumbering, appState.StoriesToShow, len(listItems), 0, 0)

		view.ShowItems(list, listItems)
		view.SetLeftMarginText(main, marginText)
		view.ClearStatusBar(main)

		if appState.State == state.OnHelpScreen {
			updateInfoScreenView(main, appState)
		}
	})
}

func setToErrorState(appState *core.ApplicationState, main *core.MainView, list *cview.List, app *cview.Application) {
	errorMessage := message.Error(messages.OfflineMessage)
	appState.IsOffline = true

	view.SetPermanentStatusBar(main, errorMessage, cview.AlignCenter)
	view.ClearList(list)
	app.Draw()
}

func resetStates(appState *core.ApplicationState, ret *handler.StoryHandler, service hn.Service) {
	resetApplicationState(appState)
	ret.Reset()
	service.Init(appState.StoriesToShow)
}

func resetApplicationState(appState *core.ApplicationState) {
	appState.CurrentPage = 0
	appState.ScreenWidth = screen.GetTerminalWidth()
	appState.ScreenHeight = screen.GetTerminalHeight()
	appState.StoriesToShow = screen.GetSubmissionsToShow(appState.ScreenHeight, 30)
	appState.IsOnAddFavoriteConfirmationMessage = false
	appState.IsOffline = false
}

func initializeView(appState *core.ApplicationState, main *core.MainView, ret *handler.StoryHandler,
	config *core.Config) {
	header := ret.GetHackerNewsHeader(appState.CurrentCategory, config.HeaderType)

	view.SetPanelToMainView(main)
	view.SetHackerNewsHeader(main, header)
	view.SetPageCounter(main,
		appState.CurrentPage,
		ret.GetMaxPages(appState.CurrentCategory, appState.StoriesToShow))
}

func Refresh(app *cview.Application, main *core.MainView, appState *core.ApplicationState) {
	afterResizeFunc := app.GetAfterResizeFunc()
	afterResizeFunc(appState.ScreenWidth, appState.ScreenHeight)

	if !appState.IsOffline {
		view.SetTemporaryStatusBar(app, main, messages.Refreshed, time.Second*2)
	}
}

func ReadSubmissionComments(app *cview.Application, main *core.MainView, list *cview.List,
	appState *core.ApplicationState, config *core.Config, r *handler.StoryHandler, reg *vim.Register, service hn.Service) {
	story := r.GetStoryAndMarkAsRead(appState.CurrentCategory, list.GetCurrentItemIndex(), appState.StoriesToShow,
		appState.CurrentPage)

	app.Suspend(func() {
		// id := strconv.Itoa(story.ID)

		comments := service.FetchStory(story.ID)
		//comments, err := comment.FetchComments(id)
		//if err != nil {
		//	errorMessage := message.Error(messages.CommentsNotFetched)
		//	view.SetTemporaryStatusBar(app, main, errorMessage, 4*time.Second)
		//
		//	return
		//}

		r.UpdateFavoriteStoryAndWriteToDisk(comments)
		screenWidth := screen.GetTerminalWidth()
		commentTree := comment.ToString(comments, config, screenWidth)

		cli.Less(commentTree)
	})

	changePage(app, list, main, appState, config, r, reg, 0, service)
}

func ReadContentInReaderViewDebug(app *cview.Application, main *core.MainView, list *cview.List,
	appState *core.ApplicationState, config *core.Config, r *handler.StoryHandler, reg *vim.Register, service hn.Service) {
	story := r.GetStory(appState.CurrentCategory, list.GetCurrentItemIndex(), appState.StoriesToShow,
		appState.CurrentPage)

	enterReaderModeDebug(app, main, list, appState, config, r, reg, story, service)
}

func ReadContentInReaderView(app *cview.Application, main *core.MainView, list *cview.List,
	appState *core.ApplicationState, config *core.Config, r *handler.StoryHandler, reg *vim.Register, service hn.Service) {
	story := r.GetStory(appState.CurrentCategory, list.GetCurrentItemIndex(), appState.StoriesToShow,
		appState.CurrentPage)
	errorMessage := validator.GetErrorMessage(story.Title, story.Domain)

	if errorMessage != "" {
		view.SetPermanentStatusBar(main, message.Warning(errorMessage), cview.AlignCenter)

		return
	}

	r.MarkAsRead(appState.CurrentCategory, list.GetCurrentItemIndex(), appState.StoriesToShow, appState.CurrentPage)
	enterReaderModeBuiltInParser(app, main, list, appState, config, r, reg, story, service)
}

func enterReaderModeDebug(app *cview.Application, main *core.MainView, list *cview.List,
	appState *core.ApplicationState, config *core.Config, r *handler.StoryHandler, reg *vim.Register,
	story *item.Item, service hn.Service) {
	fetchTimeout := false

	app.Suspend(func() {
		url := story.URL

		article, err := reader.GetNew(url)
		if err != nil {
			fetchTimeout = true

			return
		}

		article, _ = text.Wrap(article, 80)

		cli.Less(article)
	})

	if fetchTimeout {
		view.SetPermanentStatusBar(main, message.Error(messages.ArticleNotFetched), cview.AlignCenter)

		return
	}

	changePage(app, list, main, appState, config, r, reg, 0, service)
}

func enterReaderModeBuiltInParser(app *cview.Application, main *core.MainView, list *cview.List,
	appState *core.ApplicationState, config *core.Config, r *handler.StoryHandler, reg *vim.Register,
	story *item.Item, service hn.Service) {
	fetchTimeout := false

	app.Suspend(func() {
		url := story.URL

		article, err := reader.GetNew(url)
		if err != nil {
			fetchTimeout = true

			return
		}

		blocks := parser.Parse(article)
		header := renderer.CreateHeader(story.Title, story.URL, 70)
		renderedArticle := renderer.ToString(blocks, 70, config.IndentationSymbol)
		renderedArticle = postprocessor.Process(header+renderedArticle, url)

		cli.Less(renderedArticle)
	})

	if fetchTimeout {
		view.SetPermanentStatusBar(main, message.Error(messages.ArticleNotFetched), cview.AlignCenter)

		return
	}

	changePage(app, list, main, appState, config, r, reg, 0, service)
}

func OpenCommentsInBrowser(list *cview.List, appState *core.ApplicationState, r *handler.StoryHandler) {
	story := r.GetStory(appState.CurrentCategory, list.GetCurrentItemIndex(), appState.StoriesToShow,
		appState.CurrentPage)
	url := "https://news.ycombinator.com/item?id=" + strconv.Itoa(story.ID)
	browser.Open(url)
}

func OpenLinkInBrowser(list *cview.List, appState *core.ApplicationState, r *handler.StoryHandler) {
	story := r.GetStory(appState.CurrentCategory, list.GetCurrentItemIndex(), appState.StoriesToShow,
		appState.CurrentPage)
	browser.Open(story.URL)
}

func NextPage(app *cview.Application, list *cview.List, main *core.MainView, appState *core.ApplicationState,
	config *core.Config, ret *handler.StoryHandler, reg *vim.Register, service hn.Service) {
	isOnLastPage := appState.CurrentPage+1 > ret.GetMaxPages(appState.CurrentCategory, appState.StoriesToShow)
	if isOnLastPage {
		return
	}

	changePage(app, list, main, appState, config, ret, reg, 1, service)
}

func PreviousPage(app *cview.Application, list *cview.List, main *core.MainView, appState *core.ApplicationState,
	config *core.Config, ret *handler.StoryHandler, reg *vim.Register, service hn.Service) {
	isOnFirstPage := appState.CurrentPage-1 < 0
	if isOnFirstPage {
		return
	}

	changePage(app, list, main, appState, config, ret, reg, -1, service)
}

func changePage(app *cview.Application, list *cview.List, main *core.MainView, appState *core.ApplicationState,
	config *core.Config, ret *handler.StoryHandler, reg *vim.Register, delta int, service hn.Service) {
	currentlySelectedItem := list.GetCurrentItemIndex()
	appState.CurrentPage += delta

	listItems, err := ret.GetStories(appState.CurrentCategory, appState.CurrentPage,
		appState.StoriesToShow, config.HighlightHeadlines, service)
	if err != nil {
		setToErrorState(appState, main, list, app)

		return
	}

	view.ShowItems(list, listItems)
	view.SelectItem(list, currentlySelectedItem)

	ClearVimRegister(main, reg)

	marginText := ranking.GetRankings(config.RelativeNumbering, appState.StoriesToShow, len(listItems),
		list.GetCurrentItemIndex(), appState.CurrentPage)
	header := ret.GetHackerNewsHeader(appState.CurrentCategory, config.HeaderType)
	maxPages := ret.GetMaxPages(appState.CurrentCategory, appState.StoriesToShow)

	view.SetLeftMarginText(main, marginText)
	view.SetHackerNewsHeader(main, header)
	view.SetPageCounter(main, appState.CurrentPage, maxPages)
}

func ChangeCategory(app *cview.Application, event *tcell.EventKey, list *cview.List, appState *core.ApplicationState,
	main *core.MainView, config *core.Config, ret *handler.StoryHandler, reg *vim.Register, service hn.Service) {
	currentItem := list.GetCurrentItemIndex()
	appState.CurrentCategory = ret.GetNewCategory(event, appState.CurrentCategory)
	appState.CurrentPage = 0

	listItems, err := ret.GetStories(appState.CurrentCategory, appState.CurrentPage,
		appState.StoriesToShow, config.HighlightHeadlines, service)
	if err != nil {
		setToErrorState(appState, main, list, app)

		return
	}

	view.ShowItems(list, listItems)
	view.SelectItem(list, currentItem)
	ClearVimRegister(main, reg)

	header := ret.GetHackerNewsHeader(appState.CurrentCategory, config.HeaderType)
	marginText := ranking.GetRankings(config.RelativeNumbering, appState.StoriesToShow, len(listItems),
		list.GetCurrentItemIndex(), appState.CurrentPage)
	maxPages := ret.GetMaxPages(appState.CurrentCategory, appState.StoriesToShow)

	view.SetLeftMarginText(main, marginText)
	view.SetPageCounter(main, appState.CurrentPage, maxPages)
	view.SetHackerNewsHeader(main, header)
}

func ScrollSettingsOneLineUp(main *core.MainView) {
	view.ScrollInfoScreenByAmount(main, -1)
}

func ScrollSettingsOneLineDown(main *core.MainView) {
	view.ScrollInfoScreenByAmount(main, 1)
}

func ScrollSettingsOneHalfPageUp(main *core.MainView) {
	halfPage := screen.GetTerminalHeight() / 2
	view.ScrollInfoScreenByAmount(main, -halfPage)
}

func ScrollSettingsOneHalfPageDown(main *core.MainView) {
	halfPage := screen.GetTerminalHeight() / 2
	view.ScrollInfoScreenByAmount(main, halfPage)
}

func ScrollSettingsToBeginning(main *core.MainView) {
	view.ScrollInfoScreenToBeginning(main)
}

func ScrollSettingsToEnd(main *core.MainView) {
	view.ScrollInfoScreenToEnd(main)
}

func CancelConfirmation(appState *core.ApplicationState, main *core.MainView) {
	appState.IsOnAddFavoriteConfirmationMessage = false
	appState.IsOnDeleteFavoriteConfirmationMessage = false

	view.SetPermanentStatusBar(main, messages.Cancelled, cview.AlignCenter)
}

func SelectItemUp(main *core.MainView, list *cview.List, appState *core.ApplicationState, config *core.Config,
	reg *vim.Register) {
	currentItem := list.GetCurrentItemIndex()
	nextItem := reg.GetItemUp(currentItem)

	selectItem(main, list, appState, config, reg, nextItem)
}

func SelectItemDown(main *core.MainView, list *cview.List, appState *core.ApplicationState, config *core.Config,
	reg *vim.Register) {
	currentItem := list.GetCurrentItemIndex()
	itemCount := list.GetItemCount()
	nextItem := reg.GetItemDown(currentItem, itemCount)

	selectItem(main, list, appState, config, reg, nextItem)
}

func selectItem(main *core.MainView, list *cview.List, appState *core.ApplicationState, config *core.Config,
	reg *vim.Register, item int) {
	ClearVimRegister(main, reg)

	marginText := ranking.GetRankings(config.RelativeNumbering, appState.StoriesToShow, list.GetItemCount(),
		item, appState.CurrentPage)

	view.SelectItem(list, item)
	view.SetLeftMarginText(main, marginText)
	view.ClearStatusBar(main)
}

func EnterInfoScreen(main *core.MainView, appState *core.ApplicationState, reg *vim.Register) {
	appState.State = state.OnHelpScreen

	ClearVimRegister(main, reg)
	updateInfoScreenView(main, appState)
}

func updateInfoScreenView(main *core.MainView, appState *core.ApplicationState) {
	statusBarText := info.GetStatusBarText()
	infoScreenText := info.GetText(appState.ScreenWidth)

	view.SetInfoScreenText(main, infoScreenText)
	view.SetPermanentStatusBar(main, statusBarText, cview.AlignCenter)
	view.HidePageCounter(main)
	view.ClearHelpScreenHeader(main)
	view.HideLeftMarginRanks(main)
	view.SetPanelToInfoView(main)
}

func ExitInfoScreen(main *core.MainView, appState *core.ApplicationState, config *core.Config, list *cview.List,
	ret *handler.StoryHandler) {
	appState.State = state.OnSubmissionPage

	marginText := ranking.GetRankings(config.RelativeNumbering, appState.StoriesToShow, list.GetItemCount(),
		list.GetCurrentItemIndex(), appState.CurrentPage)
	header := ret.GetHackerNewsHeader(appState.CurrentCategory, config.HeaderType)
	maxPages := ret.GetMaxPages(appState.CurrentCategory, appState.StoriesToShow)

	view.SetLeftMarginText(main, marginText)
	view.SetHackerNewsHeader(main, header)
	view.SetPanelToMainView(main)
	view.SetPageCounter(main, appState.CurrentPage, maxPages)
	view.ClearStatusBar(main)
}

func LowerCaseG(main *core.MainView, appState *core.ApplicationState, list *cview.List, config *core.Config,
	reg *vim.Register) {
	currentItem := list.GetCurrentItemIndex()
	itemToJumpTo := reg.LowerCaseG(currentItem, appState.StoriesToShow, appState.CurrentPage)
	register := reg.Print()
	marginText := ranking.GetRankings(config.RelativeNumbering, appState.StoriesToShow, list.GetItemCount(),
		itemToJumpTo, appState.CurrentPage)

	view.SetLeftMarginText(main, marginText)
	view.SelectItem(list, itemToJumpTo)
	view.SetPermanentStatusBar(main, register, cview.AlignRight)
}

func UpperCaseG(main *core.MainView, appState *core.ApplicationState, list *cview.List, config *core.Config,
	reg *vim.Register, ret *handler.StoryHandler, service hn.Service) {
	stories, _ := ret.GetStories(appState.CurrentCategory, appState.CurrentPage,
		appState.StoriesToShow, config.HighlightHeadlines, service)
	storiesToShow := len(stories)
	currentItem := list.GetCurrentItemIndex()
	itemToJumpTo := reg.UpperCaseG(currentItem, storiesToShow, appState.CurrentPage)
	register := reg.Print()
	marginText := ranking.GetRankings(config.RelativeNumbering, storiesToShow, list.GetItemCount(),
		itemToJumpTo, appState.CurrentPage)

	view.SetLeftMarginText(main, marginText)
	view.SelectItem(list, itemToJumpTo)
	view.SetPermanentStatusBar(main, register, cview.AlignRight)
}

func PutDigitInRegister(main *core.MainView, number rune, reg *vim.Register) {
	reg.PutInRegister(number)

	view.SetPermanentStatusBar(main, reg.Print(), cview.AlignRight)
}

func Quit(app *cview.Application) {
	app.Stop()
}

func ClearVimRegister(main *core.MainView, reg *vim.Register) {
	reg.Clear()

	view.ClearStatusBar(main)
}

func AddToFavoritesConfirmationDialogue(main *core.MainView, appState *core.ApplicationState) {
	appState.IsOnAddFavoriteConfirmationMessage = true

	view.SetPermanentStatusBar(main, messages.AddToFavorites, cview.AlignCenter)
}

func DeleteFavoriteConfirmationDialogue(main *core.MainView, appState *core.ApplicationState) {
	appState.IsOnDeleteFavoriteConfirmationMessage = true

	view.SetPermanentStatusBar(main, messages.DeleteFromFavorites, cview.AlignCenter)
}

func AddToFavorites(app *cview.Application, list *cview.List, main *core.MainView, appState *core.ApplicationState,
	config *core.Config, ret *handler.StoryHandler, reg *vim.Register, service hn.Service) {
	statusBarMessage := ""
	appState.IsOnAddFavoriteConfirmationMessage = false
	story := ret.GetStory(appState.CurrentCategory, list.GetCurrentItemIndex(), appState.StoriesToShow,
		appState.CurrentPage)

	err := ret.AddItemToFavoritesAndWriteToFile(story)
	if err != nil {
		statusBarMessage = message.Error(messages.FavoriteNotAdded)
	} else {
		statusBarMessage = message.Success(messages.FavoriteAdded)
	}

	changePage(app, list, main, appState, config, ret, reg, 0, service)
	view.SetPermanentStatusBar(main, statusBarMessage, cview.AlignCenter)
}

func DeleteItem(app *cview.Application, list *cview.List, appState *core.ApplicationState,
	main *core.MainView, config *core.Config, ret *handler.StoryHandler, reg *vim.Register, service hn.Service) {
	appState.IsOnDeleteFavoriteConfirmationMessage = false

	ret.DeleteStoryAndWriteToFile(appState.CurrentCategory, list.GetCurrentItemIndex(), appState.StoriesToShow,
		appState.CurrentPage)

	hasDeletedLastItemOnSecondOrThirdPage := list.GetCurrentItemIndex() == 0 &&
		list.GetItemCount() == 1 && appState.CurrentPage != 0
	hasDeletedLastItemOnFirstPage := list.GetCurrentItemIndex() == 0 &&
		list.GetItemCount() == 1 && appState.CurrentPage == 0

	switch {
	case hasDeletedLastItemOnSecondOrThirdPage:
		changePage(app, list, main, appState, config, ret, reg, -1, service)

	case hasDeletedLastItemOnFirstPage:
		appState.CurrentCategory = categories.Show
		keyTab := tcell.NewEventKey(tcell.KeyTab, ' ', tcell.ModNone)
		ChangeCategory(app, keyTab, list, appState, main, config, ret, reg, service)

	default:
		changePage(app, list, main, appState, config, ret, reg, 0, service)
	}

	m := message.Success(messages.ItemDeleted)
	view.SetPermanentStatusBar(main, m, cview.AlignCenter)
}
