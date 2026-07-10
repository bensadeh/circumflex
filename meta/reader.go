package meta

// ReaderMode is the block above a reader-mode article: the byline, the
// reader-mode label, and the score stacked flush left, and the story link as
// the block's last row.
func ReaderMode(d Data) Block {
	return Block{body: func(width int) string {
		contentWidth := ContentWidth(width)

		body := stack(contentWidth,
			byline(d.Author, d.TimeAgo, d.NerdFonts),
			readerModeLabel(d.NerdFonts),
			scoreLabel(d.Points, d.NerdFonts))

		if url := urlRow(d.URL, d.URL, contentWidth, d.NerdFonts); url != "" {
			body += "\n\n" + url
		}

		return body
	}}
}

// ReaderModeURL is the block for reading a bare URL (`clx url`): just the
// reader-mode label and the link — there is no story behind it.
func ReaderModeURL(url string, nerdFonts bool) Block {
	return Block{body: func(width int) string {
		return readerModeLabel(nerdFonts) + "\n\n" + urlRow(url, url, ContentWidth(width), nerdFonts)
	}}
}
