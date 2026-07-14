package pane

import (
	"fmt"
	"os"
	"time"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/view/message"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

// View is the surface the standalone adapter drives; both detail views
// satisfy it.
type View interface {
	Init() tea.Cmd
	Update(tea.Msg) tea.Cmd
	View() string
}

// MakePageView rebuilds a reader page around a followed link — opening one,
// or walking back through the trail. entry carries the page's parse, trail
// the chain behind it.
type MakePageView func(entry message.TrailEntry, trail []message.TrailEntry, width, height int) View

const linkFetchTimeout = 15 * time.Second

// CancelKeys stops an in-flight fetch; the app's keymap shares it so both
// shells cancel on the same keys.
var CancelKeys = key.NewBinding(key.WithKeys("esc", "backspace", "ctrl+c"))

// standalone adapts a detail view to a self-contained Bubble Tea program
// for the comments/article/url subcommands. The view is created on the
// first WindowSizeMsg because the views need real dimensions at
// construction.
type standalone struct {
	makeView     func(width, height int) View
	makePageView MakePageView
	view         View
	width        int
	height       int

	// bgMsg holds a background color report that arrived before the view
	// existed, replayed once the view is created.
	bgMsg tea.Msg

	// fetch guards the one link fetch in flight; its in-flight state doubles
	// as the signal the footer spinner shows on.
	fetch   FetchGuard
	spinner spinner.Model

	// status surfaces a failed link follow on the footer row, as the full
	// app's status bar does, then expires.
	status TransientStatus

	browserErr error
}

func (s standalone) Init() tea.Cmd {
	// The response feeds image transparency in reader mode; terminals that
	// do not answer simply never deliver the message.
	return tea.RequestBackgroundColor
}

func (s standalone) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok && msg.Mod == tea.ModCtrl && msg.Code == 'c' {
		if s.fetch.InFlight() {
			return s.cancelFetch()
		}

		return s, tea.Quit
	}

	// Ctrl+L forces a full repaint — the terminal convention for healing
	// artifacts the cell-diff renderer cannot see, whatever desynced the
	// terminal from its model.
	if msg, ok := msg.(tea.KeyPressMsg); ok && msg.Mod == tea.ModCtrl && msg.Code == 'l' {
		return s, tea.ClearScreen
	}

	switch msg := msg.(type) {
	case message.BrowserOpenFailed:
		s.browserErr = msg.Err

	case message.DetailQuit:
		return s, tea.Quit

	case message.OpenReaderLink:
		return s.followLink(msg)

	case message.LinkArticleReady:
		return s.receiveLinkedPage(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd

		s.spinner, cmd = UpdateSpinner(s.spinner, msg, s.fetch.InFlight())

		return s, cmd

	case message.StatusMessageTimeout:
		if s.status.Expire(msg.Generation) {
			ClearProgress()
		}

		return s, nil

	case message.RestoreReaderPage:
		// Walking back needs no fetch: the entry carries its parse.
		if s.makePageView != nil {
			var cmd tea.Cmd

			s.view, cmd = s.buildPage(msg.Entry, msg.Trail)

			return s, cmd
		}

	case tea.BackgroundColorMsg:
		s.bgMsg = msg // forwarded below, or replayed if the view is not built yet

	case tea.WindowSizeMsg:
		grew := msg.Width > s.width
		s.width, s.height = msg.Width, msg.Height

		if s.view == nil {
			s.view = s.makeView(msg.Width, msg.Height)

			if s.bgMsg != nil {
				s.view.Update(s.bgMsg)
			}

			return s, s.view.Init()
		}

		if grew {
			return s, tea.Batch(RepaintAfterGrow(), s.view.Update(msg))
		}

		return s, s.view.Update(msg)
	}

	if s.view == nil {
		return s, nil
	}

	// While a fetch is in flight only the cancel key acts; other keys would
	// race the fetch's outcome.
	if s.fetch.InFlight() {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(keyMsg, CancelKeys) {
			return s.cancelFetch()
		}

		return s, nil
	}

	return s, s.view.Update(msg)
}

// followLink fetches a link followed inside the article so the page can be
// swapped in place, as the full app does. A view without a page factory
// (comments) sends the link to the browser instead. The current page stays
// on screen while the fetch runs; a newer follow supersedes an older one.
func (s standalone) followLink(msg message.OpenReaderLink) (tea.Model, tea.Cmd) {
	if s.makePageView == nil {
		return s, message.OpenInBrowser(msg.URL)
	}

	// The selector greys out links that fail validation, so this guard is
	// mostly a backstop.
	if err := article.ValidateURL(msg.URL); err != nil {
		return s, s.status.Set(FriendlyError(err), StatusMessageLong)
	}

	tok := s.fetch.Begin(linkFetchTimeout)
	s.spinner = NewSpinner()

	SetProgressIndeterminate()

	return s, tea.Batch(s.spinner.Tick, FetchPage(tok.Ctx, tok.ID, msg.URL, msg.Trail))
}

func (s standalone) cancelFetch() (tea.Model, tea.Cmd) {
	s.fetch.Abort()
	ClearProgress()

	return s, s.status.Set(CancelledStatus(), StatusMessageShort)
}

// receiveLinkedPage swaps the followed link's page in for the article it was
// found in. On error nothing transitions — the open article stays and the
// failure surfaces on the footer row, as the full app's status bar does.
func (s standalone) receiveLinkedPage(msg message.LinkArticleReady) (tea.Model, tea.Cmd) {
	if !s.fetch.Finish(msg.FetchID) {
		return s, nil
	}

	SyncProgress(msg.Err)

	if msg.Err != nil {
		return s, s.status.Set(FriendlyError(msg.Err), StatusMessageLong)
	}

	entry := message.TrailEntry{URL: msg.URL, Title: msg.Title, Parsed: msg.Parsed}

	var cmd tea.Cmd

	s.view, cmd = s.buildPage(entry, msg.Trail)

	return s, cmd
}

func (s standalone) buildPage(entry message.TrailEntry, trail []message.TrailEntry) (View, tea.Cmd) {
	view := s.makePageView(entry, trail, s.width, s.height)

	if s.bgMsg != nil {
		view.Update(s.bgMsg)
	}

	return view, view.Init()
}

func (s standalone) View() tea.View {
	if s.view == nil {
		return tea.NewView("")
	}

	content := s.view.View()

	// Fetch and status feedback land on the footer row, exactly as the full
	// app overlays them on its detail views.
	switch {
	case s.fetch.InFlight():
		content = OverlayStatus(content, s.spinner.View(), s.width)
	case s.status.Message() != "":
		content = OverlayStatus(content, s.status.Message(), s.width)
	}

	v := tea.NewView(content)
	v.AltScreen = true

	return v
}

// RunStandalone runs a detail view as its own program; makeView receives
// the terminal dimensions from the first WindowSizeMsg. A non-nil
// makePageView lets links followed inside the view open in place; without
// one they fall back to the browser.
func RunStandalone(makeView func(width, height int) View, makePageView MakePageView) error {
	p := tea.NewProgram(standalone{makeView: makeView, makePageView: makePageView})

	settleProgress := WireProgress(p)

	finalModel, err := p.Run()

	settleProgress()

	if err != nil {
		return err
	}

	if sm, ok := finalModel.(standalone); ok && sm.browserErr != nil {
		fmt.Fprintf(os.Stderr, "circumflex: could not open browser: %v\n", sm.browserErr)
	}

	return nil
}
