package article

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// A MediaWiki infobox is a sidebar card, not a data table: the template
// marks every fact as a th.infobox-label / td.infobox-data pair. Flattened
// through the generic table path it renders as a wall of truncated rows
// littered with image captions and a duplicate of the page title, so tables
// carrying those paired classes become a blockInfobox instead — the caption
// as its title, one row per pair, image and header rows dropped. The classes
// must ride through readability's attribute stripping to be seen here;
// fetch.go registers them in ClassesToPreserve.
var infoboxPreservedClasses = []string{"infobox-label", "infobox-data"}

func (p *domParser) parseInfobox(n *html.Node) bool {
	var rows [][]string

	title := ""

	for c := range n.Descendants() {
		if c.Type != html.ElementNode {
			continue
		}

		switch nodeAtom(c) {
		case atom.Caption:
			if c.Parent == n {
				title = cellText(c)
			}

		case atom.Tr:
			if label, data := infoboxPair(c); label != "" && data != "" {
				rows = append(rows, []string{label, data})
			}
		}
	}

	if len(rows) == 0 {
		return false
	}

	p.blocks = append(p.blocks, block{kind: blockInfobox, rows: rows, text: title})

	return true
}

func infoboxPair(tr *html.Node) (label, data string) {
	for c := range tr.ChildNodes() {
		if c.Type != html.ElementNode {
			continue
		}

		switch {
		case nodeAtom(c) == atom.Th && hasClass(c, "infobox-label"):
			label = cellText(c)

		case nodeAtom(c) == atom.Td && hasClass(c, "infobox-data"):
			data = cellText(c)
		}
	}

	return label, data
}
