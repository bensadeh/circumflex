package help

import (
	"clx/style"
	"clx/version"
	"strings"

	"charm.land/lipgloss/v2"
)

func Footer(width int) string {
	logo := style.ModeIndicator(style.Logo("{", "⌨", "}"), []style.Binding{})
	logoWidth := lipgloss.Width(logo)
	versionText := "github.com/bensadeh/circumflex • version " + version.Version
	textWidth := lipgloss.Width(versionText)
	centerPos := (width - textWidth) / 2
	gap := strings.Repeat(" ", max(0, centerPos-logoWidth))

	return logo + gap + style.Faint(versionText)
}
