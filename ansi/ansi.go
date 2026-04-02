package ansi

import (
	"regexp"
)

const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Faint     = "\033[2m"
	Italic    = "\033[3m"
	Reverse   = "\033[7m"
	ItalicOff = "\033[23m"
	Red       = "\033[31m"
	Green     = "\033[32m"
)

// jsonByte returns a regex fragment matching the JSON unicode escape for a
// given hex byte with case-insensitive hex digits (e.g. "9b" matches both
// \u009b and \u009B). The hex argument must be lowercase.
func jsonByte(hex string) string {
	d1, d2 := hex[0], hex[1]
	if d2 >= 'a' && d2 <= 'f' {
		return `\\[Uu]00` + string(d1) + `[` + string([]byte{d2, d2 - 32}) + `]`
	}

	return `\\[Uu]00` + string(d1) + string(d2)
}

var (
	// escSequences matches all 7-bit and 8-bit ANSI/VT escape sequences.
	// JSON unicode escapes (\u001B etc.) are also matched because Strip is
	// called on raw HTTP response bodies before JSON parsing.
	//
	// Covered families:
	//   - CSI  (ESC[ or \u009B)
	//   - OSC  (ESC] or \u009D), DCS (ESCP or \u0090),
	//     APC  (ESC_ or \u009F), PM (ESC^ or \u009E), SOS (ESCX or \u0098)
	//   - SS2  (ESCN or \u008E), SS3 (ESCO or \u008F)
	//   - nF   sequences: ESC + intermediate bytes (0x20-0x2F) + final byte
	//   - Two-byte ESC sequences (Fe, Fp, Fs classes)

	// jsonESC matches the JSON unicode escape for ESC (0x1B) with any hex
	// casing — the HN API returns uppercase (\u001B).
	jsonESC = `\\[Uu]001[Bb]`

	escSequences = regexp.MustCompile(
		// CSI: parameter bytes, intermediate bytes, final byte.
		`(?:\x1B\[|` + jsonESC + `\[|` + jsonByte("9b") + `)[0-?]*[ -/]*[@-~]|` +
			// String sequences (OSC, DCS, APC, PM, SOS) terminated by ST or BEL.
			`(?:\x1B[\]P_^X]|` + jsonESC + `[\]P_^X]|` +
			jsonByte("9d") + `|` + jsonByte("90") + `|` + jsonByte("9f") + `|` +
			jsonByte("9e") + `|` + jsonByte("98") + `)[\x00-\x7E]*?(?:\x1B\\|\x07|\x9C|` + jsonESC + `\\|\\[Uu]0007|` + jsonByte("9c") + `)|` +
			// SS2, SS3.
			`(?:\x1B[NO]|` + jsonESC + `[NO]|` + jsonByte("8e") + `|` + jsonByte("8f") + `).?|` +
			// nF sequences: ESC + one or more intermediate bytes + final byte.
			`(?:\x1B|` + jsonESC + `)[\x20-\x2F]+[\x30-\x7E]|` +
			// Two-byte ESC sequences: Fe (0x40-0x5F), Fp (0x30-0x3F), Fs (0x60-0x7E).
			`(?:\x1B|` + jsonESC + `)[0-~]`,
	)

	// dangerousControls strips C0 control characters that can manipulate
	// terminal state outside of escape sequences (e.g. backspace overwrites,
	// BEL noise, DEL, bare ESC). Preserves \t (0x09), \n (0x0A), \r (0x0D).
	//
	// Also strips JSON short escapes \b (backspace) and \f (form feed) which
	// json.Unmarshal would decode into raw 0x08/0x0C after Strip returns.
	dangerousControls = regexp.MustCompile(
		// Raw C0 controls (except \t, \n, \r) + ESC + DEL.
		`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]|` +
			// JSON short escapes for backspace and form feed.
			`\\[bf]|` +
			// JSON \uXXXX for C0 controls (0x00-0x08, 0x0B-0x0C, 0x0E-0x1F, 0x7F).
			`\\[Uu]00(?:0[0-8BbCcEeFf]|1[0-9A-Fa-f]|7[Ff])`,
	)
)

func Strip(text string) string {
	text = escSequences.ReplaceAllString(text, "")

	return dangerousControls.ReplaceAllString(text, "")
}
