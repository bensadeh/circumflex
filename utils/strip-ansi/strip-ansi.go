package stripansi

import (
	"regexp"
)

var combinedRegex = regexp.MustCompile(
	`(\x1B\[|\x9B\[|\\u001b\[|\\u009b\[)[0-?]*[ -/]*[@-~]|` + // CSI Sequences
		`\x1B\]|\x9B].*?(\007|\x1B\\)|\\u001b\]|\\u009b\].*?(\\u0007|\\u001b\\)`, // OSC Sequences
)

func Strip(text string) string {
	return combinedRegex.ReplaceAllString(text, "")
}
