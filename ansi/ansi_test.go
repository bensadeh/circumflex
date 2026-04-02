package ansi_test

import (
	"testing"

	"github.com/bensadeh/circumflex/ansi"
)

func TestStrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// CSI sequences (7-bit).
		{"SGR color", "\x1B[31mhello\x1B[0m", "hello"},
		{"SGR bold+color", "\x1B[1;34mblue\x1B[0m", "blue"},
		{"cursor movement", "\x1B[2Aup\x1B[5Bdown", "updown"},
		{"erase display", "\x1B[2J\x1B[Hhome", "home"},

		// OSC sequences.
		{"OSC window title BEL", "\x1B]0;evil title\x07safe", "safe"},
		{"OSC window title ST", "\x1B]0;evil title\x1B\\safe", "safe"},
		{"OSC hyperlink", "\x1B]8;;https://evil.com\x07click\x1B]8;;\x07", "click"},

		// DCS sequences.
		{"DCS 7-bit", "\x1BPq#0;2;0;0;0\x1B\\ok", "ok"},

		// APC, PM, SOS.
		{"APC 7-bit", "\x1B_payload\x1B\\text", "text"},
		{"PM 7-bit", "\x1B^message\x1B\\text", "text"},
		{"SOS 7-bit", "\x1BXstring\x1B\\text", "text"},

		// SS2, SS3 consume the introducer + one character.
		{"SS2", "\x1BNA ok", " ok"},
		{"SS3", "\x1BOB ok", " ok"},

		// nF sequences: ESC + intermediate bytes + final byte.
		{"nF DEC graphics charset", "\x1B(0text\x1B(B", "text"},
		{"nF DECALN fill screen", "\x1B#8text", "text"},
		{"nF S7C1T", "\x1B Ftext", "text"},
		{"nF S8C1T", "\x1B Gtext", "text"},

		// Two-byte ESC sequences (Fe, Fp, Fs ranges).
		{"ESC save cursor", "\x1B7text\x1B8", "text"},
		{"ESC reverse index", "\x1BMline", "line"},

		// Bare ESC stripped by dangerousControls as safety net.
		{"bare ESC at end", "text\x1B", "text"},
		// ESC + printable is a two-byte sequence (Fs class), stripped as pair.
		{"ESC + printable", "a\x1Bb", "a"},
		// ESC + space + printable is an nF sequence, stripped as triple.
		{"ESC SP b (nF)", "a\x1B b", "a"},

		// Dangerous control characters (stripped, not simulated).
		{"backspace", "ab\x08\x08cd", "abcd"},
		{"BEL standalone", "hello\x07world", "helloworld"},
		{"DEL", "ok\x7Ftext", "oktext"},
		{"NUL", "a\x00b", "ab"},
		{"FS GS RS US", "a\x1C\x1D\x1E\x1Fb", "ab"},

		// JSON short escapes for backspace and form feed.
		{`JSON \b backspace`, `normal\b\b\b\b\b\bSPOOF`, "normalSPOOF"},
		{`JSON \f form feed`, `text\fmore`, "textmore"},

		// Preserves safe whitespace.
		{"tab preserved", "a\tb", "a\tb"},
		{"newline preserved", "a\nb", "a\nb"},
		{"CR preserved", "a\rb", "a\rb"},

		// JSON unicode escapes (literal strings, pre-JSON-parse).
		{"JSON CSI lowercase", `\u001b[31mred\u001b[0m`, "red"},
		{"JSON CSI uppercase", `\u001B[32mDoes this mean HN can support color text now?`, "Does this mean HN can support color text now?"},
		{"JSON CSI mixed case", `\u001B[1;34mblue\u001b[0m`, "blue"},
		{"JSON 8-bit CSI", `\u009b31mred\u009b0m`, "red"},
		{"JSON 8-bit CSI uppercase", `\u009B31mred\u009B0m`, "red"},
		{"JSON OSC", `\u001b]0;title\u0007safe`, "safe"},
		{"JSON OSC uppercase", `\u001B]0;title\u0007safe`, "safe"},
		// \u001B + space + 'a' is a valid nF sequence, stripped as unit.
		{"JSON ESC nF", `\u001B alone`, "lone"},
		// \u001B at end of string, caught by dangerousControls.
		{"JSON bare ESC", `end\u001B`, "end"},
		{"JSON controls 1C-1F", `a\u001Cb\u001Dc\u001Ed\u001Fe`, "abcde"},

		// jsonESC must not false-match \u002b (plus sign).
		{`JSON plus not stripped`, `C\u002b\u002b`, `C\u002b\u002b`},

		// Unterminated string sequence: OSC without ST/BEL.
		// The string-sequence pattern fails; the two-byte ESC fallback
		// strips ESC] as a pair, leaving the payload as plain text.
		{"unterminated OSC", "\x1B]0;evil title", "0;evil title"},

		// Clean text passes through.
		{"plain ASCII", "hello world", "hello world"},
		{"unicode text", "café résumé naïve", "café résumé naïve"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ansi.Strip(tt.input)
			if got != tt.want {
				t.Errorf("Strip(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
