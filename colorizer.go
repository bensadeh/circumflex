package main

// ANSI escape codes
const (
	Normal        = "\033[0m"
	Bold          = "\033[1m"
	Dimmed        = "\033[2m"
	Italic        = "\033[3m"
	Red           = "\033[31;m"
	Green         = "\033[32;m"
	Yellow        = "\033[33;m"
	Blue          = "\033[34;m"
	Purple        = "\033[35;m"
	Teal          = "\033[36;m"
	Link1         = "\033]8;;"
	Link2         = "\a"
	Link3         = "\033]8;;\a"
	NewLine       = "\n"
	DoubleNewLine = "\n\n"
)

func bold(text string) string {
	return Bold + text + Normal
}

func italic(text string) string {
	return Italic + text + Normal
}

func dimmed(text string) string {
	return Dimmed + text + Normal
}

func red(text string) string {
	return Red + text + Normal
}

func green(text string) string {
	return Green + text + Normal
}
