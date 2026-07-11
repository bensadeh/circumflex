package meta

// CommentSection is the block above a story's comments: the byline in the
// frame's opening rule with the score closing it, the story's own text
// inside the frame when it has one, and the story link as the frame's last
// row. The comment counts live in the footer, not here. A submission can
// carry both text and a link; a faint rule above the link keeps it from
// reading as the text's closing paragraph.
func CommentSection(d Data) Block {
	return Block{
		title: byline(d.Author, d.TimeAgo, d.NerdFonts),
		score: scoreLabel(d.Points, d.NerdFonts),
		body: func(width int) string {
			contentWidth := ContentWidth(width)
			url := urlRow(d.URL, d.Domain, contentWidth, d.NerdFonts)

			switch {
			case d.RootComment != "" && url != "":
				return d.RootComment + "\n" + divider(contentWidth) + "\n" + url
			case d.RootComment != "":
				return d.RootComment
			default:
				return url
			}
		},
	}
}
