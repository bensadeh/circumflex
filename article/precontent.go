package article

import (
	"slices"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// preservePreContent flattens the markup inside <pre> blocks to the text
// preText will read anyway, before readability can misjudge it as page
// chrome. Doc generators and highlighters annotate code in chrome's own
// vocabulary — pkgsite wraps comments in spans classed "comment", ids field
// lines after their type (RelatedInformation.Pos), and links every type
// name — so readability deletes the spans as unlikely candidates and removes
// link-heavy declaration wrappers wholesale. Attributes and anchors inside a
// pre describe the code, never the page. Runs after preserveCodeLang, which
// reads language classes on a pre's <code> children before they go.
func preservePreContent(doc *html.Node) {
	var pres []*html.Node

	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && nodeAtom(n) == atom.Pre {
			pres = append(pres, n)
		}
	}

	for _, pre := range pres {
		var elements []*html.Node

		for d := range pre.Descendants() {
			if d.Type == html.ElementNode {
				elements = append(elements, d)
			}
		}

		for _, el := range elements {
			if nodeAtom(el) == atom.A {
				unwrap(el)

				continue
			}

			el.Attr = slices.DeleteFunc(el.Attr, func(a html.Attribute) bool {
				return a.Key == "class" || a.Key == "id"
			})
		}
	}
}
