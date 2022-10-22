package settings

type Config struct {
	CommentWidth                int
	DisableHeadlineHighlighting bool
	DisableCommentHighlighting  bool
	DisableEmojis               bool
	DoNotMarkSubmissionsAsRead  bool
	HideIndentSymbol            bool
	IndentationSymbol           string
	DebugMode                   bool
	EnableNerdFonts             bool
	LesskeyPath                 string
	AutoExpandComments          bool
	NoLessVerify                bool
}

func Default() *Config {
	return &Config{
		CommentWidth:      70,
		IndentationSymbol: " ▎",
	}
}
