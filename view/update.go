package view

import (
	"github.com/bensadeh/circumflex/favorites"
	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/view/comments"
	"github.com/bensadeh/circumflex/view/message"
	"github.com/bensadeh/circumflex/view/reader"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

func (m *model) Update(msg tea.Msg) (*model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok && msg.Mod == tea.ModCtrl && msg.Code == 'c' {
		if m.state == stateFetching {
			return m, m.handleCancelFetch()
		}

		return m, tea.Quit
	}

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
		return m.handleArticleReady(msg)

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

	case message.OpenAdjacentStory:
		return m, m.handleOpenAdjacentStory(msg)

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

func (m *model) handleStartup(msg tea.WindowSizeMsg) (*model, tea.Cmd) {
	m.setSize(msg.Width, msg.Height)

	var cmds []tea.Cmd

	spinnerCmd := m.startFetch(0)
	cmds = append(cmds, spinnerCmd)

	m.syncFavorites()

	fetchCmd := m.fetchStoriesForFirstCategory()
	cmds = append(cmds, fetchCmd)
	cmds = append(cmds, fetchMemorialStatus())
	cmds = append(cmds, scheduleTimeRefresh())

	m.helpViewport = viewport.New()
	m.resizeHelpViewport(msg.Width, msg.Height)

	return m, tea.Batch(cmds...)
}

func (m *model) handleWindowResize(msg tea.WindowSizeMsg) (*model, tea.Cmd) {
	m.setSize(msg.Width, msg.Height)

	// The detail views are sized to their pane, which is the full screen when
	// the terminal is too narrow for the wide layout.
	detailMsg := tea.WindowSizeMsg{Width: m.detailWidth(), Height: msg.Height}

	if m.state == stateReaderView {
		return m, m.readerView.Update(detailMsg)
	}

	if m.state == stateCommentView {
		return m, m.commentView.Update(detailMsg)
	}

	m.resizeHelpViewport(msg.Width, msg.Height)

	return m, nil
}

func (m *model) handleFetchingFinished(msg message.FetchingFinished) (*model, tea.Cmd) {
	if msg.FetchID != m.fetchID {
		return m, nil
	}

	m.status.StopSpinner()
	m.state = stateBrowsing

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

	if msg.Err != nil {
		m.list.RollbackTransition()
		m.state = stateBrowsing
		m.status.StopSpinner()
		m.updatePagination()

		return m, m.status.NewStatusMessageWithDuration(friendlyError(msg.Err), statusMessageLong)
	}

	m.list.EndTransition()
	m.list.SetItems(msg.Category, msg.Stories)
	m.list.SetPage(0)
	m.state = stateBrowsing
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

	var cmds []tea.Cmd

	m.list.EndTransition()
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

	m.commentView = comments.New(msg.Thread, msg.LastVisited, m.config.CommentWidth, m.config.Indent, m.config.EnableNerdFonts, m.detailWidth(), m.height)
	m.state = stateCommentView

	cmds = append(cmds, m.commentView.Init())

	if msg.HistoryWarning != nil {
		cmds = append(cmds, m.status.NewStatusMessageWithDuration("Could not save read status", statusMessageShort))
	}

	return m, tea.Batch(cmds...)
}

func (m *model) handleArticleReady(msg message.ArticleReady) (*model, tea.Cmd) {
	if msg.FetchID != m.fetchID {
		return m, nil
	}

	m.list.EndTransition()
	m.status.StopSpinner()

	if msg.Err != nil {
		m.state = stateBrowsing

		return m, m.status.NewStatusMessageWithDuration(friendlyError(msg.Err), statusMessageLong)
	}

	m.readerView = reader.NewWithArticle(msg.Parsed, msg.Title, m.config.ArticleWidth, m.detailWidth(), m.height, reader.Meta{
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
}
