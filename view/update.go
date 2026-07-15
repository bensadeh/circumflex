package view

import (
	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/favorites"
	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/meta"
	"github.com/bensadeh/circumflex/timeago"
	"github.com/bensadeh/circumflex/view/comments"
	"github.com/bensadeh/circumflex/view/message"
	"github.com/bensadeh/circumflex/view/pane"
	"github.com/bensadeh/circumflex/view/reader"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

func (m *model) Update(msg tea.Msg) (*model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok && msg.Mod == tea.ModCtrl && msg.Code == 'c' {
		if m.fetch.inFlight() {
			return m, m.handleCancelFetch()
		}

		return m, tea.Quit
	}

	// Ctrl+L forces a full repaint — the terminal convention for healing
	// artifacts the cell-diff renderer cannot see, whatever desynced the
	// terminal from its model (a torn escape sequence, glyph-width drift, a
	// multiplexer quirk).
	if msg, ok := msg.(tea.KeyPressMsg); ok && msg.Mod == tea.ModCtrl && msg.Code == 'l' {
		return m, tea.ClearScreen
	}

	// Handled before the startup gate: the terminal's answers can arrive
	// before the first WindowSizeMsg and must not be dropped.
	if bg, ok := msg.(tea.BackgroundColorMsg); ok {
		m.termBG = bg.Color

		return m, nil
	}

	if fg, ok := msg.(tea.ForegroundColorMsg); ok {
		m.termFG = fg.Color

		return m, nil
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
		var cmd tea.Cmd

		m.status.spinner, cmd = pane.UpdateSpinner(m.status.spinner, msg, m.status.showSpinner)
		cmds = append(cmds, cmd)

	case message.TimeRefreshTick:
		cmds = append(cmds, scheduleTimeRefresh())

	case message.MemorialStatusReady:
		// Only apply a definitive result; on error leave the current state so a
		// failed re-check (e.g. during a refresh) never blanks an active bar.
		if msg.Err == nil {
			header.SetMemorial(msg.Active)
		}

		m.memorialErr = msg.Err

	case message.StatusMessageTimeout:
		if m.status.text.Expire(msg.Generation) {
			pane.ClearProgress()
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

	case message.BrowserOpenFailed:
		m.browserErr = msg.Err

	case message.ArticleReady:
		return m.handleArticleReady(msg)

	case message.OpenReaderLink:
		return m, m.handleOpenReaderLink(msg)

	case message.RestoreReaderPage:
		return m.handleRestoreReaderPage(msg)

	case message.LinkArticleReady:
		return m.handleLinkArticleReady(msg)

	case message.DetailQuit:
		m.detail = nil
		m.screen = screenList

		return m, nil

	case message.ErrorViewQuit:
		m.detail = nil
		m.screen = screenList
		// Settle the terminal progress indicator in case the view is quit
		// before its timeout fires.
		pane.ClearProgress()

		return m, nil

	case message.ErrorProgressTimeout:
		if msg.FetchID == m.fetch.currentID() {
			pane.ClearProgress()
		}

	case message.CommentTreeDataReady:
		return m.handleCommentTreeDataReady(msg)

	case message.OpenAdjacentStory:
		return m, m.handleOpenAdjacentStory(msg)

	case message.StoriesReady:
		return m.handleStoriesReady(msg)
	}

	// While a fetch is in flight only the cancel key acts, whatever the
	// screen; other keys would race the fetch's outcome.
	if m.fetch.inFlight() {
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

	tok, spinnerCmd := m.startFetch(0, m.listRollback())

	pane.SetProgressIndeterminate()

	cmds = append(cmds, spinnerCmd)

	m.syncFavorites()

	fetchCmd := m.fetchStoriesForFirstCategory(tok)
	cmds = append(cmds, fetchCmd)
	cmds = append(cmds, fetchMemorialStatus())
	cmds = append(cmds, scheduleTimeRefresh())

	m.helpViewport = viewport.New()
	m.resizeHelpViewport()

	return m, tea.Batch(cmds...)
}

func (m *model) handleWindowResize(msg tea.WindowSizeMsg) (*model, tea.Cmd) {
	var cmds []tea.Cmd
	if msg.Width > m.width {
		cmds = append(cmds, pane.RepaintAfterGrow())
	}

	m.setSize(msg.Width, msg.Height)

	// Resize the help viewport unconditionally: a resize that arrives while a
	// story is open must not leave help laid out for the old dimensions.
	m.resizeHelpViewport()

	// The detail view is sized to its pane, which is the full screen when
	// the terminal is too narrow for the wide layout.
	if m.detail != nil {
		cmds = append(cmds, m.detail.Update(tea.WindowSizeMsg{Width: m.detailWidth(), Height: msg.Height}))
	}

	return m, tea.Batch(cmds...)
}

func (m *model) handleStoriesReady(msg message.StoriesReady) (*model, tea.Cmd) {
	rb, ok := m.finishFetch(msg.FetchID, msg.Err)
	if !ok {
		return m, nil
	}

	if msg.Err != nil {
		m.rollbackFetch(rb)
		m.updatePagination()

		return m, m.status.NewStatusMessageWithDuration(pane.FriendlyError(msg.Err), statusMessageLong)
	}

	m.list.EndTransition()
	m.list.SetItems(msg.Category, msg.Stories)
	m.list.SetPage(0)
	m.cat.SetIndex(msg.Index)
	m.list.SetCursorClamped(msg.Cursor)

	m.updatePagination()

	return m, nil
}

func (m *model) handleCommentTreeDataReady(msg message.CommentTreeDataReady) (*model, tea.Cmd) {
	rb, ok := m.finishFetch(msg.FetchID, msg.Err)
	if !ok {
		return m, nil
	}

	var cmds []tea.Cmd

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
			m.rollbackFetch(rb)
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

// handleOpenReaderLink starts the fetch for a link followed inside an
// article. Validation failures never leave the article: they surface on the
// status bar like a failed fetch would. The selector greys out links that
// fail this check, so this guard is mostly a backstop.
func (m *model) handleOpenReaderLink(msg message.OpenReaderLink) tea.Cmd {
	if err := article.ValidateURL(msg.URL); err != nil {
		return m.status.NewStatusMessageWithDuration(pane.FriendlyError(err), statusMessageLong)
	}

	tok, startSpinnerCmd := m.startLinkFetch(readerModeTimeout)

	pane.SetProgressIndeterminate()

	return tea.Batch(startSpinnerCmd, pane.FetchPage(tok.ctx, tok.id, msg.URL, msg.Trail))
}

// handleRestoreReaderPage steps back to a page whose parse rode along the
// walk-back chain: no fetch, no spinner — the reader rebuilds synchronously.
// The story article at the chain's root gets its story meta back, rebuilt
// from the selection, which still points at it.
func (m *model) handleRestoreReaderPage(msg message.RestoreReaderPage) (*model, tea.Cmd) {
	it := m.list.SelectedItem()

	storyHeader := meta.ReaderMode(meta.Data{
		URL:           it.URL,
		Author:        it.Author,
		TimeAgo:       timeago.RelativeTime(it.Time),
		Points:        it.Points,
		CommentsCount: it.CommentsCount,
		NerdFonts:     m.config.EnableNerdFonts,
	})

	m.detail = reader.NewPage(msg.Entry, msg.Trail, storyHeader.Render,
		m.config.ArticleWidth, m.detailWidth(), m.height, m.linkPageOptions(it.ID))

	m.screen = screenReader

	return m, m.detail.Init()
}

// linkPageOptions carries the app's display knobs into a followed-link page;
// the page's own identity fields are reader.NewPage's business.
func (m *model) linkPageOptions(storyID int) reader.Options {
	return reader.Options{
		ID:        storyID,
		NerdFonts: m.config.EnableNerdFonts,
		Images:    m.config.EnableImages,
		TermBG:    m.termBG,
		TermFG:    m.termFG,
	}
}

// handleLinkArticleReady swaps the followed link's page into the detail pane
// in place of the article it was found in. On error nothing transitions —
// the open article stays, the failure surfaces on the status bar, and the
// progress indicator settles when the message expires.
func (m *model) handleLinkArticleReady(msg message.LinkArticleReady) (*model, tea.Cmd) {
	if _, ok := m.finishFetch(msg.FetchID, msg.Err); !ok {
		return m, nil
	}

	if msg.Err != nil {
		return m, m.status.NewStatusMessageWithDuration(pane.FriendlyError(msg.Err), statusMessageLong)
	}

	entry := message.TrailEntry{URL: msg.URL, Title: msg.Title, Parsed: msg.Parsed}

	m.detail = reader.NewPage(entry, msg.Trail, nil,
		m.config.ArticleWidth, m.detailWidth(), m.height, m.linkPageOptions(m.list.SelectedItem().ID))

	m.screen = screenReader

	return m, m.detail.Init()
}

func (m *model) handleArticleReady(msg message.ArticleReady) (*model, tea.Cmd) {
	rb, ok := m.finishFetch(msg.FetchID, msg.Err)
	if !ok {
		return m, nil
	}

	// On error the story never opens, mirroring the comment section above:
	// the wide layout replaces the pane with the error view, the narrow
	// layout keeps the outgoing story on screen and rolls the selection back.
	if msg.Err != nil {
		if !m.isWide() {
			m.rollbackFetch(rb)
		}

		return m, m.showDetailError(msg.Err, screenReader)
	}

	block := meta.ReaderMode(meta.Data{
		URL:           msg.URL,
		Author:        msg.Author,
		TimeAgo:       msg.TimeAgo,
		Points:        msg.Points,
		CommentsCount: msg.CommentsCount,
		NerdFonts:     m.config.EnableNerdFonts,
	})

	m.detail = reader.NewWithArticle(msg.Parsed, msg.Title, m.config.ArticleWidth, m.detailWidth(), m.height, reader.Options{
		URL:       msg.URL,
		ID:        msg.ID,
		NerdFonts: m.config.EnableNerdFonts,
		Images:    m.config.EnableImages,
		TermBG:    m.termBG,
		TermFG:    m.termFG,
	}, block.Render)

	m.screen = screenReader

	initCmd := m.detail.Init()
	if msg.HistoryWarning != nil {
		return m, tea.Batch(initCmd,
			m.status.NewStatusMessageWithDuration("Could not save read status", statusMessageShort))
	}

	return m, initCmd
}
