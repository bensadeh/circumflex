package list

import (
	clxspinner "clx/spinner"
	"clx/view/message"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type statusBar struct {
	message     string
	generation  int
	spinner     spinner.Model
	showSpinner bool
}

func (s *statusBar) NewStatusMessageWithDuration(msg string, d time.Duration) tea.Cmd {
	s.message = msg
	s.generation++

	gen := s.generation

	return tea.Tick(d, func(time.Time) tea.Msg {
		return message.StatusMessageTimeout{Generation: gen}
	})
}

func (s *statusBar) SetPermanentStatusMessage(msg string, faint bool) {
	s.message = lipgloss.NewStyle().
		Faint(faint).
		Render(msg)
}

func (s *statusBar) hideStatusMessage() {
	s.message = ""
	s.generation++
}

func (s *statusBar) StartSpinner() tea.Cmd {
	s.spinner = spinner.New()
	s.spinner.Spinner = clxspinner.Random()
	s.spinner.Style = defaultStyles().Spinner

	s.showSpinner = true
	s.message = lipgloss.NewStyle().Faint(true).Render("fetching")

	return s.spinner.Tick
}

func (s *statusBar) StopSpinner() {
	s.showSpinner = false
	s.message = ""
}

func (s *statusBar) spinnerView() string {
	return s.spinner.View()
}

func scheduleTimeRefresh() tea.Cmd {
	return tea.Tick(time.Minute, func(time.Time) tea.Msg {
		return message.TimeRefreshTick{}
	})
}
