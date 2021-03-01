package message

import "clx/utils/format"

func Error(text string) string {
	return format.Red("✘ " + text)
}

func Success(text string) string {
	label := format.Green("✔")

	return label + " " + text
}
