package comment

// ANSI escape codes
const (
	Normal       = "\033[0m"
	Bold         = "\033[1m"
	Dimmed       = "\033[2m"
	Italic       = "\033[3m"
	Underline    = "\033[4m"
	Reversed     = "\033[7m"
	Red          = "\033[31m"
	Green        = "\033[32m"
	Yellow       = "\033[33m"
	Blue         = "\033[34m"
	Magenta      = "\033[35m"
	Cyan         = "\033[36m"
	White        = "\033[37m"
	AltRed       = "\033[31;1m"
	AltGreen     = "\033[32;1m"
	AltYellow    = "\033[33;1m"
	AltBlue      = "\033[34;1m"
	AltMagenta   = "\033[35;1m"
	AltCyan      = "\033[36;1m"
	AltWhite     = "\033[37;1m"
	Link1        = "\033]8;;"
	Link2        = "\a"
	Link3        = "\033]8;;\a"
	NewLine      = "\n"
	NewParagraph = "\n\n"
)

func bold(text string) string {
	return Bold + text + Normal
}

func dimmed(text string) string {
	return Dimmed + text + Normal
}

func dimmedAndUnderlined(text string) string {
	return Dimmed + Underline + text + Normal
}

func red(text string) string {
	return Red + text + Normal
}

func yellow(text string) string {
	return Yellow + text + Normal
}

func green(text string) string {
	return Green + text + Normal
}

func blue(text string) string {
	return Blue + text + Normal
}

func cyan(text string) string {
	return Cyan + text + Normal
}

func magenta(text string) string {
	return Magenta + text + Normal
}

func white(text string) string {
	return White + text + Normal
}

func altRed(text string) string {
	return AltRed + text + Normal
}

func altYellow(text string) string {
	return AltYellow + text + Normal
}

func altGreen(text string) string {
	return AltGreen + text + Normal
}

func altWhite(text string) string {
	return AltWhite + text + Normal
}

func paren(text string) string {
	return "(" + text + ")"
}

func getColoredIndentBlock(level int) string {
	switch level {
	case 1:
		return Red
	case 2:
		return Yellow
	case 3:
		return Green
	case 4:
		return Cyan
	case 5:
		return Blue
	case 6:
		return Magenta
	case 7:
		return White
	case 8:
		return AltRed
	case 9:
		return AltYellow
	case 10:
		return AltGreen
	case 11:
		return AltCyan
	case 12:
		return AltBlue
	case 13:
		return AltMagenta
	case 14:
		return AltWhite
	default:
		return Normal
	}
}
