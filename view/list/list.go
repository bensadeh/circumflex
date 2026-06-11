package list

import (
	"context"
	"time"

	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/favorites"
	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/history"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/hn/provider"
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
	statusBarEdgeWidth    = 5
	headerHeight          = 2
	footerHeight          = 2
	headerAndFooterHeight = headerHeight + footerHeight
	statusMessageShort    = 2 * time.Second
	statusMessageLong     = 3 * time.Second
)

type Model struct {
	styles styles

	state  viewState
	status statusBar
	pager  pager
	width  int
	height int

	itemStyles  itemStyles
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

	memorialErr error
	browserErr  error

	// Cached styles for hot-path rendering.
	contentStyle    lipgloss.Style
	underlineStyle  lipgloss.Style
	statusLeftStyle lipgloss.Style
	statusMidStyle  lipgloss.Style
	statusEndStyle  lipgloss.Style
}

func New(config *settings.Config, cat *categories.Categories, favorites *favorites.Favorites, width, height int) (*Model, error) {
	hist, err := newHistory(config.DebugMode || config.DebugFallible, config.DoNotMarkSubmissionsAsRead)
	if err != nil {
		return nil, err
	}

	return newModel(config, cat, favorites, width, height,
		provider.NewService(config.DebugMode, config.DebugFallible), hist), nil
}

func newModel(config *settings.Config, cat *categories.Categories, favorites *favorites.Favorites, width, height int, service hn.Service, hist history.History) *Model {
	s := defaultStyles()

	sp := spinner.New()
	sp.Spinner = clxspinner.Random()
	sp.Style = s.Spinner

	p := paginator.New()
	p.Type = paginator.Dots
	p.ActiveDot = s.ActivePaginationDot.String()
	p.InactiveDot = s.InactivePaginationDot.String()

	items := make([][]*hn.Story, categories.Count())

	m := Model{
		styles: s,

		state:  stateStartup,
		width:  width,
		height: height,
		pager: pager{
			items:     items,
			Paginator: p,
		},
		status: statusBar{
			spinner: sp,
		},
		itemStyles: newItemStyles(),
		history:    hist,
		config:     config,
		service:    service,
		favorites:  favorites,
		cat:        cat,
		keymap:     defaultKeyMap(),

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
	m.pager.items[categories.Favorites] = favItemsToStories(m.favorites.Items())
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
		m.state = stateBrowsing
		m.status.StopSpinner()
		m.updatePagination()

		return m, m.status.NewStatusMessageWithDuration(friendlyError(msg.Err), statusMessageLong)
	}

	m.pager.transition = nil
	m.pager.items[msg.Category] = msg.Stories
	m.pager.Paginator.Page = 0
	m.state = stateBrowsing
	m.status.StopSpinner()
	m.cat.SetIndex(msg.Index)

	itemsOnPage := m.pager.Paginator.ItemsOnPage(len(m.VisibleItems()))
	m.pager.cursor = min(msg.Cursor, itemsOnPage-1)

	m.updatePagination()

	return m, nil
}

