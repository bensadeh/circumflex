package list

import (
	"clx/categories"
	_ "embed"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"clx/app"

	"github.com/charmbracelet/bubbles/viewport"

	"clx/reader"

	"clx/browser"
	"clx/bubble/list/message"
	"clx/bubble/ranking"
	"clx/cli"
	"clx/constants/category"
	"clx/download"
	"clx/favorites"
	"clx/header"
	"clx/help"
	"clx/history"
	"clx/hn"
	"clx/hn/services/hybrid"
	"clx/hn/services/mock"
	"clx/item"
	"clx/screen"
	"clx/settings"
	"clx/tree"
	"clx/validator"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	numberOfCategories = 6
)

// Item is an item that appears in the list.
// type Item interface{}

// ItemDelegate encapsulates the general functionality for all list items. The
// benefit to separating this logic from the item itself is that you can change
// the functionality of items without changing the actual items themselves.
//
// Note that if the delegate also implements help.KeyMap delegate-related
// help items will be added to the help view.
type ItemDelegate interface {
	// Render renders the item's view.
	Render(w io.Writer, m *Model, index int, item *item.Item)

	// Height is the height of the list item.
	Height() int

	// Spacing is the size of the horizontal gap between list items in cells.
	Spacing() int

	// Update is the update loop for items. All messages in the list's update
	// loop will pass through here except when the user is setting a filter.
	// Use this method to perform item-level updates appropriate to this
	// delegate.
	Update(msg tea.Msg, m *Model) tea.Cmd
}

// Model contains the state of this component.
type Model struct {
	showTitle     bool
	showStatusBar bool
	disableInput  bool

	Title  string
	Styles Styles

	spinner                     spinner.Model
	showSpinner                 bool
	width                       int
	height                      int
	Paginator                   paginator.Model
	cursor                      int
	onStartup                   bool
	isVisible                   bool
	onAddToFavoritesPrompt      bool
	onRemoveFromFavoritesPrompt bool

	StatusMessageLifetime time.Duration

	statusMessage      string
	statusMessageTimer *time.Timer

	isBufferActive bool
	items          [][]*item.Item

	delegate  ItemDelegate
	history   history.History
	config    *settings.Config
	service   hn.Service
	favorites *favorites.Favorites
	cat       *categories.Categories

	isOnHelpScreen bool
	viewport       viewport.Model
}

func (m *Model) FetchStoriesForFirstCategory() tea.Cmd {
	return func() tea.Msg {
		itemsToFetch := m.getNumberOfItemsToFetch(m.cat.GetCurrentCategory(m.favorites.HasItems()))
		categoryToFetch := m.cat.GetCurrentCategory(m.favorites.HasItems())
		stories, errMsg := m.service.FetchItems(itemsToFetch, categoryToFetch)

		m.items[categoryToFetch] = stories

		return message.FetchingFinished{Message: errMsg}
	}
}

func (m *Model) getNumberOfItemsToFetch(cat int) int {
	switch cat {
	case category.Top:
		return m.Paginator.PerPage * 3

	case category.New:
		return m.Paginator.PerPage * 3

	case category.Best:
		return m.Paginator.PerPage * 3

	case category.Ask:
		return m.Paginator.PerPage

	case category.Show:
		return m.Paginator.PerPage

	default:
		return m.Paginator.PerPage
	}
}

func New(delegate ItemDelegate, config *settings.Config, cat *categories.Categories, favorites *favorites.Favorites, width, height int) *Model {
	styles := DefaultStyles()

	sp := spinner.New()
	sp.Spinner = getSpinner()
	sp.Style = styles.Spinner

	p := paginator.New()
	p.Type = paginator.Dots
	p.ActiveDot = styles.ActivePaginationDot.String()
	p.InactiveDot = styles.InactivePaginationDot.String()
	p.UseHLKeys = false
	p.UseJKKeys = false
	p.UseLeftRightKeys = false
	p.UsePgUpPgDownKeys = false
	p.UseUpDownKeys = false

	bufferCategory := 1
	items := make([][]*item.Item, numberOfCategories+bufferCategory)

	m := Model{
		showTitle:             true,
		showStatusBar:         true,
		Styles:                styles,
		Title:                 "List",
		StatusMessageLifetime: time.Second,

		width:        width,
		height:       height,
		delegate:     delegate,
		history:      getHistory(config.DebugMode, config.DoNotMarkSubmissionsAsRead),
		items:        items,
		Paginator:    p,
		spinner:      sp,
		onStartup:    true,
		isVisible:    true,
		disableInput: true,
		config:       config,
		service:      getService(config.DebugMode),
		favorites:    favorites,
		cat:          cat,
	}

	m.updatePagination()

	return &m
}

