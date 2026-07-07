package article

import (
	"strings"

	"github.com/bensadeh/circumflex/ansi"
)

// parseTextBlocks converts a text/plain document into verbatim blocks, one
// per blank-line-separated chunk. Line structure is preserved: documents like
// release notes and RFCs are already hand-wrapped and indented.
func parseTextBlocks(text string) []block {
	text = ansi.Strip(text)
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\t", strings.Repeat(" ", 8))

	var blocks []block

	for chunk := range strings.SplitSeq(text, "\n\n") {
		lines := strings.Split(chunk, "\n")
		for i, line := range lines {
			lines[i] = strings.TrimRight(line, " ")
		}

		chunk = strings.Trim(strings.Join(lines, "\n"), "\n")
		if strings.TrimSpace(chunk) == "" {
			continue
		}

		blocks = append(blocks, block{kind: blockVerbatim, text: chunk})
	}

	return blocks
}
