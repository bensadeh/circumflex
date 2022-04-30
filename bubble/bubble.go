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
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"os"
	"strings"
	"time"
)

const (
	gray      = "237"
	lightGray = "238"
	magenta   = "200"
	yellow    = "214"
	blue      = "33"
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
			p := termenv.ColorProfile()
			c := termenv.String(".").
				Foreground(p.Color(magenta)).
				Background(p.Color(lightGray))

			l := termenv.String(".").
				Foreground(p.Color(yellow)).
				Background(p.Color(lightGray))

			x := termenv.String(".").
				Foreground(p.Color(blue)).
				Background(p.Color(lightGray))

			filler := termenv.String(" ").
				Background(p.Color(lightGray))

			dot := spinner.Spinner{
				Frames: []string{"fetching" + strings.Repeat(filler.String(), 3),
					"fetching" + c.String() + strings.Repeat(filler.String(), 2),
					"fetching" + c.String() + l.String() + strings.Repeat(filler.String(), 1),
					"fetching" + c.String() + l.String() + x.String()},
				FPS: 600 * time.Millisecond, //nolint:gomnd
			}

			m.list.SetSpinner(dot)
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