func getHistory(debugMode bool, doNotMarkAsRead bool) history.History {
	if debugMode {
		return history.NewMockHistory()
	}

	if doNotMarkAsRead {
		return history.NewNonPersistentHistory()
	}

	return history.NewPersistentHistory()
}

func getService(debugMode bool) hn.Service {
	if debugMode {
		return mock.Service{}
	}

	return &hybrid.Service{}
}

// SetShowTitle shows or hides the title bar.
func (m *Model) SetShowTitle(v bool) {
	m.showTitle = v
	m.updatePagination()
}

func (m *Model) SetIsVisible(v bool) {
	m.isVisible = v
}

// SetShowStatusBar shows or hides the view that displays metadata about the
// list, such as item counts.
func (m *Model) SetShowStatusBar(v bool) {
	m.showStatusBar = v
	m.updatePagination()
}

// ShowStatusBar returns whether or not the status bar is set to be rendered.
func (m *Model) ShowStatusBar() bool {
	return m.showStatusBar
}

// Set the items available in the list. This returns a command.
func (m *Model) SetItems(i []*item.Item) tea.Cmd {
	var cmd tea.Cmd
	m.items[m.cat.GetCurrentCategory(m.favorites.HasItems())] = i

	m.updatePagination()

	return cmd
}

// Select selects the given index of the list and goes to its respective page.
func (m *Model) Select(index int) {
	m.Paginator.Page = index / m.Paginator.PerPage
	m.cursor = index % m.Paginator.PerPage
}

// VisibleItems returns the total items available to be shown.
func (m *Model) VisibleItems() []*item.Item {
	if m.isBufferActive {
		return m.items[category.Buffer]
	}

	return m.items[m.cat.GetCurrentCategory(m.favorites.HasItems())]
}

// SelectedItems returns the current selected item in the list.
func (m *Model) SelectedItem() *item.Item {
	i := m.Index()

	items := m.VisibleItems()
	if i < 0 || len(items) == 0 || len(items) <= i {
		// return nil
		return &item.Item{}
	}

	return items[i]
}

// Index returns the index of the currently selected item as it appears in the
// entire slice of items.
func (m *Model) Index() int {
	return m.Paginator.Page*m.Paginator.PerPage + m.cursor
}

// Cursor returns the index of the cursor on the current page.
func (m *Model) Cursor() int {
	return m.cursor
}

// CursorUp moves the cursor up. This can also move the state to the previous
// page.
func (m *Model) CursorUp() {
	m.cursor--

	// If we're at the top, stop
	if m.cursor < 0 {
		m.cursor = 0
		return
	}

	return
}

// CursorDown moves the cursor down. This can also advance the state to the
// next page.
func (m *Model) CursorDown() {
	itemsOnPage := m.Paginator.ItemsOnPage(len(m.VisibleItems()))

	m.cursor++

	// If we're at the end, stop
	if m.cursor < itemsOnPage {
		return
	}

	m.cursor = itemsOnPage - 1
}

func (m *Model) ToggleSpinner() tea.Cmd {
	if !m.showSpinner {
		return m.StartSpinner()
	}
	m.StopSpinner()
	return nil
}

func (m *Model) StartSpinner() tea.Cmd {
	// Hack: I can't get the spinner to reset properly. As a workaround, we
	// instantiate a new spinner each time we want to show it.
	m.spinner = spinner.New()
	m.spinner.Spinner = getSpinner()
	m.spinner.Style = DefaultStyles().Spinner

	m.showSpinner = true
	return m.spinner.Tick
}

func (m *Model) StopSpinner() {
	m.showSpinner = false
}

func (m *Model) NewStatusMessage(s string) tea.Cmd {
	m.statusMessage = s
	if m.statusMessageTimer != nil {
		m.statusMessageTimer.Stop()
	}

	m.statusMessageTimer = time.NewTimer(m.StatusMessageLifetime)

	// Wait for timeout
	return func() tea.Msg {
		<-m.statusMessageTimer.C
		return message.StatusMessageTimeout{}
	}
}

