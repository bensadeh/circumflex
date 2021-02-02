package message

import "clx/utils/format"

func Error(text string) string {
	label := format.BlackOnRed(" ERROR ")
	textInRed := " " + format.Red(text)

	return label + textInRed
}
