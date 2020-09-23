// alt-wordwrap retreived from github.com/mitchellh/go-wordwrap.
// Modified to add ansi codes over several lines.

package wordwrap

import (
	"bytes"
	"strings"
	"unicode"
)

const (
	normal = "\033[0m"
	dimmed = "\033[2m"
	italic = "\033[3m"
)

// WrapString wraps the given string within lim width in characters.
//
// Wrapping is currently naive and only happens at white-space. A future
// version of the library will implement smarter wrapping. This means that
// pathological cases can dramatically reach past the limit, such as a very
// long word.
func WrapString(s string, lim uint) string {
	// Initialize a buffer with a slightly larger size to account for breaks
	init := make([]byte, 0, len(s))
	buf := bytes.NewBuffer(init)

	var current uint
	var wordBuf, spaceBuf bytes.Buffer

	for _, char := range s {
		if char == '\n' {
			if wordBuf.Len() == 0 {
				if current+uint(spaceBuf.Len()) > lim {
					current = 0
				} else {
					current += uint(spaceBuf.Len())
					spaceBuf.WriteTo(buf)
				}
				spaceBuf.Reset()
			} else {
				current += uint(spaceBuf.Len() + wordBuf.Len())
				spaceBuf.WriteTo(buf)
				spaceBuf.Reset()
				wordBuf.WriteTo(buf)
				wordBuf.Reset()
			}
			buf.WriteRune(char)
			current = 0
		} else if unicode.IsSpace(char) {
			if spaceBuf.Len() == 0 || wordBuf.Len() > 0 {
				current += uint(spaceBuf.Len() + wordBuf.Len())
				spaceBuf.WriteTo(buf)
				spaceBuf.Reset()
				wordBuf.WriteTo(buf)
				wordBuf.Reset()
			}

			spaceBuf.WriteRune(char)
		} else {

			wordBuf.WriteRune(char)

			if current+uint(spaceBuf.Len()+wordBuf.Len()) > lim && uint(wordBuf.Len()) < lim {
				buf.WriteRune('\n')
				current = 0
				spaceBuf.Reset()
			}
		}
	}

	if wordBuf.Len() == 0 {
		if current+uint(spaceBuf.Len()) <= lim {
			spaceBuf.WriteTo(buf)
		}
	} else {
		spaceBuf.WriteTo(buf)
		wordBuf.WriteTo(buf)
	}

	formatted := formatWrapped(buf.String())
	return formatted
}

func formatWrapped(wrapped string) string {
	lines := strings.Split(wrapped, "\n")
	previousLineWasItalic := false
	previousLineWasDim := false
	formattedWrapped := ""
	formattedLine := ""

	for _, line := range lines {
		if previousLineWasItalic {
			formattedLine = italic + line + "\n"
			formattedWrapped += formattedLine
		} else if previousLineWasDim {
			formattedLine = dimmed + line + "\n"
			formattedWrapped += formattedLine
		} else {
			formattedLine = line + "\n"
			formattedWrapped += formattedLine
		}

		previousLineWasItalic = containsUnclosedBlock(formattedLine, italic)
		previousLineWasDim = containsUnclosedBlock(formattedLine, dimmed)
	}
	wrappedWithRemovedTrailingNewline := strings.TrimSuffix(formattedWrapped, "\n")
	return wrappedWithRemovedTrailingNewline
}

func containsUnclosedBlock(line string, block string) bool {
	numberOfOpenBlocks := strings.Count(line, block)
	numberOfClosedBlocks := strings.Count(line, normal)
	return numberOfOpenBlocks > numberOfClosedBlocks
}
