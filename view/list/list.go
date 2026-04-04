package list

import (
	"context"
	"io"
	"time"

	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/favorites"
	"github.com/bensadeh/circumflex/help"
	"github.com/bensadeh/circumflex/history"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/item"
	"github.com/bensadeh/circumflex/settings"
	clxspinner "github.com/bensadeh/circumflex/spinner"
	"github.com/bensadeh/circumflex/view/comments"
	"github.com/bensadeh/circumflex/view/message"
	"github.com/bensadeh/circumflex/view/reader"

	"charm.land/bubbles/v2/paginator"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	numberOfCategories    = 6
	statusBarEdgeWidth    = 5
	headerHeight          = 2
	footerHeight          = 2
	headerAndFooterHeight = headerHeight + footerHeight
	statusMessageShort    = 2 * time.Second
	statusMessageLong     = 3 * time.Second
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
	styles styles

	state  ViewState
	status statusBar
	pager  pager
	width  int
	height int

	delegate    ItemDelegate
	history     history.History
	config      *settings.Config
	service     hn.Service
	favorites   *favorites.Favorites
	cat         *categories.Categories
	keymap      keyMap
	fetchCtx    context.Context //nolint:containedctx // single active fetch context, accessed only from the Update goroutine
	cancelFetch context.CancelFunc
	fetchID     uint64

	viewport    viewport.Model
	commentView *comments.Model
	readerView  *reader.Model

	// Cached styles for hot-path rendering.
	contentStyle    lipgloss.Style
	underlineStyle  lipgloss.Style
	statusLeftStyle lipgloss.Style
	statusMidStyle  lipgloss.Style
	statusEndStyle  lipgloss.Style
}

func New(delegate ItemDelegate, config *settings.Config, cat *categories.Categories, favorites *favorites.Favorites, width, height int) *Model {
	return newModel(delegate, config, cat, favorites, width, height,
		hn.NewService(config.DebugMode, config.DebugFallible),
		getHistory(config.DebugMode || config.DebugFallible, config.DoNotMarkSubmissionsAsRead))
}

func newModel(delegate ItemDelegate, config *settings.Config, cat *categories.Categories, favorites *favorites.Favorites, width, height int, service hn.Service, hist history.History) *Model {
	s := defaultStyles()

	sp := spinner.New()
	sp.Spinner = clxspinner.Random()
	sp.Style = s.Spinner

	p := paginator.New()
	p.Type = paginator.Dots
	p.ActiveDot = s.ActivePaginationDot.String()
	p.InactiveDot = s.InactivePaginationDot.String()

	items := make([][]*item.Story, numberOfCategories)

	m := Model{
		styles: s,

		state:  StateStartup,
		width:  width,
		height: height,
		pager: pager{
			items:     items,
			Paginator: p,
		},
		status: statusBar{
			spinner: sp,
		},
		delegate:  delegate,
		history:   hist,
		config:    config,
		service:   service,
		favorites: favorites,
		cat:       cat,
		keymap:    defaultKeyMap(),

		contentStyle:    lipgloss.NewStyle(),
		underlineStyle:  lipgloss.NewStyle().Underline(true),
		statusLeftStyle: lipgloss.NewStyle().Inline(true).Width(statusBarEdgeWidth).MaxWidth(statusBarEdgeWidth),
		statusMidStyle:  lipgloss.NewStyle().Inline(true).Align(lipgloss.Center),
		statusEndStyle:  lipgloss.NewStyle().Inline(true).Width(statusBarEdgeWidth).Align(lipgloss.Center),
	}

	m.updatePagination()

	return &m
}

func (m *Model) syncFavorites() {
	m.pager.items[categories.Favorites] = m.favorites.Items()
	m.cat.SetFavorites(m.favorites.HasItems())
}

func (m *Model) setSize(width, height int) {
	m.width = width
	m.height = height
	m.updatePagination()
}

func (m *Model) handleCategoryFetchingFinished(msg message.CategoryFetchingFinished) (*Model, tea.Cmd) {
	if msg.FetchID != m.fetchID {
		return m, nil
	}

	if msg.Err != nil {
		if m.pager.transition != nil {
			m.cat.SetIndex(m.pager.transition.prevIndex)
		}

		m.pager.transition = nil
		m.state = StateBrowsing
		m.status.StopSpinner()
		m.updatePagination()

		return m, m.status.NewStatusMessageWithDuration(friendlyError(msg.Err), statusMessageLong)
	}

	if m.pager.transition != nil && m.pager.transition.refresh {
		clearAllCategories(m.pager.items)
	}

	m.pager.transition = nil
	m.pager.items[msg.Category] = msg.Stories
	m.pager.Paginator.Page = 0
	m.state = StateBrowsing
	m.status.StopSpinner()
	m.cat.SetIndex(msg.Index)

	itemsOnPage := m.pager.Paginator.ItemsOnPage(len(m.VisibleItems()))
	m.pager.cursor = min(msg.Cursor, itemsOnPage-1)

	m.updatePagination()

	return m, nil
}

func (m *Model) handleFetchingFinished(msg message.FetchingFinished) (*Model, tea.Cmd) {
	if msg.FetchID != m.fetchID {
		return m, nil
	}

	m.status.StopSpinner()
	m.state = StateBrowsing

	if msg.Err != nil {
		return m, m.status.NewStatusMessageWithDuration(friendlyError(msg.Err), statusMessageLong)
	}

	m.pager.items[msg.Category] = msg.Stories
	m.updatePagination()

	return m, nil
}

