package pane

import (
	"errors"
	"net"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/article"

	xansi "github.com/charmbracelet/x/ansi"

	"charm.land/lipgloss/v2"
)

var (
	redText        = lipgloss.NewStyle().Foreground(lipgloss.Red)
	statusRowStyle = lipgloss.NewStyle().Inline(true).Align(lipgloss.Center)
)

// OverlayStatus writes fetch and status feedback onto the last row of a
// detail view, which reserves that row as footer space. width is the pane
// the view fills.
func OverlayStatus(view, status string, width int) string {
	lines := strings.Split(view, "\n")
	// Width pads but never truncates, and an error message can be wider than
	// the screen.
	lines[len(lines)-1] = xansi.Truncate(statusRowStyle.Width(width).Render(status), width, "")

	return strings.Join(lines, "\n")
}

func isTimeout(err error) bool {
	var netErr net.Error

	return errors.As(err, &netErr) && netErr.Timeout()
}

func FriendlyError(err error) string {
	if isTimeout(err) {
		return "Timed out — check your connection and try again"
	}

	// Returned as-is: the generic first-letter uppercasing below would
	// mangle the leading domain ("Ft.com").
	var domainErr *article.UnsupportedDomainError
	if errors.As(err, &domainErr) {
		return strings.Replace(domainErr.Error(), domainErr.Domain, redText.Render(domainErr.Domain), 1)
	}

	// err.Error() can embed server-controlled text (a redirect target, a URL
	// echoed back). Go's url layer rejects raw control bytes there today, but
	// this is the one render path that prints a raw error, so strip defensively.
	errStr := ansi.Strip(err.Error())
	if errStr == "" {
		return "Unknown error"
	}

	first, size := utf8.DecodeRuneInString(errStr)
	msg := string(unicode.ToUpper(first)) + errStr[size:]

	if before, after, ok := strings.Cut(msg, "status "); ok {
		msg = before + "status " + redText.Render(after)
	}

	return msg
}
