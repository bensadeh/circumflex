package layout

import (
	"strings"

	xansi "github.com/charmbracelet/x/ansi"
)

// FooterSections spreads footer labels across width: the first sits flush
// left, the last ends flush right, and the space left over is shared
// equally by the gaps between them. Labels that together overrun the width
// keep a single space between them instead; the caller truncates.
func FooterSections(width int, sections ...string) string {
	gaps := max(1, len(sections)-1)

	slack := width
	for _, s := range sections {
		slack -= xansi.StringWidth(s)
	}

	var b strings.Builder

	cur, prefix := 0, 0

	for i, s := range sections {
		if s == "" {
			continue
		}

		pad := prefix + slack*i/gaps - cur
		if b.Len() > 0 {
			pad = max(pad, 1)
		} else {
			pad = max(pad, 0)
		}

		b.WriteString(strings.Repeat(" ", pad) + s)

		w := xansi.StringWidth(s)
		cur += pad + w
		prefix += w
	}

	return b.String()
}
