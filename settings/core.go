package settings

type Config struct {
	CommentWidth               int
	PlainHeadlines             bool
	HighlightHeadlines         bool
	HighlightComments          bool
	RelativeNumbering          bool
	EmojiSmileys               bool
	DoNotMarkSubmissionsAsRead bool
	HideIndentSymbol           bool
	IndentationSymbol          string
	DebugMode                  bool
	EnableNerdFonts            bool
	LesskeyPath                string
	AutoCollapseComments       bool
}

func Default() *Config {
	return &Config{
		CommentWidth:               70,
		HighlightHeadlines:         true,
		HighlightComments:          true,
		RelativeNumbering:          false,
		EmojiSmileys:               true,
		DoNotMarkSubmissionsAsRead: false,
		HideIndentSymbol:           false,
		IndentationSymbol:          " ▎",
	}
}
