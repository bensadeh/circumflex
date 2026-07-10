package meta

// CommentSection is the block above a story's comments: the story link over
// the byline and comment counts on the left, ID and score on the right, and
// the story's own text underneath when it has one.
func CommentSection(d Data) Block {
	return Block{body: func(width int) string {
		contentWidth := ContentWidth(width)

		body := urlLine(d.URL, d.Domain, contentWidth) +
			columns(contentWidth,
				byline(d.Author, d.TimeAgo, d.NerdFonts)+"\n"+
					commentsLabel(d.CommentsCount, d.NerdFonts)+newCommentsLabel(d.NewComments, d.NerdFonts),
				idLabel(d.ID, d.NerdFonts)+"\n"+scoreLabel(d.Points, d.NerdFonts))

		if d.RootComment != "" {
			body += "\n\n" + d.RootComment
		}

		return body
	}}
}
