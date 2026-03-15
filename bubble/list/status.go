package list

import (
	"clx/bubble/list/message"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type statusBar struct {
	message      string
	messageTimer *time.Timer
	lifetime     time.Duration
	spinner      spinner.Model
	showSpinner  bool
}

func (s *statusBar) NewStatusMessage(msg string) tea.Cmd {
	s.message = msg
	if s.messageTimer != nil {
		s.messageTimer.Stop()
	}

	s.messageTimer = time.NewTimer(s.lifetime)

	return func() tea.Msg {
		<-s.messageTimer.C

		return message.StatusMessageTimeout{}
	}
}

func (s *statusBar) NewStatusMessageWithDuration(msg string, d time.Duration) tea.Cmd {
	s.message = lipgloss.NewStyle().Render(msg)

	if s.messageTimer != nil {
		s.messageTimer.Stop()
	}

	s.messageTimer = time.NewTimer(d)

	return func() tea.Msg {
		<-s.messageTimer.C

		return message.StatusMessageTimeout{}
	}
}

func (s *statusBar) SetPermanentStatusMessage(msg string, faint bool) {
	s.message = lipgloss.NewStyle().
		Faint(faint).
		Render(msg)
}

func (s *statusBar) hideStatusMessage() {
	s.message = ""
	if s.messageTimer != nil {
		s.messageTimer.Stop()
	}
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
