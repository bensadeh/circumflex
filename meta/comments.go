package meta

// CommentSection is the block above a story's comments: the byline, the
// comment counts, and the score stacked flush left, the story's own text
// underneath when it has one, and the story link as the block's last row. A
// submission can carry both text and a link; a faint rule above the link
// keeps it from reading as the text's closing paragraph.
func CommentSection(d Data) Block {
	return Block{body: func(width int) string {
		contentWidth := ContentWidth(width)

		body := stack(contentWidth,
			byline(d.Author, d.TimeAgo, d.NerdFonts),
			commentsLabel(d.CommentsCount, d.NerdFonts)+newCommentsLabel(d.NewComments, d.NerdFonts),
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
