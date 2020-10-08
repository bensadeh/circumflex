package comment_parser

// ANSI escape codes
const (
	Normal        = "\033[0m"
	Bold          = "\033[1m"
	Dimmed        = "\033[2m"
	Italic        = "\033[3m"
	Underline     = "\033[4m"
	Red           = "\033[31m"
	Green         = "\033[32m"
	Yellow        = "\033[33m"
	Blue          = "\033[34m"
	Purple        = "\033[35m"
	Teal          = "\033[36m"
	White         = "\033[37m"
	Link1         = "\033]8;;"
	Link2         = "\a"
	Link3         = "\033]8;;\a"
	NewLine       = "\n"
	DoubleNewLine = "\n\n"
)

func normal(text string) string {
	return Normal + text + Normal
}

func bold(text string) string {
	return Bold + text + Normal
}

func italic(text string) string {
	return Italic + text + Normal
}

func dimmed(text string) string {
	return Dimmed + text + Normal
}

func underline(text string) string {
	return Underline + text + Normal
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

func teal(text string) string {
	return Teal + text + Normal
}

func purple(text string) string {
	return Purple + text + Normal
}

func white(text string) string {
	return White + text + Normal
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
		return Blue
	case 5:
		return Teal
	case 6:
		return Purple
	case 7:
		return White
	case 8:
		return Red
	case 9:
		return Yellow
	case 10:
		return Green
	case 11:
		return Blue
	case 12:
		return Teal
	case 13:
		return Purple
	case 14:
		return White
	default:
		return Normal
	}
}
