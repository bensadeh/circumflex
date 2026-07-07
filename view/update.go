package view

import (
	"github.com/bensadeh/circumflex/favorites"
	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/view/comments"
	"github.com/bensadeh/circumflex/view/message"
	"github.com/bensadeh/circumflex/view/reader"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

func (m *model) Update(msg tea.Msg) (*model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok && msg.Mod == tea.ModCtrl && msg.Code == 'c' {
		if m.fetching {
			return m, m.handleCancelFetch()
		}

		clearProgress()

		return m, tea.Quit
	}

	if !m.started {
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
		return m.handleArticleReady(msg)

	case message.ReaderViewQuit, message.CommentViewQuit:
		m.detail = nil
		m.screen = screenList

		return m, nil

	case message.ErrorViewQuit:
		m.detail = nil
		m.screen = screenList
		// Settle the terminal progress indicator in case the view is quit
		// before its timeout fires.
		clearProgress()

		return m, nil

	case message.ErrorProgressTimeout:
		if msg.FetchID == m.fetchID {
			clearProgress()
		}

	case message.FetchAndChangeToCategory:
		return m, m.fetchCategory(msg.Category, msg.Index, msg.Cursor)

	case message.Refresh:
		return m, tea.Batch(m.fetchCategory(msg.CurrentCategory, msg.CurrentIndex, 0), fetchMemorialStatus())

	case message.ShowStatusMessage:
		cmds = append(cmds, m.status.NewStatusMessageWithDuration(msg.Message, msg.Duration))

	case message.CommentTreeDataReady:
		return m.handleCommentTreeDataReady(msg)

	case message.OpenAdjacentStory:
		return m, m.handleOpenAdjacentStory(msg)

	case message.CategoryFetchingFinished:
		return m.handleCategoryFetchingFinished(msg)
	}

	// While a fetch is in flight only the cancel key acts, whatever the
	// screen; other keys would race the fetch's outcome.
	if m.fetching {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(keyMsg, m.keymap.Cancel) {
			cmds = append(cmds, m.handleCancelFetch())
		}

		return m, tea.Batch(cmds...)
	}

	// Route to the active view through a single exit so cmds gathered above
	// (spinner ticks, the time-refresh reschedule) always survive delegation.
	switch {
	case m.detail != nil:
		cmds = append(cmds, m.detail.Update(msg))

	case m.screen == screenHelp:
		cmds = append(cmds, m.updateHelpScreen(msg))

	default:
		cmds = append(cmds, m.handleBrowsing(msg))
	}

	return m, tea.Batch(cmds...)
}

func (m *model) handleStartup(msg tea.WindowSizeMsg) (*model, tea.Cmd) {
	m.started = true
	m.setSize(msg.Width, msg.Height)

	var cmds []tea.Cmd

	spinnerCmd := m.startFetch(0)

	setProgressIndeterminate()

	cmds = append(cmds, spinnerCmd)

	m.syncFavorites()

	fetchCmd := m.fetchStoriesForFirstCategory()
	cmds = append(cmds, fetchCmd)
	cmds = append(cmds, fetchMemorialStatus())
	cmds = append(cmds, scheduleTimeRefresh())

	m.helpViewport = viewport.New()
	m.resizeHelpViewport()

	return m, tea.Batch(cmds...)
}

func (m *model) handleWindowResize(msg tea.WindowSizeMsg) (*model, tea.Cmd) {
	m.setSize(msg.Width, msg.Height)

	// Resize the help viewport unconditionally: a resize that arrives while a
	// story is open must not leave help laid out for the old dimensions.
	m.resizeHelpViewport()

	// The detail view is sized to its pane, which is the full screen when
	// the terminal is too narrow for the wide layout.
	if m.detail != nil {
		return m, m.detail.Update(tea.WindowSizeMsg{Width: m.detailWidth(), Height: msg.Height})
	}

	return m, nil
}

func (m *model) handleFetchingFinished(msg message.FetchingFinished) (*model, tea.Cmd) {
	if msg.FetchID != m.fetchID {
		return m, nil
	}

	syncProgress(msg.Err)
	m.status.StopSpinner()
	m.fetching = false

	if msg.Err != nil {
		return m, m.status.NewStatusMessageWithDuration(friendlyError(msg.Err), statusMessageLong)
	}

	m.list.SetItems(msg.Category, msg.Stories)
	m.updatePagination()

	return m, nil
}

func (m *model) handleCategoryFetchingFinished(msg message.CategoryFetchingFinished) (*model, tea.Cmd) {
	if msg.FetchID != m.fetchID {
		return m, nil
	}

	syncProgress(msg.Err)

	m.fetching = false

	if msg.Err != nil {
		m.rollbackFetch()
		m.status.StopSpinner()
		m.updatePagination()

		return m, m.status.NewStatusMessageWithDuration(friendlyError(msg.Err), statusMessageLong)
	}

	m.list.EndTransition()
	m.list.SetItems(msg.Category, msg.Stories)
	m.list.SetPage(0)
	m.status.StopSpinner()
	m.cat.SetIndex(msg.Index)
	m.list.SetCursorClamped(msg.Cursor)

	m.updatePagination()

	return m, nil
}

func (m *model) handleCommentTreeDataReady(msg message.CommentTreeDataReady) (*model, tea.Cmd) {
	if msg.FetchID != m.fetchID {
		return m, nil
	}

	syncProgress(msg.Err)

	m.fetching = false

	var cmds []tea.Cmd

	m.status.StopSpinner()

	if msg.UpdatedStory != nil {
		if err := m.favorites.UpdateStoryAndWriteToDisk(favorites.ItemFromStory(msg.UpdatedStory)); err != nil {
			cmds = append(cmds, m.status.NewStatusMessageWithDuration("Could not update favorite on disk", statusMessageLong))
		}
	}

	// On error the story never opens. The wide layout replaces the pane with
	// the error view, its reading marker on the story that failed; the narrow
	// layout keeps the outgoing story on screen, so the selection moves back
	// to match it.
	if msg.Err != nil {
		if !m.isWide() {
			m.rollbackFetch()
		}

		cmds = append(cmds, m.showDetailError(msg.Err, screenComments))

		return m, tea.Batch(cmds...)
	}

	m.detail = comments.New(msg.Thread, msg.LastVisited, m.config.CommentWidth, m.config.Indent, m.config.EnableNerdFonts, m.detailWidth(), m.height)
	m.screen = screenComments

	cmds = append(cmds, m.detail.Init())

	if msg.HistoryWarning != nil {
		cmds = append(cmds, m.status.NewStatusMessageWithDuration("Could not save read status", statusMessageShort))
	}

	return m, tea.Batch(cmds...)
}

func (m *model) handleArticleReady(msg message.ArticleReady) (*model, tea.Cmd) {
	if msg.FetchID != m.fetchID {
		return m, nil
	}

	syncProgress(msg.Err)

	m.fetching = false
	m.status.StopSpinner()

	// On error the story never opens, mirroring the comment section above:
	// the wide layout replaces the pane with the error view, the narrow
	// layout keeps the outgoing story on screen and rolls the selection back.
	if msg.Err != nil {
		if !m.isWide() {
			m.rollbackFetch()
		}

		return m, m.showDetailError(msg.Err, screenReader)
	}

	m.detail = reader.NewWithArticle(msg.Parsed, msg.Title, m.config.ArticleWidth, m.detailWidth(), m.height, reader.Meta{
		URL:       msg.URL,
		Author:    msg.Author,
		TimeAgo:   msg.TimeAgo,
		ID:        msg.ID,
		Points:    msg.Points,
		NerdFonts: m.config.EnableNerdFonts,
	})

	m.screen = screenReader

	initCmd := m.detail.Init()
	if msg.HistoryWarning != nil {
		return m, tea.Batch(initCmd,
			m.status.NewStatusMessageWithDuration("Could not save read status", statusMessageShort))
	}

	return m, initCmd
}
