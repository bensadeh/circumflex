package settings

const (
	ConfigFileNameAbbreviated = "config"
	ConfigFilePath            = "~/.config/circumflex/config.env"
	ConfigDirPath             = "~/.config/circumflex/"

	CommentWidthKey         = "CLX_COMMENT_WIDTH"
	CommentWidthDefault     = 65
	CommentWidthDescription = "Sets the maximum number of characters on each line for comments, replies and " +
		"descriptions in settings. Set to \u001B[1m0\u001B[0m to use the whole screen."
	IndentSizeKey         = "CLX_INDENT_SIZE"
	IndentSizeDefault     = 4
	IndentSizeDescription = "The number of whitespaces prepended to each reply, " +
		"not including the color bar."
	PreserveRightMarginKey         = "CLX_PRESERVE_RIGHT_MARGIN"
	PreserveRightMarginDefault     = false
	PreserveRightMarginDescription = "Shortens replies so that the total length, including indentation, is the same " +
		"as the comment width. Best used when Indent Size is small to avoid deep replies being too short."
	HighlightHeadlinesKey         = "CLX_HIGHLIGHT_HEADLINES"
	HighlightHeadlinesDefault     = 2
	HighlightHeadlinesDescription = "Highlights YC-funded startups and text containing \033[31mShow HN\033[0m, " +
		"\033[35mAsk HN\033[0m, \033[34mTell HN\033[0m and \033[32mLaunch HN\033[0m. Can be set to \033[1m0\033[0m " +
		"(No highlighting), \u001B[1m1\u001B[0m (inverse highlighting) or \u001B[1m2\u001B[0m (colored highlighting)."
	RelativeNumberingKey         = "CLX_RELATIVE_NUMBERING"
	RelativeNumberingDefault     = false
	RelativeNumberingDescription = "Shows each line with a number relative to the currently selected element. " +
		"Similar to Vim's hybrid line number mode."
	HideYCJobsKey         = "CLX_HIDE_YC_JOBS"
	HideYCJobsDefault     = true
	HideYCJobsDescription = "Hides 'X is hiring' posts from YC-funded startups. Does not affect the monthly 'Who is " +
		"Hiring?' posts."
	UseAltIndentBlockKey         = "CLX_ALT_INDENT_BLOCK"
	UseAltIndentBlockDefault     = false
	UseAltIndentBlockDescription = "Turn this setting on if the indent block does not appear as one connected " +
		"line."
	CommentHighlightingKey         = "CLX_COMMENT_HIGHLIGHTING"
	CommentHighlightingDefault     = true
	CommentHighlightingDescription = "Enables syntax highlighting for code snippets, @mentions, $variables," +
		" IANAL and IAAL."
	EmojiSmileysKey         = "CLX_EMOJI_SMILEYS"
	EmojiSmileysDefault     = false
	EmojiSmileysDescription = "Convert smileys to emojis."
	MarkAsReadKey           = "CLX_MARK_AS_READ"
	MarkAsReadDefault       = true
	MarkAsReadDescription   = "Mark submissions as read after entering the comment section"
)
