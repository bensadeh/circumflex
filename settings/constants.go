package settings

const (
	ConfigFileNameAbbreviated = "config"
	ConfigFilePath            = "~/.config/circumflex/config.env"
	ConfigDirPath             = "~/.config/circumflex/"

	CommentWidthName        = "Comment Width"
	CommentWidthKey         = "CLX_COMMENT_WIDTH"
	CommentWidthDefault     = 65
	CommentWidthDescription = "Sets the maximum number of characters on each line for comments, replies and " +
		"descriptions in settings. Set to \u001B[1m0\u001B[0m to use the whole screen."
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
	HighlightHeadlinesDefault     = 2
	HighlightHeadlinesDescription = "Highlights YC-funded startups and text containing \033[31mShow HN\033[0m, " +
		"\033[35mAsk HN\033[0m, \033[34mTell HN\033[0m and \033[32mLaunch HN\033[0m. Can be set to \033[1m0\033[0m " +
		"(No highlighting), \u001B[1m1\u001B[0m (inverse highlighting) or \u001B[1m2\u001B[0m (colored highlighting)."
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
	UseAlternateIndentBlockName        = "Use alternate indent block"
	UseAlternateIndentBlockKey         = "CLX_ALT_INDENT_BLOCK"
	UseAlternateIndentBlockDefault     = false
	UseAlternateIndentBlockDescription = "Turn this setting on if the indent block does not appear as one connected " +
		"line."
	CommentHighlightingName        = "Show syntax highlighting in comments"
	CommentHighlightingKey         = "CLX_COMMENT_HIGHLIGHTING"
	CommentHighlightingDefault     = true
	CommentHighlightingDescription = "Enables syntax highlighting for code snippets, @mentions, $variables," +
		" IANAL and IAAL."
	EmojiSmileysName        = "Convert smileys to emojis"
	EmojiSmileysKey         = "CLX_EMOJI_SMILEYS"
	EmojiSmileysDefault     = false
	EmojiSmileysDescription = "Convert smileys to emojis."
	MarkAsReadName          = "Mark submissions as read"
	MarkAsReadKey           = "CLX_MARK_AS_READ"
	MarkAsReadDefault       = false
	MarkAsReadDescription   = "Mark submissions as read after entering the comment section"
)
