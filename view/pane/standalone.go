package pane

import (
	"fmt"
	"os"

	"github.com/bensadeh/circumflex/view/message"

	tea "charm.land/bubbletea/v2"
)

// View is the surface the standalone adapter drives; both detail views
// satisfy it.
type View interface {
	Init() tea.Cmd
	Update(tea.Msg) tea.Cmd
	View() string
}

// standalone adapts a detail view to a self-contained Bubble Tea program
// for the comments/article/url subcommands. The view is created on the
// first WindowSizeMsg because the views need real dimensions at
// construction.
type standalone struct {
	makeView func(width, height int) View
	view     View
	width    int

	// bgMsg holds a background color report that arrived before the view
	// existed, replayed once the view is created.
	bgMsg tea.Msg

	browserErr error
}

func (s standalone) Init() tea.Cmd {
	// The response feeds image transparency in reader mode; terminals that
	// do not answer simply never deliver the message.
	return tea.RequestBackgroundColor
}

func (s standalone) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok && msg.Mod == tea.ModCtrl && msg.Code == 'c' {
		return s, tea.Quit
	}

	switch msg := msg.(type) {
	case message.BrowserOpenFailed:
		s.browserErr = msg.Err

	case message.CommentViewQuit, message.ReaderViewQuit:
		return s, tea.Quit

	case tea.BackgroundColorMsg:
		s.bgMsg = msg // forwarded below, or replayed if the view is not built yet

	case tea.WindowSizeMsg:
		grew := msg.Width > s.width
		s.width = msg.Width

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
	}

	if s.view == nil {
		return s, nil
	}

	return s, s.view.Update(msg)
}

func (s standalone) View() tea.View {
	if s.view == nil {
		return tea.NewView("")
	}

	v := tea.NewView(s.view.View())
	v.AltScreen = true

	return v
}

// RunStandalone runs a detail view as its own program; makeView receives
// the terminal dimensions from the first WindowSizeMsg.
func RunStandalone(makeView func(width, height int) View) error {
	finalModel, err := tea.NewProgram(standalone{makeView: makeView}).Run()
	if err != nil {
		return err
	}

	if sm, ok := finalModel.(standalone); ok && sm.browserErr != nil {
		fmt.Fprintf(os.Stderr, "circumflex: could not open browser: %v\n", sm.browserErr)
	}

	return nil
}
