package pane

import (
	"fmt"
	"os"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/graphics"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/style"
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

	// title names the terminal window for as long as the program runs. It is
	// the page the subcommand was invoked on and stays put through followed
	// links, as the full app keeps the open story's title while the reader
	// walks a trail.
	title string

	// bgMsg and fgMsg hold terminal color reports that arrived before the
	// view existed, replayed once the view is created.
	bgMsg tea.Msg
	fgMsg tea.Msg

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
	// The background feeds image transparency in reader mode, the foreground
	// its URL selector's separator row; terminals that do not answer simply
	// never deliver the messages.
	return tea.Batch(tea.RequestBackgroundColor, tea.RequestForegroundColor,
		DetectStyledUnderline(), DetectGraphics())
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

	// Unlike the full app, the article may already be on screen when the
	// graphics probe answers; the forward lets it re-render with
	// high-resolution images.
	if HandleGraphicsReport(msg) {
		if s.view != nil {
			return s, s.view.Update(message.GraphicsChanged{})
		}

		return s, nil
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
		// Mirrors the app's rule: an in-flight fetch owns the progress
		// indicator, so only settle it when none is running.
		if s.status.Expire(msg.Generation) && !s.fetch.InFlight() {
			ClearProgress()
		}

		return s, nil

	case message.RestorePage:
		// Walking back needs no fetch: the entry carries its parse. Dropped
		// mid-fetch — minted a cycle after its keypress, it slips past the
		// in-flight key gate below — mirroring the app shell.
		if s.makePageView != nil && !s.fetch.InFlight() {
			var cmd tea.Cmd

			s.view, cmd = s.buildPage(msg.Entry, msg.Trail)

			return s, cmd
		}

	case tea.CapabilityMsg:
		// Unlike the full app, the page may already be on screen when the
		// Smulx answer arrives; the forward lets it repaint with dashed
		// links. The flag is global, so nothing needs replaying into a
		// view built later.
		if style.NoteTerminalCapability(msg.Content) && s.view != nil {
			return s, s.view.Update(msg)
		}

		return s, nil

	case tea.BackgroundColorMsg:
		s.bgMsg = msg // forwarded below, or replayed if the view is not built yet

	case tea.ForegroundColorMsg:
		s.fgMsg = msg // forwarded below, or replayed if the view is not built yet

	case tea.WindowSizeMsg:
		grew := msg.Width > s.width
		s.width, s.height = msg.Width, msg.Height

		// A resize may really be a font-size change; image geometry follows
		// the cell's pixel shape, so re-ask while a graphics terminal is
		// attached.
		cellSizeCmd := QueryCellSize()

		if s.view == nil {
			s.view = s.makeView(msg.Width, msg.Height)
			s.replayColorReports(s.view)

			return s, tea.Batch(cellSizeCmd, s.view.Init())
		}

		if grew {
			return s, tea.Batch(RepaintAfterGrow(), cellSizeCmd, s.view.Update(msg))
		}

		return s, tea.Batch(cellSizeCmd, s.view.Update(msg))
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
	// Minted a cycle after its keypress, so a rapid second press slips past
	// the in-flight key gate and lands here mid-fetch. Dropped, like every
	// other input during a fetch.
	if s.fetch.InFlight() {
		return s, nil
	}

	// A Hacker News discussion link has no native view in the standalone
	// shells — only the full app can build a comment section — so the
	// browser stands in for it here.
	if _, ok := hn.ParseItemURL(msg.URL); ok {
		return s, message.OpenInBrowser(msg.URL)
	}

	if s.makePageView == nil {
		return s, message.OpenInBrowser(msg.URL)
	}

	// The selector greys out links that fail validation, so this guard is
	// mostly a backstop.
	if err := article.ValidateURL(msg.URL); err != nil {
		return s, s.status.Set(FriendlyError(err), StatusMessageLong)
	}

	tok := s.fetch.Begin(ReaderFetchTimeout)
	s.spinner = NewSpinner()
	s.view.Update(message.LinkFetchStatus{InFlight: true})

	SetProgressIndeterminate()

	return s, tea.Batch(s.spinner.Tick, FetchPage(tok.Ctx, tok.ID, msg.URL, msg.Trail))
}

func (s standalone) cancelFetch() (tea.Model, tea.Cmd) {
	s.fetch.Abort()
	ClearProgress()
	s.view.Update(message.LinkFetchStatus{InFlight: false})

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
		s.view.Update(message.LinkFetchStatus{InFlight: false})

		return s, s.status.Set(FriendlyError(msg.Err), StatusMessageLong)
	}

	entry := message.TrailEntry{URL: msg.URL, Title: msg.Title, Parsed: msg.Parsed}

	var cmd tea.Cmd

	s.view, cmd = s.buildPage(entry, msg.Trail)

	return s, cmd
}

func (s standalone) buildPage(entry message.TrailEntry, trail []message.TrailEntry) (View, tea.Cmd) {
	view := s.makePageView(entry, trail, s.width, s.height)
	s.replayColorReports(view)

	return view, view.Init()
}

func (s standalone) replayColorReports(view View) {
	if s.bgMsg != nil {
		view.Update(s.bgMsg)
	}

	if s.fgMsg != nil {
		view.Update(s.fgMsg)
	}
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
	v.WindowTitle = s.title

	return v
}

// RunStandalone runs a detail view as its own program under the window title
// title; makeView receives the terminal dimensions from the first
// WindowSizeMsg. A non-nil makePageView lets links followed inside the view
// open in place; without one they fall back to the browser.
func RunStandalone(title string, makeView func(width, height int) View, makePageView MakePageView) error {
	p := tea.NewProgram(standalone{title: WindowTitle(title), makeView: makeView, makePageView: makePageView})

	restoreTitle := SaveWindowTitle()
	settleProgress := WireProgress(p)
	stopGraphics := WireGraphics(p)

	finalModel, err := p.Run()

	settleProgress()
	stopGraphics()
	restoreTitle()

	// Transmitted images survive the program in the terminal's memory;
	// release them now that no frame flush can interleave with the write.
	if seq := graphics.CleanupSeq(); seq != "" {
		_, _ = fmt.Fprint(os.Stdout, seq)
	}

	if err != nil {
		return err
	}

	if sm, ok := finalModel.(standalone); ok && sm.browserErr != nil {
		fmt.Fprintf(os.Stderr, "circumflex: could not open browser: %v\n", sm.browserErr)
	}

	return nil
}
