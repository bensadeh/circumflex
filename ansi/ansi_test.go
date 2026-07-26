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
		{"SGR color", "\x1B[31mhello\x1B[0m", "hello"},
		{"SGR bold+color", "\x1B[1;34mblue\x1B[0m", "blue"},
		{"cursor movement", "\x1B[2Aup\x1B[5Bdown", "updown"},
		{"erase display", "\x1B[2J\x1B[Hhome", "home"},

		{"8-bit CSI", "\xC2\x9B31mred\xC2\x9B0m", "red"},

		{"OSC window title BEL", "\x1B]0;evil title\x07safe", "safe"},
		{"OSC window title ST", "\x1B]0;evil title\x1B\\safe", "safe"},
		{"OSC hyperlink", "\x1B]8;;https://evil.com\x07click\x1B]8;;\x07", "click"},

		{"DCS 7-bit", "\x1BPq#0;2;0;0;0\x1B\\ok", "ok"},

		{"APC 7-bit", "\x1B_payload\x1B\\text", "text"},
		{"PM 7-bit", "\x1B^message\x1B\\text", "text"},
		{"SOS 7-bit", "\x1BXstring\x1B\\text", "text"},

		{"SS2", "\x1BNA ok", " ok"},
		{"SS3", "\x1BOB ok", " ok"},

		{"nF DEC graphics charset", "\x1B(0text\x1B(B", "text"},
		{"nF DECALN fill screen", "\x1B#8text", "text"},
		{"nF S7C1T", "\x1B Ftext", "text"},
		{"nF S8C1T", "\x1B Gtext", "text"},

		{"ESC save cursor", "\x1B7text\x1B8", "text"},
		{"ESC reverse index", "\x1BMline", "line"},

		{"bare ESC at end", "text\x1B", "text"},
		{"ESC + printable", "a\x1Bb", "a"},
		{"ESC SP b (nF)", "a\x1B b", "a"},

		// Controls are stripped, not simulated: backspace doesn't delete prior chars.
		{"backspace", "ab\x08\x08cd", "abcd"},
		{"BEL standalone", "hello\x07world", "helloworld"},
		{"DEL", "ok\x7Ftext", "oktext"},
		{"NUL", "a\x00b", "ab"},
		{"FS GS RS US", "a\x1C\x1D\x1E\x1Fb", "ab"},
		{"raw form feed", "text\fmore", "textmore"},

		{"tab preserved", "a\tb", "a\tb"},
		{"newline preserved", "a\nb", "a\nb"},

		// Tab and newline are layout; a carriage return is a control like any
		// other, drawing back over the line it shares.
		{"CR removed", "a\rb", "ab"},

		// Bidi overrides are a field concern; content keeps its directional runs.
		{"bidi override kept", "a\u202Eb", "a\u202Eb"},

		// A payload the terminator search once missed for holding non-ASCII,
		// leaving the whole sequence behind as visible text.
		{"OSC with unicode payload", "\x1B]0;café\x07safe", "safe"},
		{"OSC 8 with unicode target", "\x1B]8;;https://é.com\x1B\\link", "link"},

		// Deleting the inner sequence splices its neighbours into a live one,
		// which the repeated pass removes rather than leave as debris.
		{"spliced sequence", "\x1B\x1B[0m[31mhi", "hi"},

		// Regression for #201: pre-decode stripping used to corrupt `\\func`
		// (JSON for `\func`) by treating the inner `\f` as a short escape.
		{"literal backslash f", `\func inputs`, `\func inputs`},
		{"literal backslash b", `\begin{document}`, `\begin{document}`},
		{"double backslash", `\\path\\to\\file`, `\\path\\to\\file`},

		// Unterminated OSC: the string-sequence branch fails to match, so the
		// two-byte ESC fallback strips ESC] as a pair and leaves the payload.
		{"unterminated OSC", "\x1B]0;evil title", "0;evil title"},

		// Bare C1 runes with no sequence to anchor them are still controls
		// on terminals that decode C1 from UTF-8.
		{"trailing C1 CSI", "a\xC2\x9B", "a"},
		{"unterminated C1 OSC", "a\xC2\x9D0;evil", "a0;evil"},
		{"stray C1 NEL and RI", "a\xC2\x85\xC2\x8Db", "ab"},

		// Removing the C0 byte must not splice the invalid bytes around it
		// into a valid C1 control rune.
		{"invalid UTF-8 splice to CSI", "\xc2\x11\x9b", "��"},
		{"invalid UTF-8 splice to NEL", "\xc2\x11\x85", "��"},

		{"plain ASCII", "hello world", "hello world"},
		{"unicode text", "café résumé naïve", "café résumé naïve"},
		{"invalid UTF-8 replaced", "ok\xffbytes", "ok�bytes"},
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

func TestNeutralize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"SGR sequence", "\x1B[31mred\x1B[0m", "␛[31mred␛[0m"},
		{"OSC window title", "\x1B]0;title\x07after", "␛]0;title␇after"},
		{"C0 pictures", "a\x00\x08b", "a␀␈b"},
		{"DEL", "a\x7Fb", "a␡b"},

		// A C1 rune renders as its 7-bit ␛-pair equivalent, so an 8-bit
		// dump reads like a 7-bit one.
		{"8-bit CSI", "\xC2\x9B31m", "␛[31m"},
		{"8-bit OSC", "\xC2\x9D0;t", "␛]0;t"},

		{"kept whitespace", "a\tb\nc", "a\tb\nc"},

		// A carriage return would draw back over the line it shares.
		{"CR pictured", "harmless\rEnter password:", "harmless␍Enter password:"},
		{"notation untouched", `\x1b[31m and ESC[0m`, `\x1b[31m and ESC[0m`},
		{"invalid UTF-8 replaced", "ok\xffbytes", "ok�bytes"},
		{"invalid UTF-8 around C0", "\xc2\x11\x9b", "�␑�"},
		{"plain text", "hello world", "hello world"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ansi.Neutralize(tt.input)
			if got != tt.want {
				t.Errorf("Neutralize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestField(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"escapes removed", "\x1B[31mred\x1B[0m", "red"},

		// A field is one row of a fixed layout: a newline would forge a row of
		// its own and a carriage return would draw over the row already there.
		{"newline folded", "safe title\n  fake story row", "safe title fake story row"},
		{"CR folded", "safe title\rfake story row", "safe title fake story row"},
		{"tab folded", "a\tb", "a b"},
		{"edges trimmed", "  padded  ", "padded"},
		{"runs collapsed", "a   b", "a b"},

		// Without this the URL reads as example.com/exe.png on screen while
		// the link opens example.com/gnp.exe.
		{"bidi override removed", "example.com/\u202Egnp.exe", "example.com/gnp.exe"},
		{"bidi isolate removed", "\u2066example.com\u2069", "example.com"},

		{"plain text untouched", "Ask HN: what are you working on?", "Ask HN: what are you working on?"},
		{"unicode text untouched", "café résumé naïve", "café résumé naïve"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ansi.Field(tt.input)
			if got != tt.want {
				t.Errorf("Field(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHyperlink_StripsControlCharsFromURL(t *testing.T) {
	// A BEL in the target would close the OSC 8 sequence early and inject
	// whatever follows; the sink strips it regardless of caller.
	got := ansi.Hyperlink("https://ok.com/\x07evil", "label")

	want := "\x1b]8;;https://ok.com/evil\x1b\\label\x1b]8;;\x1b\\"
	if got != want {
		t.Errorf("Hyperlink() = %q, want %q", got, want)
	}
}
