package bubble

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"os/exec"
)

var docStyle = lipgloss.NewStyle().Margin(0, 0)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

type editorFinishedMsg struct{ err error }

func openEditor() tea.Cmd {
	c := exec.Command("less", "--help")

	return tea.Exec(tea.WrapExecCommand(c), func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "enter" {
			return m, openEditor()
		}
		if msg.String() == "e" {
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
	items := []list.Item{
		item{title: "Raspberry Pi’s", desc: "I have ’em all over my house"},
		item{title: "Nutella", desc: "It's good on toast"},
		item{title: "Bitter melon", desc: "It cools you down"},
		item{title: "Nice socks", desc: "And by that I mean socks without holes"},
		item{title: "Eight hours of sleep", desc: "I had this once"},
		item{title: "Cats", desc: "Usually"},
		item{title: "Plantasia, the album", desc: "My plants love it too"},
		item{title: "Pour over coffee", desc: "It takes forever to make though"},
		item{title: "VR", desc: "Virtual reality...what is there to say?"},
		item{title: "Noguchi Lamps", desc: "Such pleasing organic forms"},
		item{title: "Linux", desc: "Pretty much the best OS"},
		item{title: "Business school", desc: "Just kidding"},
		item{title: "Pottery", desc: "Wet clay is a great feeling"},
		item{title: "Shampoo", desc: "Nothing like clean hair"},
		item{title: "Table tennis", desc: "It’s surprisingly exhausting"},
		item{title: "Milk crates", desc: "Great for packing in your extra stuff"},
		item{title: "Afternoon tea", desc: "Especially the tea sandwich part"},
		item{title: "Stickers", desc: "The thicker the vinyl the better"},
		item{title: "20° Weather", desc: "Celsius, not Fahrenheit"},
		item{title: "Warm light", desc: "Like around 2700 Kelvin"},
		item{title: "The vernal equinox", desc: "The autumnal equinox is pretty good too"},
		item{title: "Gaffer’s tape", desc: "Basically sticky fabric"},
		item{title: "Terrycloth", desc: "In other words, towel fabric"},
		item{title: "Raspberry Pi’s", desc: "I have ’em all over my house"},
		item{title: "Nutella", desc: "It's good on toast"},
		item{title: "Bitter melon", desc: "It cools you down"},
		item{title: "Nice socks", desc: "And by that I mean socks without holes"},
		item{title: "Eight hours of sleep", desc: "I had this once"},
		item{title: "Cats", desc: "Usually"},
		item{title: "Plantasia, the album", desc: "My plants love it too"},
		item{title: "Pour over coffee", desc: "It takes forever to make though"},
		item{title: "VR", desc: "Virtual reality...what is there to say?"},
		item{title: "Noguchi Lamps", desc: "Such pleasing organic forms"},
		item{title: "Linux", desc: "Pretty much the best OS"},
		item{title: "Business school", desc: "Just kidding"},
		item{title: "Pottery", desc: "Wet clay is a great feeling"},
		item{title: "Shampoo", desc: "Nothing like clean hair"},
		item{title: "Table tennis", desc: "It’s surprisingly exhausting"},
		item{title: "Milk crates", desc: "Great for packing in your extra stuff"},
		item{title: "Afternoon tea", desc: "Especially the tea sandwich part"},
		item{title: "Stickers", desc: "The thicker the vinyl the better"},
		item{title: "20° Weather", desc: "Celsius, not Fahrenheit"},
		item{title: "Warm light", desc: "Like around 2700 Kelvin"},
		item{title: "The vernal equinox", desc: "The autumnal equinox is pretty good too"},
		item{title: "Gaffer’s tape", desc: "Basically sticky fabric"},
		item{title: "Terrycloth", desc: "In other words, towel fabric"},
	}

	// Create a new default delegate
	d := list.NewDefaultDelegate()
	d.SetSpacing(0)

	// Change colors
	c := lipgloss.Color("#6f03fc")
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(c).
		UnsetPadding().
		UnsetBorderStyle().
		UnsetMargins()
	d.Styles.NormalTitle = d.Styles.NormalTitle.
		UnsetBorderStyle().
		UnsetPadding().
		UnsetMargins()
	d.Styles.NormalDesc = d.Styles.NormalTitle.Copy() // reuse the title style here

	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		UnsetBorderStyle().
		UnsetPadding()

	d.Styles.DimmedTitle = d.Styles.DimmedTitle.
		UnsetBorderStyle().
		UnsetPadding()
	//d.Styles.SelectedTitle = d.Styles.SelectedTitle.UnsetPadding()
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.BorderLeft(false)
	d.Styles.SelectedDesc = d.Styles.SelectedTitle.Copy()           // reuse the title style here
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.BorderLeft(false) // reuse the title style here

	d.Styles.NormalTitle = d.Styles.NormalTitle.BorderLeft(false)
	d.Styles.NormalDesc = d.Styles.NormalDesc.BorderLeft(false)

	m := model{list: list.New(items, d, 0, 0)}

	m.list.SetShowTitle(false)
	m.list.SetShowPagination(false)
	m.list.SetShowHelp(false)
	m.list.SetFilteringEnabled(false)
	m.list.SetShowStatusBar(false)
	m.list.Styles.Title = m.list.Styles.Title.UnsetBorderStyle().
		UnsetPadding().
		UnsetMargins()

	m.list.SetSpinner(spinner.Globe)

	p := tea.NewProgram(m, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
