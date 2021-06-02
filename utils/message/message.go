package message

import "clx/utils/formatter"

func Error(text string) string {
	label := formatter.BlackOnRed(" ✘ ")

	return format(label, text)
}

func Success(text string) string {
	label := formatter.BlackOnGreen(" ✔ ")

	return format(label, text)
}

func Warning(text string) string {
	label := formatter.BlackOnYellow(" ! ")

	return format(label, text)
}

func format(label string, text string) string {
	return label + " " + text
}
