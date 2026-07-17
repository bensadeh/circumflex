package article

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// roleImageLabel returns the accessible description of a graphic that declares
// role="img", whose subtree ARIA defines as presentational — CSS bar charts,
// SVG plots and the like, unrenderable as text. The shape guard keeps inline
// uses (emoji spans) flowing as text.
func roleImageLabel(n *html.Node) string {
	if attr(n, "role") != "img" || !graphicShaped(n) {
		return ""
	}

	return strings.TrimSpace(attr(n, "aria-label"))
}

// graphicShaped separates drawn graphics — an svg, a wrapper holding one, or a
// construct built from block elements — from inline role="img" idioms like
// <span role="img" aria-label="tada">🎉</span>.
func graphicShaped(n *html.Node) bool {
	if nodeAtom(n) == atom.Svg || hasBlockDescendant(n) {
		return true
	}

	return descendantElement(n, atom.Svg) != nil
}

// normalizeRoleImages runs before readability: SVG charts and their wrappers
// hold no paragraph text, so readability deletes them wholesale, description
// and all. Injecting the aria-label as a paragraph gives the node text weight
// to survive on; the walker renders the label from the attribute and skips the
// subtree, so the injected copy never shows. A bare <svg role="img"> becomes a
// labeled wrapper of the same shape, since readability strips svg regardless.
func normalizeRoleImages(root *html.Node) {
	var graphics []*html.Node

	for n := range root.Descendants() {
		if n.Type == html.ElementNode && roleImageLabel(n) != "" {
			graphics = append(graphics, n)
		}
	}

	for _, n := range graphics {
		label := strings.TrimSpace(attr(n, "aria-label"))

		if nodeAtom(n) != atom.Svg {
			n.AppendChild(labelParagraph(label))

			continue
		}

		if n.Parent == nil {
			continue
		}

		div := &html.Node{
			Type:     html.ElementNode,
			DataAtom: atom.Div,
			Data:     "div",
			Attr: []html.Attribute{
				{Key: "role", Val: "img"},
				{Key: "aria-label", Val: label},
			},
		}
		div.AppendChild(labelParagraph(label))

		n.Parent.InsertBefore(div, n)
		n.Parent.RemoveChild(n)
	}
}

func labelParagraph(label string) *html.Node {
	p := &html.Node{Type: html.ElementNode, DataAtom: atom.P, Data: "p"}
	p.AppendChild(&html.Node{Type: html.TextNode, Data: label})

	return p
}
