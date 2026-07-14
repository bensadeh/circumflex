package view

import (
	"time"

	"github.com/bensadeh/circumflex/view/message"
	"github.com/bensadeh/circumflex/view/pane"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

type statusBar struct {
	text        pane.TransientStatus
	spinner     spinner.Model
	showSpinner bool
}

func (s *statusBar) NewStatusMessageWithDuration(msg string, d time.Duration) tea.Cmd {
	return s.text.Set(msg, d)
}

func (s *statusBar) SetPermanentStatusMessage(msg string) {
	s.text.SetPermanent(msg)
}

func (s *statusBar) StartSpinner() tea.Cmd {
	s.spinner = pane.NewSpinner()
	s.showSpinner = true

	return s.spinner.Tick
}

func (s *statusBar) StopSpinner() {
	s.showSpinner = false
	s.text.SetPermanent("")
}

func (s *statusBar) spinnerView() string {
	return s.spinner.View()
}

func scheduleTimeRefresh() tea.Cmd {
	return tea.Tick(time.Minute, func(time.Time) tea.Msg {
		return message.TimeRefreshTick{}
	})
}
