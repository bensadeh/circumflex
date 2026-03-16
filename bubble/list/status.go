package list

import (
	"clx/bubble/list/message"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type statusBar struct {
	message     string
	generation  int
	lifetime    time.Duration
	spinner     spinner.Model
	showSpinner bool
}

func (s *statusBar) NewStatusMessage(msg string) tea.Cmd {
	s.message = msg
	s.generation++

	gen := s.generation

	return tea.Tick(s.lifetime, func(time.Time) tea.Msg {
		return message.StatusMessageTimeout{Generation: gen}
	})
}

func (s *statusBar) NewStatusMessageWithDuration(msg string, d time.Duration) tea.Cmd {
	s.message = lipgloss.NewStyle().Render(msg)
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
	s.spinner.Spinner = getSpinner()
	s.spinner.Style = DefaultStyles().Spinner

	s.showSpinner = true

	return s.spinner.Tick
}

func (s *statusBar) StopSpinner() {
	s.showSpinner = false
}

func (s *statusBar) spinnerView() string {
	return s.spinner.View()
}

func scheduleTimeRefresh() tea.Cmd {
	return tea.Tick(time.Minute, func(time.Time) tea.Msg {
		return message.TimeRefreshTick{}
	})
}
