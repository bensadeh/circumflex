package pane

import (
	"os"
	"strconv"
	"strings"

	"github.com/bensadeh/circumflex/style"

	tea "charm.land/bubbletea/v2"
)

// DetectStyledUnderline resolves whether the terminal can draw the dashed
// underline inert reader links use. iTerm2, Alacritty and VTE-based
// terminals (0.52+) draw styled underlines but do not answer XTGETTCAP, so
// their environment markers enable it directly. Apple's Terminal.app prints
// DCS queries to the screen instead of consuming them, so it is never
// queried. Every other terminal is asked for the Smulx terminfo capability
// and affirms it with a CapabilityMsg; no answer leaves the plain underline.
func DetectStyledUnderline() tea.Cmd {
	termProgram := os.Getenv("TERM_PROGRAM")

	vte, _ := strconv.Atoi(os.Getenv("VTE_VERSION"))

	if termProgram == "iTerm.app" || strings.HasPrefix(os.Getenv("TERM"), "alacritty") || vte >= 5200 {
		style.EnableDashedUnderline()

		return nil
	}

	if termProgram == "Apple_Terminal" {
		return nil
	}

	return tea.RequestCapability("Smulx")
}
