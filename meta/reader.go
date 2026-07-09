package meta

// ReaderMode is the block above a reader-mode article: the story link over
// the byline and the reader-mode label on the left, ID and score on the
// right.
func ReaderMode(d Data) Block {
	return Block{body: func(width int) string {
		contentWidth := width - paddingSize

		return urlLine(d.URL, d.URL, contentWidth) +
			columns(contentWidth,
				byline(d.Author, d.TimeAgo, d.NerdFonts)+"\n"+readerModeLabel(d.NerdFonts),
				idLabel(d.ID, d.NerdFonts)+"\n"+scoreLabel(d.Points, d.NerdFonts))
	}}
}

// ReaderModeURL is the block for reading a bare URL (`clx url`): just the
// link and the reader-mode label — there is no story behind it.
func ReaderModeURL(url string, nerdFonts bool) Block {
	return Block{body: func(width int) string {
		return urlLine(url, url, width-paddingSize) + readerModeLabel(nerdFonts)
	}}
}
