package comments

import (
	"fmt"
	"os"

	"github.com/bensadeh/circumflex/comment"
	"github.com/bensadeh/circumflex/view/message"

	tea "charm.land/bubbletea/v2"
)

// standaloneModel adapts Model to a self-contained Bubble Tea program for the
// comments subcommand. The view is created on the first WindowSizeMsg because
// Model needs real dimensions at construction.
type standaloneModel struct {
	view *Model

	thread          *comment.Thread
	lastVisited     int64
	commentWidth    int
	indent          int
	enableNerdFonts bool

	browserErr error
}

func (m standaloneModel) Init() tea.Cmd {
	return nil
}

func (m standaloneModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok && msg.Mod == tea.ModCtrl && msg.Code == 'c' {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case message.BrowserOpenFailed:
		m.browserErr = msg.Err

	case tea.WindowSizeMsg:
		if m.view == nil {
			m.view = New(m.thread, m.lastVisited, m.commentWidth, m.indent, m.enableNerdFonts, msg.Width, msg.Height)
			m.view.DisableStoryNavigation()

			return m, m.view.Init()
		}

	case message.CommentViewQuit:
		return m, tea.Quit
	}

	if m.view == nil {
		return m, nil
	}

	return m, m.view.Update(msg)
}

func (m standaloneModel) View() tea.View {
	if m.view == nil {
		return tea.NewView("")
	}

	v := tea.NewView(m.view.View())
	v.AltScreen = true

	return v
}

func Run(thread *comment.Thread, lastVisited int64, commentWidth, indent int, enableNerdFonts bool) error {
	p := tea.NewProgram(standaloneModel{
		thread:          thread,
		lastVisited:     lastVisited,
		commentWidth:    commentWidth,
		indent:          indent,
		enableNerdFonts: enableNerdFonts,
	})

	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	if sm, ok := finalModel.(standaloneModel); ok && sm.browserErr != nil {
		fmt.Fprintf(os.Stderr, "circumflex: could not open browser: %v\n", sm.browserErr)
	}

	return nil
}
