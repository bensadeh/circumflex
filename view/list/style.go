package list

import (
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

const (
	bullet   = "•"
	ellipsis = "…"
)

type styles struct {
	Spinner lipgloss.Style

	NoItems lipgloss.Style

	ActivePaginationDot   lipgloss.Style
	InactivePaginationDot lipgloss.Style
}

func defaultStyles() (s styles) {
	s.Spinner = lipgloss.NewStyle()

	s.NoItems = lipgloss.NewStyle().
		Foreground(compat.AdaptiveColor{Light: lipgloss.Color("#909090"), Dark: lipgloss.Color("#626262")})

	s.ActivePaginationDot = lipgloss.NewStyle().
		SetString(bullet)

	s.InactivePaginationDot = lipgloss.NewStyle().
		Faint(true).
		SetString(bullet)

	return s
}
