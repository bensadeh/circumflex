package list

import (
	"io"
	"time"

	"clx/categories"

	"charm.land/bubbles/v2/viewport"

	"clx/reader"

	"clx/bubble/list/message"
	"clx/cli"
	"clx/constants/category"
	"clx/favorites"
	"clx/help"
	"clx/history"
	"clx/hn"
	"clx/item"
	"clx/settings"
	"clx/tree"
	"clx/validator"

	"charm.land/bubbles/v2/paginator"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	numberOfCategories = 6
)


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

func New(delegate ItemDelegate, config *settings.Config, cat *categories.Categories, favorites *favorites.Favorites, width, height int) *Model {
	return newModel(delegate, config, cat, favorites, width, height,
		getService(config.DebugMode),
		getHistory(config.DebugMode, config.DoNotMarkSubmissionsAsRead))
}

func newModel(delegate ItemDelegate, config *settings.Config, cat *categories.Categories, favorites *favorites.Favorites, width, height int, service hn.Service, hist history.History) *Model {
	styles := DefaultStyles()

	sp := spinner.New()
	sp.Spinner = getSpinner()
	sp.Style = styles.Spinner

	p := paginator.New()
	p.Type = paginator.Dots
	p.ActiveDot = styles.ActivePaginationDot.String()
	p.InactiveDot = styles.InactivePaginationDot.String()

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
		history:      hist,
		items:        items,
		Paginator:    p,
		spinner:      sp,
		onStartup:    true,
		isVisible:    true,
		disableInput: true,
		config:       config,
		service:      service,
		favorites:    favorites,
		cat:          cat,
	}

	m.updatePagination()

	return &m
}

func (m *Model) SetIsVisible(v bool) {
	m.isVisible = v
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
		m.setSize(windowSizeMsg.Width-h, windowSizeMsg.Height-v)

		var cmds []tea.Cmd

		spinnerCmd := m.StartSpinner()
		cmds = append(cmds, spinnerCmd)

		m.SetOnStartup(false)

		m.items[category.Favorites] = m.favorites.GetItems()

		fetchCmd := m.FetchStoriesForFirstCategory()
		cmds = append(cmds, fetchCmd)

		heightOfHeaderAndStatusLine := 4

		m.viewport = viewport.New(viewport.WithWidth(windowSizeMsg.Width), viewport.WithHeight(windowSizeMsg.Height-heightOfHeaderAndStatusLine))

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
		m.items[msg.Category] = msg.Stories
		m.StopSpinner()
		m.disableInput = false
		m.updatePagination()
		cmd := m.NewStatusMessage(msg.Message)

		return m, cmd

	case message.StatusMessageTimeout:
		m.hideStatusMessage()

	case message.AddToFavorites:
		m.favorites.Add(msg.Item)
		m.items[category.Favorites] = m.favorites.GetItems()

		m.favorites.Write()

		m.updatePagination()

	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().GetFrameSize()
		m.setSize(msg.Width-h, msg.Height-v)

		headerHeight := 2
		footerHeight := 2
		verticalMarginHeight := headerHeight + footerHeight

		m.viewport.SetWidth(msg.Width)
		m.viewport.SetHeight(msg.Height - verticalMarginHeight)

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
		return m, m.fetchAndChangeToCategory(msg)

	case message.Refresh:
		return m, m.refresh(msg)

	case message.ShowStatusMessage:
		cmds = append(cmds, m.NewStatusMessageWithDuration(msg.Message, msg.Duration))

	case message.CategoryFetchingFinished:
		if m.isBufferActive {
			clearAllCategories(m.items)
		}
		m.items[msg.Category] = msg.Stories
		m.Paginator.Page = 0
		m.SetDisabledInput(false)
		m.StopSpinner()
		m.isBufferActive = false
		m.cat.SetIndex(msg.Index)

		itemsOnPage := m.Paginator.ItemsOnPage(len(m.VisibleItems()))
		m.cursor = min(msg.Cursor, itemsOnPage-1)

		cmd := m.NewStatusMessage(msg.Message)
		cmds = append(cmds, cmd)

		m.updatePagination()
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
		m.setSize(msg.Width-h, msg.Height-v)

		headerHeight := lipgloss.Height("")
		footerHeight := lipgloss.Height("")
		verticalMarginHeight := headerHeight + footerHeight

		m.viewport.SetWidth(msg.Width)
		m.viewport.SetHeight(msg.Height - verticalMarginHeight)

		m.width = msg.Width
		m.height = msg.Height

		content := lipgloss.NewStyle().
			Width(msg.Width).
			AlignHorizontal(lipgloss.Center).
			SetString(help.GetHelpScreen(m.config.EnableNerdFonts))

		m.viewport.SetContent(content.String())

		return m, nil

	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) updateCursor() {
	m.cursor = min(m.cursor, m.Paginator.ItemsOnPage(len(m.VisibleItems()))-1)
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


