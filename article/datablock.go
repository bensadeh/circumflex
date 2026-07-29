package article

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// normalizeDataBlocks flattens BBC News articles, keyed by the data-block
// attribute their renderer stamps on each section (text, image, subheadline,
// …). Every section sits in its own chain of styled-component wrapper divs,
// so readability scores each chain in isolation: the longest text section
// becomes top candidate, its siblings fall under the join threshold, and the
// rest of the article — the other text sections and every figure — is
// silently dropped. Splicing each section's content directly into the shared
// container turns the page back into the single article-shaped subtree
// readability expects.
func normalizeDataBlocks(root *html.Node) {
	for _, section := range dataBlockSections(root) {
		flattenSection(section)
	}
}

// dataBlockSections returns the members of sibling groups of data-block
// sections. Requiring a group — several labeled siblings, one of them text —
// keeps the rewrite off pages that merely reuse the attribute name.
func dataBlockSections(root *html.Node) []*html.Node {
	var sections []*html.Node

	for n := range root.Descendants() {
		if n.Type != html.ElementNode {
			continue
		}

		var group []*html.Node

		hasText := false

		for c := range n.ChildNodes() {
			if c.Type != html.ElementNode || attr(c, "data-block") == "" {
				continue
			}

			group = append(group, c)

			if attr(c, "data-block") == "text" {
				hasText = true
			}
		}

		if len(group) >= 2 && hasText {
			sections = append(sections, group...)
		}
	}

	return sections
}

// flattenSection replaces a section with its content: wrapper divs holding
// nothing but a single element are peeled away, and when what remains is
// still a div — a container of paragraphs — its children are spliced in
// directly.
func flattenSection(section *html.Node) {
	parent := section.Parent
	if parent == nil {
		return
	}

	content := section
	for {
		child := loneElementChild(content)
		if child == nil {
			break
		}

		content = child
	}

	if nodeAtom(content) == atom.Div {
		for content.FirstChild != nil {
			c := content.FirstChild
			content.RemoveChild(c)
			parent.InsertBefore(c, section)
		}
	} else {
		content.Parent.RemoveChild(content)
		parent.InsertBefore(content, section)
	}

	parent.RemoveChild(section)
}

// loneElementChild returns the single element inside a wrapper div carrying
// no text of its own, and nil for anything else.
func loneElementChild(n *html.Node) *html.Node {
	if nodeAtom(n) != atom.Div {
		return nil
	}

	var only *html.Node

	for c := range n.ChildNodes() {
		switch c.Type {
		case html.ElementNode:
			if only != nil {
				return nil
			}

			only = c

		case html.TextNode:
			if strings.TrimSpace(c.Data) != "" {
				return nil
			}
		}
	}

	return only
}
