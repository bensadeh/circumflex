package stripansi

import (
	"regexp"
)

const esc = "[\u001B\u009B]"
const oscSequence = "(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007"
const csiSequence = "(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~])"
const ansi = esc + "[[\\]()#;?]*" + "(?:" + oscSequence + "|" + csiSequence + ")"

var ansiRegex = regexp.MustCompile(ansi)

func Strip(text string) string {
	return ansiRegex.ReplaceAllString(text, "")
}
