package list

import (
	"clx/bubble/list/message"
	"clx/categories"
	"clx/cli"
	"clx/favorites"
	"clx/help"
	"clx/history"
	"clx/hn"
	"clx/item"
	"clx/settings"
	"context"
	"io"
	"time"

	"charm.land/bubbles/v2/paginator"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
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
	Render(w io.Writer, m *Model, index int, item *item.Story)

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

	Title  string
	Styles Styles

	state       ViewState
	spinner     spinner.Model
	showSpinner bool
	width       int
	height      int
	Paginator   paginator.Model
	cursor      int

	StatusMessageLifetime time.Duration

	statusMessage      string
	statusMessageTimer *time.Timer
	items              [][]*item.Story

	delegate  ItemDelegate
	history   history.History
	config    *settings.Config
	service   hn.Service
	favorites *favorites.Favorites
	cat       *categories.Categories
	keymap    KeyMap

	viewport viewport.Model
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
	items := make([][]*item.Story, numberOfCategories+bufferCategory)

	m := Model{
		showTitle:             true,
		showStatusBar:         true,
		Styles:                styles,
		Title:                 "List",
		StatusMessageLifetime: time.Second,

		state:     StateStartup,
		width:     width,
		height:    height,
		delegate:  delegate,
		history:   hist,
		items:     items,
		Paginator: p,
		spinner:   sp,
		config:    config,
		service:   service,
		favorites: favorites,
		cat:       cat,
		keymap:    DefaultKeyMap(),
	}

	m.updatePagination()

	return &m
}

func (m *Model) setSize(width, height int) {
	m.width = width
	m.height = height
	m.updatePagination()
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	windowSizeMsg, isWindowSizeMsg := msg.(tea.WindowSizeMsg)

	// Since this program is using the full size of the viewport we
	// need to wait until we've received the window dimensions before
	// we can initialize the viewport. The initial dimensions come in
	// quickly, though asynchronously, which is why we wait for them
	// here.
	if m.state == StateStartup && !isWindowSizeMsg {
		return m, nil
	}

	if m.state == StateStartup && isWindowSizeMsg {
		h, v := lipgloss.NewStyle().GetFrameSize()
		m.setSize(windowSizeMsg.Width-h, windowSizeMsg.Height-v)

		var cmds []tea.Cmd

		spinnerCmd := m.StartSpinner()
		cmds = append(cmds, spinnerCmd)

		m.state = StateLoading

		m.items[categories.Favorites] = m.favorites.GetItems()

		fetchCmd := m.FetchStoriesForFirstCategory()
		cmds = append(cmds, fetchCmd)

		heightOfHeaderAndStatusLine := 4

		m.viewport = viewport.New(viewport.WithWidth(windowSizeMsg.Width), viewport.WithHeight(windowSizeMsg.Height-heightOfHeaderAndStatusLine))

		content := lipgloss.NewStyle().
			Width(windowSizeMsg.Width).
			AlignHorizontal(lipgloss.Center).
			SetString(help.GetHelpScreen(m.config.EnableNerdFonts, m.keymap.MainMenuBindings()))

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
		m.state = StateBrowsing
		m.updatePagination()
		cmd := m.NewStatusMessage(msg.Message)

		return m, cmd

	case message.StatusMessageTimeout:
		m.hideStatusMessage()

	case message.AddToFavorites:
		m.favorites.Add(msg.Item)
		m.items[categories.Favorites] = m.favorites.GetItems()

		if err := m.favorites.Write(); err != nil {
			cmds = append(cmds, m.NewStatusMessageWithDuration("Could not save favorites", time.Second*3))
		}

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
			SetString(help.GetHelpScreen(m.config.EnableNerdFonts, m.keymap.MainMenuBindings()))

		m.viewport.SetContent(content.String())

		return m, nil

	case message.EnteringCommentSection:
		return m, m.handleEnteringCommentSection(msg)

	case message.OpeningLink:
		_ = m.history.MarkAsReadAndWriteToDisk(msg.Id, msg.CommentCount)

	case message.OpeningCommentsInBrowser:
		_ = m.history.MarkAsReadAndWriteToDisk(msg.Id, msg.CommentCount)

	case message.EnteringReaderMode:
		return m, m.handleEnteringReaderMode(msg)

	case message.CommentTreeReady:
		if msg.UpdatedStory != nil {
			m.favorites.UpdateStoryAndWriteToDisk(msg.UpdatedStory)
		}

		if msg.Error != "" {
			return m, tea.Batch(
				m.NewStatusMessageWithDuration(msg.Error, time.Second*3),
				func() tea.Msg {
					return message.EditorFinishedMsg{Err: nil}
				},
			)
		}

		command := cli.Less(context.Background(), msg.Content, m.config)

		return m, tea.ExecProcess(command, func(err error) tea.Msg {
			return message.EditorFinishedMsg{Err: err}
		})

	case message.ArticleReady:
		if msg.Error != "" {
			return m, tea.Batch(
				m.NewStatusMessageWithDuration(msg.Error, time.Second*3),
				func() tea.Msg {
					return message.EditorFinishedMsg{Err: nil}
				},
			)
		}

		command := cli.Less(context.Background(), msg.Content, m.config)

		return m, tea.ExecProcess(command, func(err error) tea.Msg {
			return message.EditorFinishedMsg{Err: err}
		})

	case message.EditorFinishedMsg:
		m.state = StateBrowsing

	case message.FetchAndChangeToCategory:
		return m, m.fetchAndChangeToCategory(msg)

	case message.Refresh:
		return m, m.refresh(msg)

	case message.ShowStatusMessage:
		cmds = append(cmds, m.NewStatusMessageWithDuration(msg.Message, msg.Duration))

	case message.CategoryFetchingFinished:
		m.items[categories.Buffer] = nil

		if msg.Message != "" {
			m.cat.SetIndex(msg.PrevIndex)
			m.state = StateBrowsing
			m.StopSpinner()
			m.updatePagination()

			return m, m.NewStatusMessageWithDuration(msg.Message, time.Second*3)
		}

		if m.state == StateRefreshing {
			clearAllCategories(m.items)
		}
		m.items[msg.Category] = msg.Stories
		m.Paginator.Page = 0
		m.state = StateBrowsing
		m.StopSpinner()
		m.cat.SetIndex(msg.Index)

		itemsOnPage := m.Paginator.ItemsOnPage(len(m.VisibleItems()))
		m.cursor = min(msg.Cursor, itemsOnPage-1)

		cmd := m.NewStatusMessage(msg.Message)
		cmds = append(cmds, cmd)

		m.updatePagination()
	}

	if m.state == StateHelpScreen {
		return m.updateHelpScreen(msg)
	}

	cmds = append(cmds, m.handleBrowsing(msg))

	return m, tea.Batch(cmds...)
}
