package settings

const (
	ConfigFileNameAbbreviated = "config"
	ConfigFileNameFull        = "config.env"

	CommentWidthName        = "Comment Width"
	CommentWidthKey         = "CLX_COMMENT_WIDTH"
	CommentWidthDefault     = 70
	CommentWidthDescription = "Sets the maximum number of characters on each line for comments, replies and " +
		"descriptions in settings. Set to \u001B[31m0\u001B[0m to use the whole screen."
	IndentSizeName        = "Indent Size"
	IndentSizeKey         = "CLX_INDENT_SIZE"
	IndentSizeDefault     = 4
	IndentSizeDescription = "The number of whitespaces prepended to each reply, " +
		"not including the color bar."
	PreserveRightMarginName        = "Preserve Right Margin"
	PreserveRightMarginKey         = "CLX_PRESERVE_RIGHT_MARGIN"
	PreserveRightMarginDefault     = false
	PreserveRightMarginDescription = "Shortens replies so that the total length, including indentation, is the same " +
		"as the comment width. Best used when Indent Size is small to avoid deep replies being too short."
	HighlightHeadlinesName        = "Highlight Headlines"
	HighlightHeadlinesKey         = "CLX_HIGHLIGHT_HEADLINES"
	HighlightHeadlinesDefault     = 0
	HighlightHeadlinesDescription = "Highlights text containing \033[31mShow HN\033[0m, \033[35mAsk HN\033[0m, " +
		"\033[34mTell HN\033[0m or \033[32mLaunch HN\033[0m. Can be set to \033[1m0\033[0m (No highlighting), " +
		"\u001B[1m1\u001B[0m (inverse highlighting) or \u001B[1m2\u001B[0m (colored highlighting). YC-funded " +
		"startups are colorized in both option \u001B[1m1\u001B[0m and \u001B[1m2\u001B[0m."
	RelativeNumberingName        = "Use Relative Numbering"
	RelativeNumberingKey         = "CLX_RELATIVE_NUMBERING"
	RelativeNumberingDefault     = false
	RelativeNumberingDescription = "Shows each line with a number relative to the currently selected element. " +
		"Similar to Vim's hybrid line number mode."
	HideYCJobsName        = "Hide YC hiring posts"
	HideYCJobsKey         = "CLX_HIDE_YC_JOBS"
	HideYCJobsDefault     = true
	HideYCJobsDescription = "Hides 'X is hiring' posts from YC-funded startups. Does not affect the monthly 'Who is " +
		"Hiring?' posts."
)
