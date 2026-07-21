package help

import (
	"strings"

	"github.com/bensadeh/circumflex/frame"
	"github.com/bensadeh/circumflex/style"

	"charm.land/lipgloss/v2"
)

const (
	cellGap          = 4
	keyDescGap       = 2
	singleColumnMin  = 60
	minPanelWidth    = 24
	sectionSeparator = "\n\n"
	ellipsis         = "…"
)

type entry struct {
	key    string
	desc   string
	rawKey bool
}

type section struct {
	title  string
	groups [][]entry
}

type keyList struct {
	sections []*section
}

func (k *keyList) addSection(title string) *section {
	s := &section{title: title, groups: [][]entry{nil}}
	k.sections = append(k.sections, s)

	return s
}

func (s *section) addKey(key, desc string) {
	s.append(entry{key: key, desc: desc})
}

func (s *section) addLabel(key, desc string) {
	s.append(entry{key: key, desc: desc, rawKey: true})
}

func (s *section) addBreak() {
	if len(s.groups[len(s.groups)-1]) == 0 {
		return
	}

	s.groups = append(s.groups, nil)
}

func (s *section) append(e entry) {
	last := len(s.groups) - 1
	s.groups[last] = append(s.groups[last], e)
}

func (s *section) hasEntries() bool {
	for _, g := range s.groups {
		if len(g) > 0 {
			return true
		}
	}

	return false
}

func (k *keyList) print(width int) string {
	if width < minPanelWidth {
		return ""
	}

	parts := make([]string, 0, len(k.sections))
	keyWidth := k.maxKeyWidth()

	for _, s := range k.sections {
		if !s.hasEntries() && s.title == "" {
			continue
		}

		parts = append(parts, renderPanel(s, width, keyWidth))
	}

	return strings.Join(parts, sectionSeparator)
}

func (k *keyList) maxKeyWidth() int {
	m := 0

	for _, s := range k.sections {
		for _, g := range s.groups {
			if w := maxKeyWidth(g); w > m {
				m = w
			}
		}
	}

	return m
}

func renderPanel(s *section, panelWidth, keyWidth int) string {
	innerWidth := max(frame.ContentWidth(panelWidth), 1)

	grids := make([]string, 0, len(s.groups))

	for _, g := range s.groups {
		if len(g) == 0 {
			continue
		}

		grids = append(grids, renderGrid(g, innerWidth, keyWidth))
	}

	title := s.title
	if title != "" {
		title = lipgloss.NewStyle().Bold(true).Foreground(style.HeaderTertiary()).Render(title)
	}

	rows := []string{frame.OpeningRule(title, nil, panelWidth)}

	for _, line := range bodyLines(strings.Join(grids, "\n")) {
		rows = append(rows, frame.Row(line, panelWidth))
	}

	return frame.Join(append(rows, frame.ClosingRule(nil, panelWidth)), panelWidth)
}

func bodyLines(body string) []string {
	lines := strings.Split(body, "\n")

	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return lines
}

func renderGrid(entries []entry, width, keyWidth int) string {
	if len(entries) == 0 {
		return ""
	}

	if width < singleColumnMin {
		return renderColumn(entries, width, keyWidth)
	}

	cellWidth := (width - cellGap) / 2
	half := (len(entries) + 1) / 2

	var b strings.Builder

	for i := range half {
		left := renderCell(entries[i], keyWidth, cellWidth)

		var right string
		if j := i + half; j < len(entries) {
			right = renderCell(entries[j], keyWidth, cellWidth)
		} else {
			right = strings.Repeat(" ", cellWidth)
		}

		b.WriteString(left)
		b.WriteString(strings.Repeat(" ", cellGap))
		b.WriteString(right)
		b.WriteString("\n")
	}

	return b.String()
}

func renderColumn(entries []entry, width, keyWidth int) string {
	var b strings.Builder

	for _, e := range entries {
		b.WriteString(renderCell(e, keyWidth, width))
		b.WriteString("\n")
	}

	return b.String()
}

func renderCell(e entry, keyWidth, cellWidth int) string {
	key := e.key
	if !e.rawKey {
		key = style.Bold(key)
	}

	keyPad := max(keyWidth-lipgloss.Width(e.key), 0)
	keyCell := strings.Repeat(" ", keyPad) + key
	gap := strings.Repeat(" ", keyDescGap)

	descWidth := cellWidth - keyWidth - keyDescGap
	desc := truncateToWidth(e.desc, descWidth)
	descPad := max(descWidth-lipgloss.Width(desc), 0)

	return keyCell + gap + desc + strings.Repeat(" ", descPad)
}

func truncateToWidth(s string, w int) string {
	if w <= 0 {
		return ""
	}

	if lipgloss.Width(s) <= w {
		return s
	}

	if strings.ContainsRune(s, '\x1b') {
		return s
	}

	if w == 1 {
		return ellipsis
	}

	runes := []rune(s)

	return string(runes[:w-1]) + ellipsis
}

func maxKeyWidth(entries []entry) int {
	m := 0

	for _, e := range entries {
		if w := lipgloss.Width(e.key); w > m {
			m = w
		}
	}

	return m
}
