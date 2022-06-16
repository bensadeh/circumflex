package settings

type Config struct {
	CommentWidth       int
	PlainHeadlines     bool
	HighlightHeadlines bool
	HighlightComments  bool
	RelativeNumbering  bool
	EmojiSmileys       bool
	MarkAsRead         bool
	HideIndentSymbol   bool
	IndentationSymbol  string
	DebugMode          bool
	EnableNerdFonts    bool
}

func New() *Config {
	return &Config{
		CommentWidth:       70,
		HighlightHeadlines: true,
		HighlightComments:  true,
		RelativeNumbering:  false,
		EmojiSmileys:       true,
		MarkAsRead:         true,
		HideIndentSymbol:   false,
		IndentationSymbol:  " ▎",
	}
}
