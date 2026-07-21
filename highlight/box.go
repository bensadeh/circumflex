package highlight

import (
	"github.com/bensadeh/circumflex/style"
)

// Boxed wraps highlighted code to the content width and frames it with the
// language label — the one shared rendering of a detected code block, so the
// reader and comment views cannot drift apart.
func Boxed(highlighted, lang string, wrapWidth, boxWidth int) string {
	wrapped := style.WrapWithin(highlighted, wrapWidth-style.RoundedBoxChrome)

	return style.RoundedBox(wrapped, boxWidth, Label(lang))
}
