package view

import (
	"time"

	"github.com/bensadeh/circumflex/view/message"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
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

func (s *statusBar) SetPermanentStatusMessage(msg string) {
	s.message = msg
}

func (s *statusBar) hideStatusMessage() {
	s.message = ""
	s.generation++
}

func (s *statusBar) StartSpinner() tea.Cmd {
	s.spinner = newSpinner()
	s.showSpinner = true

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
