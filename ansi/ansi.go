package ansi

import (
	"regexp"
)

const (
	Reset             = "\033[0m"
	Bold              = "\033[1m"
	Faint             = "\033[2m"
	Italic            = "\033[3m"
	Underline         = "\033[4m"
	Reverse           = "\033[7m"
	Strikethrough     = "\033[9m"
	NormalIntensity   = "\033[22m" // clears both Bold and Faint
	ItalicOff         = "\033[23m"
	UnderlineOff      = "\033[24m"
	StrikethroughOff  = "\033[29m"
	Red               = "\033[31m"
	Green             = "\033[32m"
	Yellow            = "\033[33m"
	Blue              = "\033[34m"
	Cyan              = "\033[36m"
	DefaultForeground = "\033[39m"
	BgBrightBlack     = "\033[100m"
)

const (
	hyperlinkOpen       = "\033]8;;"
	hyperlinkTerminator = "\033\\"
)

// Hyperlink wraps text in an OSC 8 hyperlink pointing at url.
func Hyperlink(url, text string) string {
	return hyperlinkOpen + url + hyperlinkTerminator + text + hyperlinkOpen + hyperlinkTerminator
}

var (
	escSequences = regexp.MustCompile(
		`(?:\x1B\[|\x9B)[0-?]*[ -/]*[@-~]|` +
			`(?:\x1B[\]P_^X]|[\x9D\x90\x9F\x9E\x98])[\x00-\x7E]*?(?:\x1B\\|\x07|\x9C)|` +
			`(?:\x1B[NO]|[\x8E\x8F]).?|` +
			`\x1B[\x20-\x2F]+[\x30-\x7E]|` +
			`\x1B[0-~]`,
	)

	// C0 controls except \t \n \r, plus DEL.
	dangerousControls = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
)

func Strip(text string) string {
	text = escSequences.ReplaceAllString(text, "")

	return dangerousControls.ReplaceAllString(text, "")
}
