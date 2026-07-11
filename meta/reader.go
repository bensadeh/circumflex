package meta

// ReaderMode is the block above a reader-mode article: one row with the
// byline on the left and the score ending on the block's right edge, and
// the story link as the block's last row. The reader-mode label is the
// footer's business, not the block's.
func ReaderMode(d Data) Block {
	return Block{body: func(width int) string {
		contentWidth := ContentWidth(width)

		body := columns(contentWidth,
			byline(d.Author, d.TimeAgo, d.NerdFonts),
			scoreLabel(d.Points, d.NerdFonts))

		if url := urlRow(d.URL, d.URL, contentWidth, d.NerdFonts); url != "" {
			body += "\n\n" + url
		}

		return body
	}}
}

// ReaderModeURL is the block for reading a bare URL (`clx url`): just the
// link — there is no story behind it.
func ReaderModeURL(url string, nerdFonts bool) Block {
	return Block{body: func(width int) string {
		return urlRow(url, url, ContentWidth(width), nerdFonts)
	}}
}
