// Package layout owns the terminal geometry: the wide-layout split arithmetic
// and the width every text column derives from the pane it renders in.
//
// The detail views (comment section, reader) never see the pane split. The
// parent hands each view a WindowSizeMsg sized to its pane, and the view lays
// itself out as if that were the whole terminal — the same code path serves
// the narrow full-screen layout, the wide split, and the standalone
// subcommands. Widths named "pane" here are that logical screen.
package layout

import "github.com/bensadeh/circumflex/scrollbar"

const (
	RankWidth                      = 6
	RankRightPadding               = 1
	MainViewLeftMargin             = RankWidth + RankRightPadding
	HeaderLogoLeftPadding          = 2
	HeaderLogoRightPadding         = 2
	MainViewRightMarginPageCounter = 5
	HeaderLeftMargin               = 2
	CommentSectionLeftMargin       = HeaderLeftMargin
	ReaderViewLeftMargin           = HeaderLeftMargin

	// MinCommentWidth is the floor for a comment's text column. When an indent
	// would push the available comment width below this value, the indent
	// plateaus — deeper comments share an ancestor's indent depth rather than
	// squeezing text to unreadable widths. Nesting is still conveyed by the
	// colored indent symbol.
	MinCommentWidth = 40

	// PaneHeaderHeight and PaneFooterHeight are the rows a detail or list pane
	// reserves for its header and footer; the remainder holds scrollable
	// content. PaneChromeHeight is their sum.
	PaneHeaderHeight = 2
	PaneFooterHeight = 2
	PaneChromeHeight = PaneHeaderHeight + PaneFooterHeight

	// PaneDividerWidth is the columns between the two panes in the wide
	// layout: a one-column rule with a space of breathing room on each side.
	PaneDividerWidth = 3

	// WideViewFloor is the narrowest terminal the split layout renders sanely;
	// below it the wide view stays off even when configured "always".
	WideViewFloor = 40
)

// Frame is the terminal geometry for one render: the screen size and whether
// the two-pane wide layout is active. Every pane width and the pane content
// height derive from it, so the split arithmetic lives in exactly one place.
type Frame struct {
	Width  int
	Height int
	Wide   bool
}

// ListWidth is the width the story list renders at: the left pane when wide,
// the full screen otherwise. The divider sits in the middle, so both panes get
// an equal share.
func (f Frame) ListWidth() int {
	if f.Wide {
		return (f.Width - PaneDividerWidth) / 2
	}

	return f.Width
}

// DetailWidth is the width the comment section and reader render at: the right
// pane when wide, the full screen otherwise.
func (f Frame) DetailWidth() int {
	if f.Wide {
		return f.Width - f.ListWidth() - PaneDividerWidth
	}

	return f.Width
}

// PaneContentHeight is the rows available inside a pane, below its header and
// above its footer.
func (f Frame) PaneContentHeight() int {
	return max(0, f.Height-PaneChromeHeight)
}

// ReaderContentWidth is the article text column: the configured maximum,
// clamped to the pane minus symmetric margins. Never below one column, so
// degenerate panes render narrow instead of overflowing.
func ReaderContentWidth(paneWidth, maxWidth int) int {
	return max(1, min(paneWidth-2*ReaderViewLeftMargin, maxWidth))
}

// CommentContentWidth is the columns a comment may span: from the left margin
// to the scrollbar column. Deeper comments indent within it.
func CommentContentWidth(paneWidth int) int {
	return max(1, paneWidth-CommentSectionLeftMargin-scrollbar.Width)
}

// CommentColumnWidth is the top-level comment text column: the configured
// width, clamped to the pane. The meta header and the footer indicator share
// it, so all three right edges stay aligned however narrow the pane gets.
func CommentColumnWidth(paneWidth, configuredWidth int) int {
	return min(CommentContentWidth(paneWidth), configuredWidth)
}