func (m *Model) NewStatusMessageWithDuration(s string, d time.Duration) tea.Cmd {
	m.statusMessage = lipgloss.NewStyle().Render(s)

	if m.statusMessageTimer != nil {
		m.statusMessageTimer.Stop()
	}

	m.statusMessageTimer = time.NewTimer(d)

	// Wait for timeout
	return func() tea.Msg {
		<-m.statusMessageTimer.C
		return message.StatusMessageTimeout{}
	}
}

func (m *Model) SetPermanentStatusMessage(s string, faint bool) {
	m.statusMessage = lipgloss.NewStyle().
		Faint(faint).
		Render(s)
}

// SetSize sets the width and height of this component.
func (m *Model) SetSize(width, height int) {
	m.setSize(width, height)
}

func (m *Model) setSize(width, height int) {
	m.width = width
	m.height = height
	m.updatePagination()
}

// Update pagination according to the amount of items for the current state.
func (m *Model) updatePagination() {
	index := m.Index()
	availHeight := m.height

	if m.showTitle {
		availHeight -= lipgloss.Height(m.titleView())
	}
	if m.showStatusBar {
		// We subtract one from the height because we don't want any spacing
		availHeight -= lipgloss.Height(m.statusAndPaginationView()) - 1
	}

	m.Paginator.PerPage = max(1, availHeight/(m.delegate.Height()+m.delegate.Spacing()))

	if pages := len(m.VisibleItems()); pages < 1 {
		m.Paginator.SetTotalPages(1)
	} else {
		m.Paginator.SetTotalPages(pages)
	}

	// Restore index
	m.Paginator.Page = index / m.Paginator.PerPage
	m.cursor = index % m.Paginator.PerPage

	// Make sure the page stays in bounds
	if m.Paginator.Page >= m.Paginator.TotalPages-1 {
		m.Paginator.Page = max(0, m.Paginator.TotalPages-1)
	}
}

