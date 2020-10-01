package wordwrap

import (
	"bytes"
	"strings"
	"unicode"

	term "github.com/MichaelMure/go-term-text"
)

const (
	normal = "\033[0m"
	dimmed = "\033[2m"
	italic = "\033[3m"
)

// WrapString wraps the given string within lim width in characters.
func WrapString(s string, lim uint) string {
	// Initialize a buffer with a slightly larger size to account for breaks
	init := make([]byte, 0, term.Len(s))
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
	allFormattedLines := ""

	for _, line := range lines {
		currentLineIsUnclosedItalic := containsUnclosedBlock(line, italic)
		currentLineIsUnclosedDim := containsUnclosedBlock(line, dimmed)

		if previousLineWasItalic {
			if lineClosesBlock(line, italic) {
				previousLineWasItalic = false
			} else {
				previousLineWasItalic = true
			}
			allFormattedLines += italic + line + normal + "\n"
			continue
		}

		if previousLineWasDim {
			if lineClosesBlock(line, dimmed) {
				previousLineWasDim = false
			} else {
				previousLineWasDim = true
			}
			allFormattedLines += dimmed + line + normal + "\n"
			continue
		}

		if currentLineIsUnclosedItalic || currentLineIsUnclosedDim {
			allFormattedLines += line + normal + "\n"
		} else {
			allFormattedLines += line + "\n"
		}

		previousLineWasItalic = currentLineIsUnclosedItalic
		previousLineWasDim = currentLineIsUnclosedDim

	}
	wrappedWithRemovedTrailingNewline := strings.TrimSuffix(allFormattedLines, "\n")
	return wrappedWithRemovedTrailingNewline
}

func lineClosesBlock(line string, block string) bool {
	numberOfOpenBlocks := strings.Count(line, block)
	numberOfClosedBlocks := strings.Count(line, normal)
	return numberOfOpenBlocks < numberOfClosedBlocks
}

func containsUnclosedBlock(line string, block string) bool {
	numberOfOpenBlocks := strings.Count(line, block)
	numberOfClosedBlocks := strings.Count(line, normal)
	return numberOfOpenBlocks > numberOfClosedBlocks
}

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
