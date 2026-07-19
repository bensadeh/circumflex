package ansi

import (
	"regexp"
	"strings"
)

const (
	Reset             = "\033[0m"
	Bold              = "\033[1m"
	Faint             = "\033[2m"
	Italic            = "\033[3m"
	Underline         = "\033[4m"
	UnderlineDashed   = "\033[4:5m" // only valid after Underline: terminals without styled underlines ignore it, keeping the plain underline
	Reverse           = "\033[7m"
	Strikethrough     = "\033[9m"
	NormalIntensity   = "\033[22m" // clears both Bold and Faint
	ItalicOff         = "\033[23m"
	UnderlineOff      = "\033[24m"
	UnderlineColorOff = "\033[59m" // back to the default underline color: the text color
	ReverseOff        = "\033[27m"
	StrikethroughOff  = "\033[29m"
	Red               = "\033[31m"
	Green             = "\033[32m"
	Yellow            = "\033[33m"
	Blue              = "\033[34m"
	Cyan              = "\033[36m"
	DefaultForeground = "\033[39m"
	DefaultBackground = "\033[49m"
	BgBrightBlack     = "\033[100m"
)

const (
	hyperlinkOpen       = "\033]8;;"
	hyperlinkTerminator = "\033\\"
)

// Hyperlink wraps text in an OSC 8 hyperlink pointing at url.
func Hyperlink(url, text string) string {
	return hyperlinkOpen + url + hyperlinkTerminator + text + hyperlinkOpen + hyperlinkTerminator
}

var (
	escSequences = regexp.MustCompile(
		`(?:\x1B\[|\x9B)[0-?]*[ -/]*[@-~]|` +
			`(?:\x1B[\]P_^X]|[\x9D\x90\x9F\x9E\x98])[\x00-\x7E]*?(?:\x1B\\|\x07|\x9C)|` +
			`(?:\x1B[NO]|[\x8E\x8F]).?|` +
			`\x1B[\x20-\x2F]+[\x30-\x7E]|` +
			`\x1B[0-~]`,
	)

	// C0 controls except \t \n \r, plus DEL and the C1 range: a bare C1 rune
	// — an unterminated U+009D string opener, a stray U+009B — is a live
	// control on terminals that decode C1 from UTF-8.
	dangerousControls = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F\x80-\x9F]`)
)

// Neutralize makes control characters visible instead of removing them: ESC
// becomes ␛, other C0 controls and DEL their control pictures, and a C1 rune
// its 7-bit ␛-pair equivalent. Content paths use it so an article
// demonstrating escape sequences shows ␛[31m where Strip would delete the
// sequence outright; \t \n \r pass through, as in Strip.
func Neutralize(text string) string {
	var sb strings.Builder

	sb.Grow(len(text))

	for _, r := range text {
		switch {
		case r == '\t' || r == '\n' || r == '\r':
			sb.WriteRune(r)

		case r < 0x20:
			sb.WriteRune(0x2400 + r)

		case r == 0x7F:
			sb.WriteRune('␡')

		case 0x80 <= r && r <= 0x9F:
			sb.WriteRune('␛')
			sb.WriteRune(0x40 + r - 0x80)

		default:
			sb.WriteRune(r)
		}
	}

	return sb.String()
}

func Strip(text string) string {
	// Invalid UTF-8 is repaired first: removing a C0 byte could otherwise
	// splice the bytes around it into a C1 control rune ("\xc2\x11\x9b"
	// would strip to the CSI rune U+009B).
	text = strings.ToValidUTF8(text, "�")
	text = escSequences.ReplaceAllString(text, "")

	return dangerousControls.ReplaceAllString(text, "")
}
