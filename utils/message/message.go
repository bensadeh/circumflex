package message

import "clx/utils/formatter"

func Error(text string) string {
	return formatter.Red("✘ " + text)
}

func Success(text string) string {
	label := formatter.Green("✔")

	return label + " " + text
}

func Warning(text string) string {
	label := formatter.Yellow("!")

	return label + " " + text
}
