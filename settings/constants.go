package settings

const (
	ConfigFileNameAbbreviated = "config"

	CommentWidthKey         = "CLX_COMMENT_WIDTH"
	CommentWidthDefault     = 70
	CommentWidthDescription = "Sets the maximum number of characters on each line for comments, replies and " +
		"descriptions in settings. Set to \u001B[1m0\u001B[0m to use the whole screen."
	HighlightHeadlinesKey         = "CLX_HIGHLIGHT_HEADLINES"
	HighlightHeadlinesDefault     = true
	HighlightHeadlinesDescription = "Enables syntax highlighting for the headlines."
	HighlightCommentsKey          = "CLX_HIGHLIGHT_COMMENTS"
	HighlightCommentsDefault      = true
	HighlightCommentsDescription  = "Enables syntax highlighting in the comment section."
	RelativeNumberingKey          = "CLX_RELATIVE_NUMBERING"
	RelativeNumberingDefault      = false
	RelativeNumberingDescription  = "Shows each line with a number relative to the currently selected element. " +
		"Similar to Vim's hybrid line number mode."
	HideYCJobsKey         = "CLX_HIDE_YC_JOBS"
	HideYCJobsDefault     = true
	HideYCJobsDescription = "Hides 'X is hiring' posts from YC-funded startups. Does not affect the monthly 'Who is " +
		"Hiring?' posts."
	UseAltIndentBlockKey         = "CLX_ALT_INDENT_BLOCK"
	UseAltIndentBlockDefault     = false
	UseAltIndentBlockDescription = "Turn this setting on if the indent block does not appear as one connected " +
		"line."
	EmojiSmileysKey             = "CLX_EMOJI_SMILEYS"
	EmojiSmileysDefault         = false
	EmojiSmileysDescription     = "Convert smileys to emojis."
	MarkAsReadKey               = "CLX_MARK_AS_READ"
	MarkAsReadDefault           = true
	MarkAsReadDescription       = "Mark submissions as read after entering the comment section"
	HideIndentSymbolKey         = "CLX_HIDE_INDENT_SYMBOL"
	HideIndentSymbolDefault     = false
	HideIndentSymbolDescription = "Hides the indent symbol from the comment section"
	OrangeHeaderKey             = "CLX_ORANGE_HEADER"
	OrangeHeaderDefault         = false
	OrangeHeaderDescription     = "Sets the background color of the header to orange"
)
