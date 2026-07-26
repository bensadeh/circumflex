package ansi_test

import (
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/bensadeh/circumflex/ansi"
)

var sanitizerSeeds = []string{
	"",
	"plain text",
	"café résumé naïve 🎉",
	"\x1B[31mred\x1B[0m",
	"\x1B]0;title\x07after",
	"\x1B]8;;https://evil.com\x1B\\click\x1B]8;;\x1B\\",
	"\x1BPq#0;2;0;0;0\x1B\\sixel",
	"\x1B_apc\x1B\\ \x1B^pm\x1B\\ \x1BXsos\x1B\\",
	"\xC2\x9B31m \xC2\x9D0;t \xC2\x8E \xC2\x9C",
	"\x1B\x1B[0m[31mspliced",
	"\x1B]0;unterminated",
	"\xc2\x11\x9b",
	"ok\xffbytes",
	"a\x00\x08\x7Fb",
	"tab\there\nnewline\rreturn",
	"example.com/\u202Egnp.exe",
	"\u2066isolated\u2069",
	`\x1b[31m notation`,
}

// The explicit directional formatting characters, the set Field removes. The
// weaker marks LRM and RLM stay: they nudge punctuation in mixed-direction
// text without reordering the runs around them.
const bidiOverrides = "\u202A\u202B\u202C\u202D\u202E\u2066\u2067\u2068\u2069"

// Strip clears the terminal's own vocabulary out of untrusted text; the
// escape-sequence pass is only there to keep the leftovers tidy, so the
// guarantee is stated over the runes that survive rather than over any
// sequence the regexes claim to know.
func FuzzStrip(f *testing.F) {
	for _, s := range sanitizerSeeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		got := ansi.Strip(input)

		assertValidUTF8(t, "Strip", input, got)
		assertNoControlsBeyond(t, "Strip", input, got, "\t\n")

		if again := ansi.Strip(got); again != got {
			t.Fatalf("Strip(%q) is not settled: %q became %q", input, got, again)
		}
	})
}

// Field carries Strip's guarantee and adds the two a row of a fixed layout
// depends on: it occupies exactly one line, and it cannot reorder itself.
func FuzzField(f *testing.F) {
	for _, s := range sanitizerSeeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		got := ansi.Field(input)

		assertValidUTF8(t, "Field", input, got)
		assertNoControlsBeyond(t, "Field", input, got, "")

		for _, r := range got {
			if strings.ContainsRune(bidiOverrides, r) {
				t.Fatalf("Field(%q) = %q kept bidi override %U", input, got, r)
			}
		}

		if got != strings.Join(strings.Fields(got), " ") {
			t.Fatalf("Field(%q) = %q is not folded onto one line", input, got)
		}
	})
}

// Neutralize shows the terminal's vocabulary instead of deleting it, so the
// same runes must be gone from its output — replaced by pictures rather than
// by nothing. Tab and newline stay: the layout is built from them.
func FuzzNeutralize(f *testing.F) {
	for _, s := range sanitizerSeeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		got := ansi.Neutralize(input)

		assertValidUTF8(t, "Neutralize", input, got)
		assertNoControlsBeyond(t, "Neutralize", input, got, "\t\n")
	})
}

func assertNoControlsBeyond(t *testing.T, fn, input, got, allowed string) {
	t.Helper()

	for _, r := range got {
		if unicode.IsControl(r) && !strings.ContainsRune(allowed, r) {
			t.Fatalf("%s(%q) = %q kept control %U", fn, input, got, r)
		}
	}
}

func assertValidUTF8(t *testing.T, fn, input, got string) {
	t.Helper()

	if !utf8.ValidString(got) {
		t.Fatalf("%s(%q) = %q is not valid UTF-8", fn, input, got)
	}
}
