package comments

import (
	"github.com/bensadeh/circumflex/comment"
	"github.com/bensadeh/circumflex/view/pane"
)

func Run(thread *comment.Thread, lastVisited int64, commentWidth, indent int, enableNerdFonts bool) error {
	return pane.RunStandalone(thread.Title, func(width, height int) pane.View {
		m := New(thread, lastVisited, commentWidth, indent, enableNerdFonts, width, height)
		m.DisableAppKeys()

		return m
	}, nil)
}