func (m *Model) hideStatusMessage() {
	m.statusMessage = ""
	if m.statusMessageTimer != nil {
		m.statusMessageTimer.Stop()
	}
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	windowSizeMsg, isWindowSizeMsg := msg.(tea.WindowSizeMsg)

	// Since this program is using the full size of the viewport we
	// need to wait until we've received the window dimensions before
	// we can initialize the viewport. The initial dimensions come in
	// quickly, though asynchronously, which is why we wait for them
	// here.
	if m.onStartup && !isWindowSizeMsg {
		return m, nil
	}

	if m.onStartup && isWindowSizeMsg {
		h, v := lipgloss.NewStyle().GetFrameSize()
		m.SetSize(windowSizeMsg.Width-h, windowSizeMsg.Height-v)

		var cmds []tea.Cmd

		spinnerCmd := m.StartSpinner()
		cmds = append(cmds, spinnerCmd)

		m.SetOnStartup(false)

		m.items[category.Favorites] = m.favorites.GetItems()

		fetchCmd := m.FetchStoriesForFirstCategory()
		cmds = append(cmds, fetchCmd)

		heightOfHeaderAndStatusLine := 4

		m.viewport = viewport.New(windowSizeMsg.Width, windowSizeMsg.Height-heightOfHeaderAndStatusLine)
		m.viewport.YPosition = 2
		m.viewport.HighPerformanceRendering = false

		content := lipgloss.NewStyle().
			Width(windowSizeMsg.Width).
			AlignHorizontal(lipgloss.Center).
			SetString(help.GetHelpScreen(m.config.EnableNerdFonts))

		m.viewport.SetContent(content.String())

		return m, tea.Batch(cmds...)
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case spinner.TickMsg:
		newSpinnerModel, cmd := m.spinner.Update(msg)
		m.spinner = newSpinnerModel
		if m.showSpinner {
			cmds = append(cmds, cmd)
		}

	case message.FetchingFinished:
		m.StopSpinner()
		h, v := lipgloss.NewStyle().GetFrameSize()
		m.setSize(screen.GetTerminalWidth()-h, screen.GetTerminalHeight()-v)
		m.disableInput = false
		m.NewStatusMessage(msg.Message)

		return m, nil

	case message.StatusMessageTimeout:
		m.hideStatusMessage()

	case message.AddToFavorites:
		m.favorites.Add(msg.Item)
		m.items[category.Favorites] = m.favorites.GetItems()

		m.favorites.Write()

		m.updatePagination()

	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().GetFrameSize()
		m.SetSize(msg.Width-h, msg.Height-v)

		headerHeight := 2
		footerHeight := 2
		verticalMarginHeight := headerHeight + footerHeight

		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - verticalMarginHeight

		m.width = msg.Width
		m.height = msg.Height

		content := lipgloss.NewStyle().
			Width(msg.Width).
			AlignHorizontal(lipgloss.Center).
			SetString(help.GetHelpScreen(m.config.EnableNerdFonts))

		m.viewport.SetContent(content.String())

		return m, nil

	case message.EnteringCommentSection:
		lastVisited := m.history.GetLastVisited(msg.Id)

		m.history.MarkAsReadAndWriteToDisk(msg.Id, msg.CommentCount)

		story := m.service.FetchComments(msg.Id)

		if m.cat.GetCurrentCategory(m.favorites.HasItems()) == category.Favorites {
			m.favorites.UpdateStoryAndWriteToDisk(story)
		}

		commentTree := tree.Print(story, m.config, m.width, lastVisited)

		command := cli.Less(commentTree, m.config)

		return m, tea.ExecProcess(command, func(err error) tea.Msg {
			return message.EditorFinishedMsg{Err: err}
		})

	case message.OpeningLink:
		m.history.MarkAsReadAndWriteToDisk(msg.Id, msg.CommentCount)

	case message.OpeningCommentsInBrowser:
		m.history.MarkAsReadAndWriteToDisk(msg.Id, msg.CommentCount)

	case message.EnteringReaderMode:
		errorMessage := validator.GetErrorMessage(msg.Title, msg.Domain)
		if errorMessage != "" {
			cmds = append(cmds, m.NewStatusMessageWithDuration(errorMessage, time.Second*3))
			cmds = append(cmds, func() tea.Msg {
				return message.EditorFinishedMsg{Err: nil}
			})
			return m, tea.Batch(cmds...)
		}

		article, err := reader.GetArticle(msg.Url, msg.Title, m.config.CommentWidth, m.config.IndentationSymbol)
		if err != nil {
			cmds = append(cmds, m.NewStatusMessageWithDuration("Could not read article in Reader Mode", time.Second*3))
			cmds = append(cmds, func() tea.Msg {
				return message.EditorFinishedMsg{Err: nil}
			})
			return m, tea.Batch(cmds...)
		}

		command := cli.Less(article, m.config)

		m.history.MarkAsReadAndWriteToDisk(msg.Id, msg.CommentCount)

		return m, tea.ExecProcess(command, func(err error) tea.Msg {
			return message.EditorFinishedMsg{Err: err}
		})

	case message.EditorFinishedMsg:
		m.SetIsVisible(true)
		m.SetDisabledInput(false)

	case message.FetchAndChangeToCategory:
		return m, func() tea.Msg {
			itemsToFetch := m.getNumberOfItemsToFetch(msg.Category)
			stories, errMsg := m.service.FetchItems(itemsToFetch, msg.Category)

			m.items[msg.Category] = stories

			return message.CategoryFetchingFinished{Index: msg.Index, Cursor: msg.Cursor, Message: errMsg}
		}

	case message.Refresh:
		return m, func() tea.Msg {
			itemsToFetch := m.getNumberOfItemsToFetch(msg.CurrentCategory)
			stories, errMsg := m.service.FetchItems(itemsToFetch, msg.CurrentCategory)

			m.items[category.Top] = []*item.Item{}
			m.items[category.New] = []*item.Item{}
			m.items[category.Ask] = []*item.Item{}
			m.items[category.Show] = []*item.Item{}
			m.items[category.Best] = []*item.Item{}

			m.items[msg.CurrentCategory] = stories

			return message.CategoryFetchingFinished{Index: msg.CurrentIndex, Cursor: 0, Message: errMsg}
		}

	case message.ShowStatusMessage:
		cmds = append(cmds, m.NewStatusMessageWithDuration(msg.Message, msg.Duration))

	case message.CategoryFetchingFinished:
		m.Paginator.Page = 0
		m.SetDisabledInput(false)
		m.StopSpinner()
		m.isBufferActive = false
		m.cat.SetIndex(msg.Index)

		itemsOnPage := m.Paginator.ItemsOnPage(len(m.VisibleItems()))
		m.cursor = min(msg.Cursor, itemsOnPage-1)

		m.NewStatusMessage(msg.Message)

		m.updatePagination()

	case message.DownloadFile:
		statusMessage := "Document downloaded successfully"
		m.hideStatusMessage()
		m.disableInput = false

		filename, _ := download.GuessDownloadFileName(msg.Url)
		downloadDir, _ := download.GetDownloadDir()
		absoluteFilepath := fmt.Sprintf("%s/%s", downloadDir, filename)
		err := download.DownloadFile(msg.Url, absoluteFilepath)
		if err != nil {
			statusMessage = "Could not download document"
		}

		return m, m.NewStatusMessageWithDuration(statusMessage, time.Second*3)
	}

	if m.isOnHelpScreen {
		return m.updateHelpScreen(msg)
	}

	cmds = append(cmds, m.handleBrowsing(msg))

	return m, tea.Batch(cmds...)
}

