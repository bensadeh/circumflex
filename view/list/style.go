package list

import (
	"charm.land/lipgloss/v2"
)

const (
	bullet   = "•"
	ellipsis = "…"
)

type styles struct {
	Spinner lipgloss.Style

	ActivePaginationDot   lipgloss.Style
	InactivePaginationDot lipgloss.Style
}

func defaultStyles() (s styles) {
	s.Spinner = lipgloss.NewStyle()

	s.ActivePaginationDot = lipgloss.NewStyle().
		SetString(bullet)

	s.InactivePaginationDot = lipgloss.NewStyle().
		Faint(true).
		SetString(bullet)

	return s
}
