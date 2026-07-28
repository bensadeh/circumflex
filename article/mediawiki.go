package article

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// normalizeMediaWiki rewrites MediaWiki markup that readability mishandles,
// keyed by the wiki's own class names so it covers any MediaWiki site.
//
// Heading wrappers (<div class="mw-heading"><h2>…</h2><span
// class="mw-editsection">[edit]</span></div>) score as link-heavy chrome:
// readability deletes the wrapper outright when the heading text is short, so
// "See also" and "References" vanish while their sections stay and the
// stop-at-heading site rules never fire; where the wrapper survives, the edit
// link splits off into a literal "[edit]" paragraph.
//
// Math wrappers (<span class="mwe-math-element">) hold each formula twice, as
// MathML inside a display:none accessibility span and as an <img> fallback.
// Readability drops hidden nodes, so only the image reaches the parser and
// every formula renders as an image block torn out of its sentence. Keeping
// just the MathML child lets it render as inline TeX text.
func normalizeMediaWiki(root *html.Node) {
	var mathWrappers, editLinks, headingWrappers []*html.Node

	for n := range root.Descendants() {
		if n.Type != html.ElementNode {
			continue
		}

		switch {
		case hasClass(n, "mwe-math-element"):
			mathWrappers = append(mathWrappers, n)

		case hasClass(n, "mw-editsection"):
			editLinks = append(editLinks, n)

		case hasClass(n, "mw-heading"):
			headingWrappers = append(headingWrappers, n)
		}
	}

	for _, wrapper := range mathWrappers {
		if math := descendantElement(wrapper, atom.Math); math != nil && wrapper.Parent != nil {
			math.Parent.RemoveChild(math)
			wrapper.Parent.InsertBefore(math, wrapper)
			wrapper.Parent.RemoveChild(wrapper)
		}
	}

	for _, link := range editLinks {
		if link.Parent != nil {
			link.Parent.RemoveChild(link)
		}
	}

	for _, wrapper := range headingWrappers {
		unwrap(wrapper)
	}
}

func descendantElement(n *html.Node, a atom.Atom) *html.Node {
	for c := range n.Descendants() {
		if c.Type == html.ElementNode && nodeAtom(c) == a {
			return c
		}
	}

	return nil
}

func unwrap(n *html.Node) {
	if n.Parent == nil {
		return
	}

	for n.FirstChild != nil {
		c := n.FirstChild
		n.RemoveChild(c)
		n.Parent.InsertBefore(c, n)
	}

	n.Parent.RemoveChild(n)
}