func (m *Model) updateHelpScreen(msg tea.Msg) (*Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" || k == "i" || k == "?" {
			m.isOnHelpScreen = false

			return m, nil
		}

	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().GetFrameSize()
		m.SetSize(msg.Width-h, msg.Height-v)

		headerHeight := lipgloss.Height("")
		footerHeight := lipgloss.Height("")
		verticalMarginHeight := headerHeight + footerHeight

		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - verticalMarginHeight

		m.width = msg.Width
		m.height = msg.Height

		content := lipgloss.NewStyle().
			Width(msg.Width).
			AlignHorizontal(lipgloss.Center).
			SetString(help.GetHelpScreen(m.config.EnableNerdFonts))

		m.viewport.SetContent(content.String())

		return m, viewport.Sync(m.viewport)

	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) updateCursor() {
	m.cursor = min(m.cursor, m.Paginator.ItemsOnPage(len(m.VisibleItems()))-1)
}

func (m *Model) categoryHasStories(cat int) bool {
	return len(m.items[cat]) != 0
}

func (m *Model) changeToNextCategory() {
	m.cat.Next(m.favorites.HasItems())
	currentCategory := m.cat.GetCurrentCategory(m.favorites.HasItems())

	m.Paginator.Page = 0
	m.cursor = min(m.cursor, len(m.items[currentCategory])-1)
	m.updatePagination()
}

func (m *Model) changeToPrevCategory() {
	m.cat.Prev(m.favorites.HasItems())
	currentCategory := m.cat.GetCurrentCategory(m.favorites.HasItems())

	m.Paginator.Page = 0
	m.cursor = min(m.cursor, len(m.items[currentCategory])-1)
	m.updatePagination()
}

