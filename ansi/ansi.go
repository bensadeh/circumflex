package ansi

import (
	"regexp"
)

const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Faint     = "\033[2m"
	Italic    = "\033[3m"
	Reverse   = "\033[7m"
	ItalicOff = "\033[23m"
	Red       = "\033[31m"
	Green     = "\033[32m"
)

var combinedRegex = regexp.MustCompile(
	`(\x1B\[|\x9B\[|\\u001b\[|\\u009b\[)[0-?]*[ -/]*[@-~]|` + // CSI Sequences
		`\x1B\]|\x9B].*?(\007|\x1B\\)|\\u001b\]|\\u009b\].*?(\\u0007|\\u001b\\)`, // OSC Sequences
)

func Strip(text string) string {
	return combinedRegex.ReplaceAllString(text, "")
}
