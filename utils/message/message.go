package message

import "clx/utils/formatter"

func Error(text string) string {
	label := formatter.Red("✘")

	return format(label, text)
}

func Success(text string) string {
	label := formatter.Green("✔")

	return format(label, text)
}

func Warning(text string) string {
	label := formatter.Yellow("!")

	return format(label, text)
}

func format(label string, text string) string {
	return label + " " + text
}