func (m *Model) handleCommentTreeDataReady(msg message.CommentTreeDataReady) (*Model, tea.Cmd) {
	if msg.FetchID != m.fetchID {
		return m, nil
	}

	var cmds []tea.Cmd

	m.pager.transition = nil
	m.status.StopSpinner()

	if msg.UpdatedStory != nil {
		if err := m.favorites.UpdateStoryAndWriteToDisk(favorites.ItemFromStory(msg.UpdatedStory)); err != nil {
			cmds = append(cmds, m.status.NewStatusMessageWithDuration("Could not update favorite on disk", statusMessageLong))
		}
	}

	if msg.Err != nil {
		m.state = stateBrowsing
		cmds = append(cmds, m.status.NewStatusMessageWithDuration(friendlyError(msg.Err), statusMessageLong))

		return m, tea.Batch(cmds...)
	}

	m.commentView = comments.New(msg.Thread, msg.LastVisited, m.config.CommentWidth, m.config.Indent, m.config.EnableNerdFonts, m.width, m.height)
	m.state = stateCommentView

	cmds = append(cmds, m.commentView.Init())

	if msg.HistoryWarning != nil {
		cmds = append(cmds, m.status.NewStatusMessageWithDuration("Could not save read status", statusMessageShort))
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) handleFetchingFinished(msg message.FetchingFinished) (*Model, tea.Cmd) {
	if msg.FetchID != m.fetchID {
		return m, nil
	}

	m.status.StopSpinner()
	m.state = stateBrowsing

	if msg.Err != nil {
		return m, m.status.NewStatusMessageWithDuration(friendlyError(msg.Err), statusMessageLong)
	}

	m.pager.items[msg.Category] = msg.Stories
	m.updatePagination()

	return m, nil
}

func (m *Model) handleWindowResize(msg tea.WindowSizeMsg) (*Model, tea.Cmd) {
	m.setSize(msg.Width, msg.Height)

	if m.state == stateReaderView {
		return m, m.readerView.Update(msg)
	}

	if m.state == stateCommentView {
		return m, m.commentView.Update(msg)
	}

	m.resizeHelpViewport(msg.Width, msg.Height)

	return m, nil
}

func (m *Model) MemorialErr() error { return m.memorialErr }

func (m *Model) BrowserErr() error { return m.browserErr }

func (m *Model) handleStartup(msg tea.WindowSizeMsg) (*Model, tea.Cmd) {
	m.setSize(msg.Width, msg.Height)

	var cmds []tea.Cmd

	spinnerCmd := m.startFetch(0)
	cmds = append(cmds, spinnerCmd)

	m.syncFavorites()

	fetchCmd := m.fetchStoriesForFirstCategory()
	cmds = append(cmds, fetchCmd)
	cmds = append(cmds, fetchMemorialStatus())
	cmds = append(cmds, scheduleTimeRefresh())

	m.viewport = viewport.New()
	m.resizeHelpViewport(msg.Width, msg.Height)

	return m, tea.Batch(cmds...)
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	if m.state == stateStartup {
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

	case message.MemorialStatusReady:
		// Only apply a definitive result; on error leave the current state so a
		// failed re-check (e.g. during a refresh) never blanks an active bar.
		if msg.Err == nil {
			header.SetMemorial(msg.Active)
		}

		m.memorialErr = msg.Err

	case message.FetchingFinished:
		return m.handleFetchingFinished(msg)

	case message.StatusMessageTimeout:
		if msg.Generation == m.status.generation {
			m.status.hideStatusMessage()
			clearProgress()
		}

	case message.AddToFavorites:
		m.favorites.Add(favorites.ItemFromStory(msg.Item))

		if err := m.favorites.Write(); err != nil {
			if removeErr := m.favorites.RemoveLast(); removeErr != nil {
				cmds = append(cmds, m.status.NewStatusMessageWithDuration("Could not save or rollback favorite", statusMessageLong))
			} else {
				cmds = append(cmds, m.status.NewStatusMessageWithDuration("Could not save favorite to disk", statusMessageLong))
			}

			m.syncFavorites()

			break
		}

		m.syncFavorites()
		m.updatePagination()

	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg)

	case message.EnteringCommentSection:
		return m, m.handleEnteringCommentSection(msg)

	case message.BrowserOpenFailed:
		m.browserErr = msg.Err

	case message.EnteringReaderMode:
		return m, m.handleEnteringReaderMode(msg)

	case message.ArticleReady:
		if msg.FetchID != m.fetchID {
			return m, nil
		}

		m.pager.transition = nil
		m.status.StopSpinner()

		if msg.Err != nil {
			m.state = stateBrowsing

			return m, m.status.NewStatusMessageWithDuration(friendlyError(msg.Err), statusMessageLong)
		}

		m.readerView = reader.NewWithArticle(msg.Parsed, msg.Title, m.config.ArticleWidth, m.width, m.height, reader.Meta{
			URL:       msg.URL,
			Author:    msg.Author,
			TimeAgo:   msg.TimeAgo,
			ID:        msg.ID,
			Points:    msg.Points,
			NerdFonts: m.config.EnableNerdFonts,
		})

		m.state = stateReaderView

		initCmd := m.readerView.Init()
		if msg.HistoryWarning != nil {
			return m, tea.Batch(initCmd,
				m.status.NewStatusMessageWithDuration("Could not save read status", statusMessageShort))
		}

		return m, initCmd

	case message.ReaderViewQuit:
		m.readerView = nil
		m.state = stateBrowsing

		return m, nil

	case message.FetchAndChangeToCategory:
		return m, m.fetchCategory(msg.Category, msg.Index, msg.Cursor)

	case message.Refresh:
		return m, tea.Batch(m.fetchCategory(msg.CurrentCategory, msg.CurrentIndex, 0), fetchMemorialStatus())

	case message.ShowStatusMessage:
		cmds = append(cmds, m.status.NewStatusMessageWithDuration(msg.Message, msg.Duration))

	case message.CommentTreeDataReady:
		return m.handleCommentTreeDataReady(msg)

	case message.CommentViewQuit:
		m.commentView = nil
		m.state = stateBrowsing

		return m, nil

	case message.CategoryFetchingFinished:
		return m.handleCategoryFetchingFinished(msg)
	}

	if m.state == stateReaderView {
		return m, m.readerView.Update(msg)
	}

	if m.state == stateCommentView {
		return m, m.commentView.Update(msg)
	}

	if m.state == stateHelpScreen {
		return m.updateHelpScreen(msg)
	}

	cmds = append(cmds, m.handleBrowsing(msg))

	return m, tea.Batch(cmds...)
}

func favItemsToStories(items []*favorites.Item) []*hn.Story {
	stories := make([]*hn.Story, len(items))

	for i, it := range items {
		stories[i] = &hn.Story{
			ID:            it.ID,
			Title:         it.Title,
			Points:        it.Points,
			Author:        it.Author,
			Time:          it.Time,
			URL:           it.URL,
			Domain:        it.Domain,
			CommentsCount: it.CommentsCount,
		}
	}

	return stories
}
