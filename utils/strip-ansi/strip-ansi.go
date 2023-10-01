package stripansi

import (
	"regexp"
)

var combinedRegex = regexp.MustCompile(`(\x1B\[|\x9B\[|\\u001b\[|\\u009b\[)[0-?]*[ -/]*[@-~]`)

func Strip(text string) string {
	return combinedRegex.ReplaceAllString(text, "")
}
