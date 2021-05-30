package message

import "clx/utils/formatter"

func Error(text string) string {
	label := formatter.Red("✘")

	return label + " " + text
}

func Success(text string) string {
	label := formatter.Green("✔")

	return label + " " + text
}

func Warning(text string) string {
	label := formatter.Yellow("!")

	return label + " " + text
}
