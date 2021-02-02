package message

import "clx/utils/format"

func Error(text string) string {
	label := format.BlackOnRed(" ERROR ")

	return label + " " + text
}
