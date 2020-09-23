// wordwrap retreived from github.com/eidolon/wordwrap. Modified
// to not include hidden characters when calculating word and line length.
package wordwrap

import (
	"bytes"
	"strings"

	term "github.com/MichaelMure/go-term-text"
)

// Indent a string with the given prefix at the start of either the first, or all lines.
//
//  input     - The input string to indent.
//  prefix    - The prefix to add.
//  prefixAll - If true, prefix all lines with the given prefix.
//
// Example usage:
//
//  indented := wordwrap.Indent("Hello\nWorld", "-", true)
func Indent(input string, prefix string, prefixAll bool) string {
	lines := strings.Split(input, "\n")
	prefixLen := term.Len(prefix)
	result := make([]string, len(lines))

	for i, line := range lines {
		if prefixAll || i == 0 {
			result[i] = prefix + line
		} else {
			result[i] = strings.Repeat(" ", prefixLen) + line
		}
	}

	return strings.Join(result, "\n")
}

// WrapperFunc takes a given input string, and returns some wrapped output. The wrapping may
// be altered by currying the wrapper function.
type WrapperFunc func(string) string

// Wrapper creates a curried wrapper function (see WrapperFunc) with the given options applied to
// it. Create a WrapperFunc and store it is a variable, then re-use it elsewhere.
//
//  limit      - The maximum number of characters for a line.
//  breakWords - Whether or not to break long words onto new lines.
//
// Example usage:
//
//  wrapper := wordwrap.Wrapper(10, false)
//  wrapped := wrapper("This string would be split onto several new lines")
func Wrapper(limit int, breakWords bool) WrapperFunc {
	if limit < 1 {
		panic("Wrapper limit cannot be less than 1.")
	}

	return func(input string) string {
		var wrapped string

		// Split string into array of words
		words := strings.Fields(input)

		if len(words) == 0 {
			return wrapped
		}

		remaining := limit

		if breakWords {
			words = doBreakWords(words, limit)
		}

		for _, word := range words {
			if term.Len(word)+1 > remaining {
				if term.Len(wrapped) > 0 {
					wrapped += "\n"
				}

				wrapped += word
				remaining = limit - term.Len(word)
			} else {
				if term.Len(wrapped) > 0 {
					wrapped += " "
				}

				wrapped += word
				remaining = remaining - (term.Len(word) + 1)
			}
		}

		return wrapped
	}
}

// Break up any words in a given array of words that exceed the given limit.
func doBreakWords(words []string, limit int) []string {
	var result []string

	for _, word := range words {
		if term.Len(word) > limit {
			var parts []string
			var partBuf bytes.Buffer

			for _, char := range word {
				atLimit := partBuf.Len() == limit

				if atLimit {
					parts = append(parts, partBuf.String())

					partBuf.Reset()
				}

				partBuf.WriteRune(char)
			}

			if partBuf.Len() > 0 {
				parts = append(parts, partBuf.String())
			}

			for _, part := range parts {
				result = append(result, part)
			}
		} else {
			result = append(result, word)
		}
	}

	return result
}
