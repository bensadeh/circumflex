package article

import (
	"bytes"
	nurl "net/url"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// restoreLeadImage prepends the original document's featured image when
// readability dropped it: WordPress themes commonly render the hero (marked
// wp-post-image) in the post header, outside the extracted article container.
func restoreLeadImage(body []byte, blocks []block) []block {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return blocks
	}

	img := featuredImage(doc)
	if img == nil {
		return blocks
	}

	src := imageSrc(img)
	if src == "" {
		return blocks
	}

	key := imageSourceKey(src)
	for i := range blocks {
		if blocks[i].kind == blockImage && imageSourceKey(blocks[i].imageURL) == key {
			return blocks
		}
	}

	return append([]block{imageBlock(altText(img), src, imageDisplayWidth(img))}, blocks...)
}

func featuredImage(root *html.Node) *html.Node {
	for c := range root.Descendants() {
		if c.Type == html.ElementNode && nodeAtom(c) == atom.Img &&
			slices.Contains(strings.Fields(attr(c, "class")), "wp-post-image") {
			return c
		}
	}

	return nil
}

// wpSizeVariant matches WordPress size-suffixed uploads (photo-300x200.jpg,
// photo-scaled.jpg), which all show the same underlying image.
var wpSizeVariant = regexp.MustCompile(`-(?:\d+x\d+|scaled)(\.\w+)$`)

// imageSourceKey normalizes an image URL so size variants of the same upload
// compare equal.
func imageSourceKey(raw string) string {
	u, err := nurl.Parse(raw)
	if err != nil {
		return raw
	}

	return u.Host + wpSizeVariant.ReplaceAllString(u.Path, "$1")
}
