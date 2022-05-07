package bubble

import (
	"clx/bubble/list"
	"clx/cli"
	"clx/comment"
	"clx/core"
	"clx/history"
	"clx/hn/services/mock"
	"clx/screen"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"time"
)

var docStyle = lipgloss.NewStyle()

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

type editorFinishedMsg struct{ err error }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.list.OnStartup() && m.list.Width() == 0 {
		m.list.SetSize(screen.GetTerminalWidth(), screen.GetTerminalHeight())

		cmd := m.list.StartSpinner()

		m.list.SetOnStartup(false)

		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "e" {
			cmd := m.list.NewStatusMessageWithDuration("Test", 2*time.Second)

			return m, cmd
		}
		if msg.String() == "f" {
			cmd := m.list.NewStatusMessageWithDuration("ABCDEF", 1*time.Second)

			return m, cmd
		}
		if msg.String() == "enter" {
			id := m.list.SelectedItem().ID
			cmd := openEditor(id)

			return m, cmd
		}
		if msg.String() == "u" {
			cmd := m.list.StartSpinner()

			return m, cmd
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func openEditor(id int) tea.Cmd {
	comments := new(mock.Service).FetchStory(id)

	screenWidth := screen.GetTerminalWidth()
	commentTree := comment.ToString(comments, core.GetConfigWithDefaults(), screenWidth, 0)

	c := cli.WrapLess(commentTree)

	return tea.Exec(tea.WrapExecCommand(c), func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func Run() {
	service := new(mock.Service)
	stories := service.FetchStories(0, 0)

	m := model{list: list.New(stories, list.NewDefaultDelegate(), history.Initialize(true), 0, 0)}
	m.list.Title = "My Fave Things"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	cli.ClearScreen()
}