func (m *Model) handleBrowsing(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	numItems := len(m.VisibleItems())

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case msg.String() == "i" || msg.String() == "?":
			m.isOnHelpScreen = true

			return nil

		case m.onAddToFavoritesPrompt && msg.String() == "y":
			m.onAddToFavoritesPrompt = false
			m.disableInput = false

			addToFavorites := func() tea.Msg {
				return message.AddToFavorites{Item: m.SelectedItem()}
			}

			cmds = append(cmds, addToFavorites)
			cmds = append(cmds, m.NewStatusMessageWithDuration("Item added", time.Second*2))

			return tea.Batch(cmds...)

		case m.onRemoveFromFavoritesPrompt && msg.String() == "y":
			m.onRemoveFromFavoritesPrompt = false
			m.disableInput = false

			//
			m.favorites.Remove(m.Index())
			m.items[category.Favorites] = m.favorites.GetItems()

			m.favorites.Write()

			//
			isOnLastItem := m.Index() == len(m.items[category.Favorites])
			hasOnlyOneItem := len(m.items[category.Favorites]) == 0

			itemRemovedMessage := "Item removed"

			if hasOnlyOneItem {
				m.cat.SetIndex(0)
				m.updateCursor()
				m.updatePagination()

				cmds = append(cmds, func() tea.Msg {
					return message.FetchAndChangeToCategory{Index: m.cat.GetCurrentIndex(), Category: m.cat.GetCurrentCategory(false), Cursor: 0}
				})
				cmds = append(cmds, m.NewStatusMessageWithDuration(itemRemovedMessage, time.Second*2))

				return tea.Batch(cmds...)
			}

			if isOnLastItem {
				m.cursor = m.cursor - 1
			}

			m.updatePagination()

			return m.NewStatusMessageWithDuration(itemRemovedMessage, time.Second*2)

		case m.onAddToFavoritesPrompt || m.onRemoveFromFavoritesPrompt:
			m.onAddToFavoritesPrompt = false
			m.onRemoveFromFavoritesPrompt = false
			m.disableInput = false

			m.hideStatusMessage()

			return nil

		case m.disableInput:
			return nil

		case msg.String() == "q" || msg.String() == "esc" || msg.String() == "ctrl+c":
			return tea.Quit

		case msg.String() == "up" || msg.String() == "k":
			if m.cursor == 0 && !m.Paginator.OnFirstPage() {
				m.Paginator.PrevPage()
				itemsOnPage := m.Paginator.ItemsOnPage(len(m.VisibleItems()))
				m.cursor = max(m.cursor, itemsOnPage-1)

				return nil
			}
			m.CursorUp()

			return nil

		case msg.String() == "down" || msg.String() == "j":
			itemsOnPage := m.Paginator.ItemsOnPage(len(m.VisibleItems()))
			if m.cursor == itemsOnPage-1 && !m.Paginator.OnLastPage() {
				m.Paginator.NextPage()
				m.cursor = 0

				return nil
			}
			m.CursorDown()

			return nil

		case msg.String() == "left" || msg.String() == "h":
			m.Paginator.PrevPage()
			m.updateCursor()

			return nil

		case msg.String() == "right" || msg.String() == "l":
			m.Paginator.NextPage()
			m.updateCursor()

			return nil

		case msg.String() == "tab":
			nextIndex := m.cat.GetNextIndex(m.favorites.HasItems())
			nextCat := m.cat.GetNextCategory(m.favorites.HasItems())

			if m.categoryHasStories(nextCat) {
				m.changeToNextCategory()

				return nil
			}

			m.SetDisabledInput(true)
			startSpinnerCmd := m.StartSpinner()

			//m.categoryToDisplay = nextCat

			changeCatCmd := func() tea.Msg {
				return message.FetchAndChangeToCategory{Index: nextIndex, Category: nextCat, Cursor: m.cursor}
			}

			cmds = append(cmds, startSpinnerCmd)
			cmds = append(cmds, changeCatCmd)

			return tea.Batch(cmds...)

		case msg.String() == "shift+tab":
			prevIndex := m.cat.GetPrevIndex(m.favorites.HasItems())
			prevCat := m.cat.GetPrevCategory(m.favorites.HasItems())

			if m.categoryHasStories(prevCat) {
				m.changeToPrevCategory()

				return nil
			}

			m.SetDisabledInput(true)
			startSpinnerCmd := m.StartSpinner()

			//m.categoryToDisplay = prevCat

			changeCatCmd := func() tea.Msg {
				return message.FetchAndChangeToCategory{Index: prevIndex, Category: prevCat, Cursor: m.cursor}
			}

			cmds = append(cmds, startSpinnerCmd)
			cmds = append(cmds, changeCatCmd)

			return tea.Batch(cmds...)

		case msg.String() == "g":
			m.cursor = 0

			return nil

		case msg.String() == "G":
			m.cursor = m.Paginator.ItemsOnPage(numItems) - 1

			return nil

		case msg.String() == "o":
			if m.SelectedItem().URL == "" {
				url := "https://news.ycombinator.com/item?id=" + strconv.Itoa(m.SelectedItem().ID)
				browser.Open(url)
			} else {
				browser.Open(m.SelectedItem().URL)
			}

			cmd := func() tea.Msg {
				return message.OpeningLink{
					Id:           m.SelectedItem().ID,
					CommentCount: m.SelectedItem().CommentsCount,
				}
			}

			return cmd

		case msg.String() == "c":
			url := "https://news.ycombinator.com/item?id=" + strconv.Itoa(m.SelectedItem().ID)
			browser.Open(url)

			cmd := func() tea.Msg {
				return message.OpeningCommentsInBrowser{
					Id:           m.SelectedItem().ID,
					CommentCount: m.SelectedItem().CommentsCount,
				}
			}

			return cmd

		case msg.String() == "r" && m.cat.GetCurrentCategory(m.favorites.HasItems()) != category.Favorites:
			currentCategory := m.cat.GetCurrentCategory(m.favorites.HasItems())
			currentPage := m.Paginator.Page

			m.items[category.Buffer] = m.items[currentCategory]

			m.isBufferActive = true
			m.Paginator.Page = 0
			m.cursor = min(m.cursor, len(m.items[currentCategory])-1)
			m.updatePagination()

			m.SetDisabledInput(true)
			m.cursor = 0
			m.Paginator.Page = currentPage

			changeCatCmd := func() tea.Msg {
				return message.Refresh{CurrentIndex: m.cat.GetCurrentIndex(), CurrentCategory: currentCategory}
			}

			cmds = append(cmds, m.StartSpinner())
			cmds = append(cmds, changeCatCmd)

			return tea.Batch(cmds...)

		case msg.String() == "f" || msg.String() == "V":
			m.SetPermanentStatusMessage(getAddItemConfirmationMessage(), false)
			m.onAddToFavoritesPrompt = true
			m.disableInput = true

			return nil

		case msg.String() == "x" && m.cat.GetCurrentCategory(m.favorites.HasItems()) == category.Favorites:
			m.SetPermanentStatusMessage(getRemoveItemConfirmationMessage(), false)
			m.onRemoveFromFavoritesPrompt = true
			m.disableInput = true

			return nil

		case msg.String() == "enter":
			m.SetIsVisible(false)
			m.SetDisabledInput(true)

			cmd := func() tea.Msg {
				return message.EnteringCommentSection{
					Id:           m.SelectedItem().ID,
					CommentCount: m.SelectedItem().CommentsCount,
				}
			}

			return cmd

		case msg.String() == " ":
			m.SetIsVisible(false)
			m.SetDisabledInput(true)

			return func() tea.Msg {
				return message.EnteringReaderMode{
					Url:          m.SelectedItem().URL,
					Title:        m.SelectedItem().Title,
					Domain:       m.SelectedItem().Domain,
					Id:           m.SelectedItem().ID,
					CommentCount: m.SelectedItem().CommentsCount,
				}
			}

		case msg.String() == "d":
			if !strings.Contains(m.SelectedItem().Title, "[pdf]") {
				return m.NewStatusMessage("Only PDFs are currently supported ...")
			}
			m.SetDisabledInput(true)

			downloadCmd := func() tea.Msg {
				return message.DownloadFile{
					Url: m.SelectedItem().URL,
				}
			}

			cmds = append(cmds, downloadCmd)
			cmds = append(cmds, m.NewStatusMessage("Downloading ..."))

			return tea.Batch(cmds...)
		}
	}

	cmd := m.delegate.Update(msg, m)
	cmds = append(cmds, cmd)

	// Keep the index in bounds when paginating
	itemsOnPage := m.Paginator.ItemsOnPage(len(m.VisibleItems()))
	if m.cursor > itemsOnPage-1 {
		m.cursor = max(0, itemsOnPage-1)
	}

	return tea.Batch(cmds...)
}

