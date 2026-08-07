package comments

import (
	"github.com/bensadeh/circumflex/comment"
	"github.com/bensadeh/circumflex/view/pane"
)

// Run hosts the comment section as its own program. shell supplies the link
// factories from the cmd layer, so followed links open in place as the full
// app's do.
func Run(thread *comment.Thread, lastVisited int64, commentWidth, indent int, enableNerdFonts bool, shell pane.StandaloneOptions) error {
	return pane.RunStandalone(thread.Title, func(width, height int) pane.View {
		m := New(thread, lastVisited, commentWidth, indent, enableNerdFonts, width, height)
		m.DisableAppKeys()

		return m
	}, shell)
}
