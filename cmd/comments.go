package cmd

import (
	"clx/bubble/comments"
	"clx/bubble/list/message"
	"clx/comment"
	"clx/convert"
	"fmt"
	"os"
	"strconv"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
)

// commentModel wraps comments.Model so it can be used as a standalone Bubble Tea program.
type commentModel struct {
	view   *comments.Model
	ready  bool
	thread *comment.Thread
	config commentLaunchConfig
}

type commentLaunchConfig struct {
	lastVisited int64
}

func (m commentModel) Init() tea.Cmd {
	return nil
}

func (m commentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			m.ready = true

			wmsg := msg
			m.view = comments.New(m.thread, m.config.lastVisited, getConfig(), wmsg.Width, wmsg.Height)

			return m, m.view.Init()
		}
	case message.CommentViewQuitMsg:
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
		Use:   "comments",
		Short: "Go directly to the comment section by ID",
		Long: "Directly enter the comment section for a given item without going through the main " +
			"view first",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Fprintln(os.Stderr, "Argument must be a valid ID")
				os.Exit(1)
			}

			service := newService()

			story, err := service.FetchComments(cmd.Context(), id)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			thread := convert.StoryToThread(story)

			m := commentModel{
				thread: thread,
				config: commentLaunchConfig{
					lastVisited: time.Now().Unix(),
				},
			}

			p := tea.NewProgram(m)
			if _, err := p.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}
}