func (m *Model) showHelpScreen() tea.Cmd {
	helpScreen := help.GetHelpScreen(m.config.EnableNerdFonts)

	command := cli.Less(helpScreen, m.config)

	return tea.ExecProcess(command, func(err error) tea.Msg {
		return message.EditorFinishedMsg{Err: err}
	})
}

// View renders the component.
func (m *Model) View() string {
	if m.isOnHelpScreen {
		categoriesWithoutFavorites := m.cat.GetCategories(false)
		hasFavorites := m.favorites.HasItems()

		return fmt.Sprintf("%s\n%s\n%s",
			header.GetHeader(
				categoriesWithoutFavorites,
				hasFavorites,
				m.cat.GetCurrentIndex(),
				m.width),
			m.viewport.View(),
			m.statusAndPaginationView())
	}

	var (
		sections    []string
		availHeight = m.height
	)

	if !m.isVisible {
		return ""
	}

	if m.showTitle {
		v := m.titleView()
		sections = append(sections, v)
		availHeight -= lipgloss.Height(v)
	}

	if m.showStatusBar {
		v := m.statusAndPaginationView()
		availHeight -= lipgloss.Height(v)

	}

	content := lipgloss.NewStyle().Height(availHeight).Render(m.populatedView())
	totalItems := len(m.items[m.cat.GetCurrentCategory(m.favorites.HasItems())])
	rankings := ranking.GetRankings(
		false,
		m.Paginator.PerPage,
		totalItems,
		m.cursor,
		m.Paginator.Page,
		m.Paginator.TotalPages)

	rankingsAndContent := lipgloss.JoinHorizontal(lipgloss.Top, rankings, content)
	sections = append(sections, rankingsAndContent)

	if m.showStatusBar {
		v := m.statusAndPaginationView()
		sections = append(sections, v)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Model) titleView() string {
	categoriesWithoutFavorites := m.cat.GetCategories(false)
	hasFavorites := m.favorites.HasItems()

	return header.GetHeader(
		categoriesWithoutFavorites,
		hasFavorites,
		m.cat.GetCurrentIndex(),
		m.width)
}

func (m *Model) statusAndPaginationView() string {
	centerContent := ""
	rightContent := ""
	underscore := lipgloss.NewStyle().Underline(true).Render(" ")
	underline := strings.Repeat(underscore, screen.GetTerminalWidth())

	if m.isOnHelpScreen {
		centerContent = lipgloss.NewStyle().Faint(true).Render(
			"github.com/bensadeh/circumflex • version " + app.Version)
	} else if m.showSpinner {
		centerContent = m.spinnerView()
	} else {
		centerContent = m.statusMessage
	}

	if m.isOnHelpScreen {
		rightContent = ""
	} else {
		rightContent = m.Paginator.View()
	}

	left := lipgloss.NewStyle().
		Inline(true).
		Width(5).
		MaxWidth(5).
		Render("")

	center := lipgloss.NewStyle().
		Inline(true).
		Width(m.width - 5 - 5).
		Align(lipgloss.Center).
		Render(centerContent)

	right := lipgloss.NewStyle().
		Inline(true).
		Width(5).
		Align(lipgloss.Center).
		Render(rightContent)

	return underline + "\n" +
		m.Styles.StatusBar.Render(left) +
		m.Styles.StatusBar.Render(center) +
		m.Styles.StatusBar.Render(right)
}

func (m *Model) OnStartup() bool {
	return m.onStartup
}

func (m *Model) IsInputDisabled() bool {
	return m.disableInput
}

func (m *Model) SetDisabledInput(value bool) {
	m.disableInput = value
}

func (m *Model) SetOnStartup(value bool) {
	m.onStartup = value
}

func (m *Model) populatedView() string {
	allItems := m.VisibleItems()

	var b strings.Builder

	// Empty states
	if len(allItems) == 0 {
		return m.Styles.NoItems.Render("")
	}

	if len(allItems) > 0 {
		start, end := m.Paginator.GetSliceBounds(len(allItems))
		itemsToShow := allItems[start:end]

		for i, item := range itemsToShow {
			m.delegate.Render(&b, m, i+start, item)
			if i != len(itemsToShow)-1 {
				fmt.Fprint(&b, strings.Repeat("\n", m.delegate.Spacing()+1))
			}
		}
	}

	// If there aren't enough items to fill up this page (always the last page)
	// then we need to add some newlines to fill up the space where items would
	// have been.
	itemsOnPage := m.Paginator.ItemsOnPage(len(allItems))
	if itemsOnPage < m.Paginator.PerPage {
		n := (m.Paginator.PerPage - itemsOnPage) * (m.delegate.Height() + m.delegate.Spacing())
		if len(allItems) == 0 {
			n -= m.delegate.Height() - 1
		}
		_, _ = fmt.Fprint(&b, strings.Repeat("\n", n))
	}

	return b.String()
}

func (m *Model) spinnerView() string {
	return m.spinner.View()
}

func getAddItemConfirmationMessage() string {
	normal := lipgloss.NewStyle()
	green := normal.Copy().
		Foreground(lipgloss.Color("2"))
	bold := normal.Copy().
		Foreground(lipgloss.Color("4")).
		Bold(true)

	return green.Render("Add") + normal.Render(" to Favorites? Press ") + bold.Render("y") +
		normal.Render(" to confirm")
}

func getRemoveItemConfirmationMessage() string {
	normal := lipgloss.NewStyle()
	red := normal.Copy().
		Foreground(lipgloss.Color("1"))
	bold := normal.Copy().
		Foreground(lipgloss.Color("4")).
		Bold(true)

	return red.Render("Remove") + normal.Render(" from Favorites? Press ") + bold.Render("y") +
		normal.Render(" to confirm")
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}
