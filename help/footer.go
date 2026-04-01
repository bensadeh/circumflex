package help

import (
	"clx/style"
	"clx/version"
	"strings"

	"charm.land/lipgloss/v2"
)

func Footer(width int) string {
	versionText := "github.com/bensadeh/circumflex • version " + version.Version
	textWidth := lipgloss.Width(versionText)
	leftPad := strings.Repeat(" ", max(0, (width-textWidth)/2))

	return leftPad + style.Faint(versionText)
}
