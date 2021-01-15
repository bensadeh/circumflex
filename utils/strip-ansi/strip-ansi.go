package stripansi

import (
	"regexp"
)

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|" +
	"(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

func Strip(text string) string {
	expression := regexp.MustCompile(ansi)
	return expression.ReplaceAllString(text, "")
}
