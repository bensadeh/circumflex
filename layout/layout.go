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
