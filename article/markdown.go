package article

import (
	"bytes"
	"fmt"
	nurl "net/url"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	htmlrenderer "github.com/yuin/goldmark/renderer/html"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// isMarkdown reports whether the page should be parsed as markdown: served as
// text/markdown, or a .md path under the generic text/plain label GitHub's
// raw endpoints use — unless the body is mislabeled HTML.
func isMarkdown(contentType string, parsedURL *nurl.URL, body []byte) bool {
	if strings.HasPrefix(contentType, "text/markdown") {
		return true
	}

	if !strings.HasPrefix(contentType, "text/plain") {
		return false
	}

	path := strings.ToLower(parsedURL.Path)

	return (strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".markdown")) && !looksLikeHTML(body)
}

// parseMarkdownBlocks converts markdown to HTML and hands it to the block
// parser used for fetched pages, so both formats share one renderer.
// Readability is skipped — there is no chrome to strip — so link targets are
// resolved here instead; image sources resolve later, in fetchImages. A
// leading h1 is the document's own title: it moves to the returned title like
// readability's, instead of duplicating right under the title header.
func parseMarkdownBlocks(body []byte, base *nurl.URL) ([]block, string, error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.Footnote),
		goldmark.WithRendererOptions(htmlrenderer.WithUnsafe()),
	)

	var buf bytes.Buffer
	if err := md.Convert(body, &buf); err != nil {
		return nil, "", fmt.Errorf("could not parse markdown from %s: %w", base.Host, err)
	}

	node, err := html.Parse(&buf)
	if err != nil {
		return nil, "", fmt.Errorf("could not parse page from %s: %w", base.Host, err)
	}

	resolveLinkTargets(node, base)

	blocks := parseBlocks(node)

	if len(blocks) > 0 && blocks[0].kind == blockHeading && blocks[0].level == 1 {
		return blocks[1:], blocks[0].text, nil
	}

	return blocks, "", nil
}

// resolveLinkTargets absolutizes anchor hrefs against base, as readability
// does for fetched pages. Fragment refs (goldmark's footnote links) stay
// relative, and so stay plain text under isLinkableHref.
func resolveLinkTargets(root *html.Node, base *nurl.URL) {
	for n := range root.Descendants() {
		if n.Type != html.ElementNode || n.DataAtom != atom.A {
			continue
		}

		for i, a := range n.Attr {
			if a.Key != "href" || a.Val == "" || strings.HasPrefix(a.Val, "#") {
				continue
			}

			ref, err := nurl.Parse(a.Val)
			if err != nil {
				continue
			}

			n.Attr[i].Val = base.ResolveReference(ref).String()
		}
	}
}
