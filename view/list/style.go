package list

import (
	"charm.land/lipgloss/v2"
)

const (
	bullet   = "•"
	ellipsis = "…"
)

type styles struct {
	ActivePaginationDot   lipgloss.Style
	InactivePaginationDot lipgloss.Style
}

func defaultStyles() (s styles) {
	s.ActivePaginationDot = lipgloss.NewStyle().
		SetString(bullet)

	s.InactivePaginationDot = lipgloss.NewStyle().
		Faint(true).
		SetString(bullet)

	return s
}