func (m *Model) handleWindowResize(msg tea.WindowSizeMsg) (*Model, tea.Cmd) {
	m.setSize(msg.Width, msg.Height)

	if m.state == StateReaderView {
		return m, m.readerView.Update(msg)
	}

	if m.state == StateCommentView {
		return m, m.commentView.Update(msg)
	}

	m.viewport.SetWidth(msg.Width)
	m.viewport.SetHeight(msg.Height - headerAndFooterHeight)

	content := lipgloss.NewStyle().
		Width(msg.Width).
		SetString(help.MainMenuHelpScreen(msg.Width, m.keymap.MainMenuBindings()))

	m.viewport.SetContent(content.String())

	return m, nil
}

func (m *Model) handleStartup(msg tea.WindowSizeMsg) (*Model, tea.Cmd) {
	m.setSize(msg.Width, msg.Height)

	var cmds []tea.Cmd

	spinnerCmd := m.startFetch(0)
	cmds = append(cmds, spinnerCmd)

	m.syncFavorites()

	fetchCmd := m.FetchStoriesForFirstCategory()
	cmds = append(cmds, fetchCmd)
	cmds = append(cmds, scheduleTimeRefresh())

	m.viewport = viewport.New(viewport.WithWidth(msg.Width), viewport.WithHeight(msg.Height-headerAndFooterHeight))

	content := lipgloss.NewStyle().
		Width(msg.Width).
		SetString(help.MainMenuHelpScreen(msg.Width, m.keymap.MainMenuBindings()))

	m.viewport.SetContent(content.String())

	return m, tea.Batch(cmds...)
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	if m.state == StateStartup {
		if windowSizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
			return m.handleStartup(windowSizeMsg)
		}

		return m, nil
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		newSpinnerModel, cmd := m.status.spinner.Update(msg)

		m.status.spinner = newSpinnerModel
		if m.status.showSpinner {
			cmds = append(cmds, cmd)
		}

	case message.TimeRefreshTick:
		cmds = append(cmds, scheduleTimeRefresh())

	case message.FetchingFinished:
		return m.handleFetchingFinished(msg)

	case message.StatusMessageTimeout:
		if msg.Generation == m.status.generation {
			m.status.hideStatusMessage()
			clearProgress()
		}

	case message.AddToFavorites:
		m.favorites.Add(msg.Item)

		if err := m.favorites.Write(); err != nil {
			_ = m.favorites.Remove(len(m.favorites.Items()) - 1)
			m.syncFavorites()
			cmds = append(cmds, m.status.NewStatusMessageWithDuration("Could not save favorite to disk", statusMessageLong))

			break
		}

		m.syncFavorites()
		m.updatePagination()

	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg)

	case message.EnteringCommentSection:
		return m, m.handleEnteringCommentSection(msg)

	case message.BrowserOpenFailed:
		cmds = append(cmds, m.status.NewStatusMessageWithDuration("Could not open browser", statusMessageLong))

	case message.EnteringReaderMode:
		return m, m.handleEnteringReaderMode(msg)

	case message.ArticleReady:
		if msg.FetchID != m.fetchID {
			return m, nil
		}

		m.pager.transition = nil
		m.status.StopSpinner()

		if msg.Err != nil {
			m.state = StateBrowsing

			return m, m.status.NewStatusMessageWithDuration(friendlyError(msg.Err), statusMessageLong)
		}

		if msg.Parsed != nil {
			m.readerView = reader.NewWithArticle(msg.Parsed, msg.Title, m.config.ArticleWidth, m.width, m.height, reader.Meta{
				URL:       msg.URL,
				Author:    msg.Author,
				TimeAgo:   msg.TimeAgo,
				ID:        msg.ID,
				Points:    msg.Points,
				NerdFonts: m.config.EnableNerdFonts,
			})
		} else {
			m.readerView = reader.New(msg.Content, msg.Title, m.width, m.height)
		}

		m.state = StateReaderView

		return m, m.readerView.Init()

	case message.ReaderViewQuitMsg:
		m.readerView = nil
		m.state = StateBrowsing

		return m, nil

	case message.FetchAndChangeToCategory:
		return m, m.fetchAndChangeToCategory(msg)

	case message.Refresh:
		return m, m.refresh(msg)

	case message.ShowStatusMessage:
		cmds = append(cmds, m.status.NewStatusMessageWithDuration(msg.Message, msg.Duration))

	case message.CommentTreeDataReady:
		if msg.FetchID != m.fetchID {
			return m, nil
		}

		m.pager.transition = nil
		m.status.StopSpinner()

		if msg.UpdatedStory != nil {
			_ = m.favorites.UpdateStoryAndWriteToDisk(msg.UpdatedStory)
		}

		if msg.Err != nil {
			m.state = StateBrowsing

			return m, m.status.NewStatusMessageWithDuration(friendlyError(msg.Err), statusMessageLong)
		}

		m.commentView = comments.New(msg.Thread, msg.LastVisited, m.config.CommentWidth, m.config.EnableNerdFonts, m.width, m.height)
		m.state = StateCommentView

		return m, m.commentView.Init()

	case message.CommentViewQuitMsg:
		m.commentView = nil
		m.state = StateBrowsing

		return m, nil

	case message.CategoryFetchingFinished:
		return m.handleCategoryFetchingFinished(msg)
	}

	if m.state == StateReaderView {
		return m, m.readerView.Update(msg)
	}

	if m.state == StateCommentView {
		return m, m.commentView.Update(msg)
	}

	if m.state == StateHelpScreen {
		return m.updateHelpScreen(msg)
	}

	cmds = append(cmds, m.handleBrowsing(msg))

	return m, tea.Batch(cmds...)
}
