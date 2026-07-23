package pane

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bensadeh/circumflex/ansi"
)

// The window title names the terminal tab after what is on screen. The OSC 2
// write itself is Bubble Tea's, driven by tea.View.WindowTitle, so it is
// serialized with the frame flushes rather than racing them — the same
// hazard progress.go describes.
//
// The title stack around it is ours: CSI 22;2t saves the title the terminal
// carried before circumflex started and CSI 23;2t puts it back on exit, so
// quitting leaves the window named whatever it was named before. Bubble Tea's
// own shutdown only blanks the title, which would leave the tab nameless
// until the shell's next prompt renamed it. Terminals without a title stack
// ignore both sequences.

var TitleOut io.Writer = os.Stdout

const (
	pushWindowTitleSeq = "\033[22;2t"
	popWindowTitleSeq  = "\033[23;2t"
)

// SaveWindowTitle pushes the terminal's current window title and returns the
// pop that restores it. Call it before the program starts and the restore
// after it exits: like the progress sequences, a direct write is only safe
// where no frame flush can interleave with it.
func SaveWindowTitle() (restore func()) {
	_, _ = fmt.Fprint(TitleOut, pushWindowTitleSeq)

	return func() {
		_, _ = fmt.Fprint(TitleOut, popWindowTitleSeq)
	}
}

// WindowTitle prepares text for use as a window title. OSC 2 ends at the
// first BEL or ST, so a control character in a story title would close the
// sequence early and let whatever follows land on the terminal as commands;
// a newline in the payload would spill the rest onto the screen. Titles are
// already stripped at ingest — this is the sink saying so itself, as
// ansi.Hyperlink does for its target.
func WindowTitle(text string) string {
	return strings.Join(strings.Fields(ansi.Strip(text)), " ")
}
