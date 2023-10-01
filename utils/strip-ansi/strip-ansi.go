package stripansi

import (
	"regexp"
)

const (
	esc         = "[\u001B\u009B]"
	oscSequence = "(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007"
	csiSequence = "(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~])"
	ansi        = esc + "[[\\]()#;?]*" + "(?:" + oscSequence + "|" + csiSequence + ")"
)

var ansiRegex = regexp.MustCompile(ansi)

func Strip(text string) string {
	return ansiRegex.ReplaceAllString(text, "")
}
