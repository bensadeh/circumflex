package bubble

import (
	"clx/bubble/list"
	"clx/hn/services/mock"
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"time"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	title, desc, url string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) URL() string         { return i.url }

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

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
		if msg.String() == "u" {
			dot := spinner.Spinner{
				Frames: []string{"⣾ ", "⣷ ", "⣯ ", "⣟ ", "⡿ ", "⢿ ", "⣻ ", "⣽ "},
				FPS:    time.Second / 7, //nolint:gomnd
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

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func Run() {
	var items []list.Item

	service := new(mock.Service)
	stories := service.FetchStories(0, 0)

	for _, story := range stories {
		items = append(items, item{title: story.Title, desc: story.User, url: story.URL})
	}

	m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "My Fave Things"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
