package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/bensadeh/circumflex/comment"
	"github.com/bensadeh/circumflex/settings"
	"github.com/bensadeh/circumflex/style"
	"github.com/bensadeh/circumflex/view/comments"
	"github.com/bensadeh/circumflex/view/message"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
)

// commentModel wraps comments.Model so it can be used as a standalone Bubble Tea program.
type commentModel struct {
	view        *comments.Model
	ready       bool
	thread      *comment.Thread
	lastVisited int64
	config      *settings.Config
	browserErr  error
}

func (m commentModel) Init() tea.Cmd {
	return nil
}

func (m commentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok && msg.Mod == tea.ModCtrl && msg.Code == 'c' {
		return m, tea.Quit
	}

	if failed, ok := msg.(message.BrowserOpenFailed); ok {
		m.browserErr = failed.Err
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			m.ready = true

			m.view = comments.New(m.thread, m.lastVisited, m.config.CommentWidth, m.config.Indent, m.config.EnableNerdFonts, msg.Width, msg.Height)
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

func (m commentModel) View() tea.View {
	if m.view == nil {
		return tea.NewView("")
	}

	v := tea.NewView(lipgloss.NewStyle().Render(m.view.View()))
	v.AltScreen = true

	return v
}

func commentsCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "comments [id]",
		Short:                 "read the comment section of a story",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				return err
			}

			config, err := getConfig()
			if err != nil {
				return err
			}

			style.SetTheme(config.Theme)

			service := newService()

			tree, err := service.FetchComments(cmd.Context(), id, nil)
			if err != nil {
				return err
			}

			m := commentModel{
				thread:      comment.ToThread(tree),
				lastVisited: time.Now().Unix(),
				config:      config,
			}

			p := tea.NewProgram(m)

			finalModel, err := p.Run()
			if err != nil {
				return err
			}

			if cm, ok := finalModel.(commentModel); ok && cm.browserErr != nil {
				fmt.Fprintf(os.Stderr, "circumflex: could not open browser: %v\n", cm.browserErr)
			}

			return nil
		},
	}
}
