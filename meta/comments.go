package meta

// CommentSection is the block above a story's comments: the byline in the
// frame's opening rule with the comment count and score closing it, the
// story's own text inside the frame when it has one, and the story link as
// the frame's last row. A submission can carry both text and a link; a faint
// rule above the link keeps it from reading as the text's closing paragraph.
func CommentSection(d Data) Block {
	return Block{
		title:         byline(d.Author, d.TimeAgo, d.NerdFonts),
		labels:        statLabels(d),
		closingLabels: []string{idLabel(d.ID, d.NerdFonts)},
		body: func(width int) string {
			contentWidth := ContentWidth(width)
			url := urlRow(d.URL, d.Domain, contentWidth)

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
