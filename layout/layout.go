package layout

const (
	RankWidth                      = 6
	RankRightPadding               = 1
	MainViewLeftMargin             = RankWidth + RankRightPadding
	HeaderLogoLeftPadding          = 2
	HeaderLogoRightPadding         = 2
	MainViewRightMarginPageCounter = 5
	HelpScreenWidth                = 80
	HeaderLeftMargin               = 2
	CommentSectionLeftMargin       = HeaderLeftMargin
	ReaderViewLeftMargin           = HeaderLeftMargin

	// MinCommentWidth is the floor for a comment's text column. When an indent
	// would push the available comment width below this value, the indent
	// plateaus — deeper comments share an ancestor's indent depth rather than
	// squeezing text to unreadable widths. Nesting is still conveyed by the
	// colored indent symbol.
	MinCommentWidth = 40
)

// ReaderContentWidth returns the usable article width given the full
// screen width and an upper-bound cap.
func ReaderContentWidth(screenWidth, maxWidth int) int {
	w := screenWidth - 2*ReaderViewLeftMargin
	if w <= 0 {
		return maxWidth
	}

	return min(w, maxWidth)
}
