package meta

// ReaderMode is the block above a reader-mode article: the byline in the
// frame's opening rule with the comment count and score closing it, and the
// story link inside the frame. The reader-mode label is the footer's
// business, not the block's.
func ReaderMode(d Data) Block {
	return Block{
		title:  byline(d.Author, d.TimeAgo, d.NerdFonts),
		labels: statLabels(d),
		body: func(width int) string {
			return urlRow(d.URL, d.URL, ContentWidth(width))
		},
	}
}

// ReaderModeURL is the block for reading a bare URL (`clx url`): just the
// link in an untitled frame — there is no story behind it.
func ReaderModeURL(url string) Block {
	return Block{body: func(width int) string {
		return urlRow(url, url, ContentWidth(width))
	}}
}
