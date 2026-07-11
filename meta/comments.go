package meta

// CommentSection is the block above a story's comments: one row with the
// byline on the left and the score ending on the block's right edge, the
// story's own text underneath when it has one, and the story link as the
// block's last row. The comment counts live in the footer, not here. A
// submission can carry both text and a link; a faint rule above the link
// keeps it from reading as the text's closing paragraph.
func CommentSection(d Data) Block {
	return Block{body: func(width int) string {
		contentWidth := ContentWidth(width)

		body := columns(contentWidth,
			byline(d.Author, d.TimeAgo, d.NerdFonts),
			scoreLabel(d.Points, d.NerdFonts))

		if d.RootComment != "" {
			body += "\n\n" + d.RootComment
		}

		if url := urlRow(d.URL, d.Domain, contentWidth, d.NerdFonts); url != "" {
			if d.RootComment != "" {
				body += "\n" + divider(contentWidth) + "\n" + url
			} else {
				body += "\n\n" + url
			}
		}

		return body
	}}
}
