package colors

// ANSI escape codes
const (
	Underline        = "\033[4m"
	Reversed         = "\033[7m"
	Red              = "\033[31m"
	Green            = "\033[32m"
	Yellow           = "\033[33m"
	Blue             = "\033[34m"
	Magenta          = "\033[35m"
	Cyan             = "\033[36m"
	White            = "\033[37m"
	AltRed           = "\033[31;1m"
	AltGreen         = "\033[32;1m"
	AltYellow        = "\033[33;1m"
	AltBlue          = "\033[34;1m"
	AltMagenta       = "\033[35;1m"
	AltCyan          = "\033[36;1m"
	AltWhite         = "\033[37;1m"
	OrangeBackground = "\033[48;5;214m"
	NearBlack        = "\033[38;5;232m"
	Link1            = "\033]8;;"
	Link2            = "\a"
	Link3            = "\033]8;;\a"
	// NewLine          = "\n"
	// NewParagraph     = "\n\n"
)

//func ToBold(text string) string {
//	return Bold + text + Normal
//}
//
//func ToDimmed(text string) string {
//	return Dimmed + text + Normal
//}

//func ToDimmedAndUnderlined(text string) string {
//	return Dimmed + Underline + text + Normal
//}
//
//func ToRed(text string) string {
//	return Red + text + Normal
//}
//
//func ToYellow(text string) string {
//	return Yellow + text + Normal
//}
//
//func ToGreen(text string) string {
//	return Green + text + Normal
//}
//
//func ToBlue(text string) string {
//	return Blue + text + Normal
//}
//
//func ToCyan(text string) string {
//	return Cyan + text + Normal
//}
//
//func ToMagenta(text string) string {
//	return Magenta + text + Normal
//}
//
//func ToWhite(text string) string {
//	return White + text + Normal
//}
//
//func ToBrightRed(text string) string {
//	return AltRed + text + Normal
//}
//
//func ToBrightYellow(text string) string {
//	return AltYellow + text + Normal
//}
//
//func ToBrightGreen(text string) string {
//	return AltGreen + text + Normal
//}
//
//func ToBrightWhite(text string) string {
//	return AltWhite + text + Normal
//}
//
//func SurroundWithParen(text string) string {
//	return "(" + text + ")"
//}

//func GetIndentBlockColor(level int) string {
//	switch level {
//	case 1:
//		return Red
//	case 2:
//		return Yellow
//	case 3:
//		return Green
//	case 4:
//		return Cyan
//	case 5:
//		return Blue
//	case 6:
//		return Magenta
//	case 7:
//		return AltRed
//	case 8:
//		return AltYellow
//	case 9:
//		return AltGreen
//	case 10:
//		return AltCyan
//	case 11:
//		return AltBlue
//	case 12:
//		return AltMagenta
//	default:
//		return Normal
//	}
//}
