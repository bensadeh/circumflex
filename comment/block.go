package comment

// The comment body pipeline mirrors the article package: Parse turns HN's
// comment HTML into typed blocks holding semantic spans, and RenderBlocks maps
// each block and span kind to exactly one visual decision. No ANSI exists
// before the renderer.

type blockKind int

const (
	blockParagraph blockKind = iota
	blockQuote
	blockCode
	blockRemoved
)

// Removed reports whether a comment body is one of HN's placeholder markers:
// text withheld by deletion, moderation, or the author's delay setting.
// ToThread prunes removed comments whose subtrees are removed throughout;
// ones with a surviving reply stay to anchor the thread.
func Removed(content string) bool {
	return content == "[deleted]" || content == "[flagged]" || content == "[delayed]"
}

type spanFormat int

const (
	spanPlain spanFormat = iota
	spanItalic
	spanLink
	spanCodeInline
	spanMention
	spanVariable
	spanReference
	spanAbbreviation
	spanYCLabel
)

// Block is one paragraph-level unit of a comment body. Empty blocks are
// meaningful: HN separates paragraphs with lone <p> tags, and an empty
// paragraph still occupies a join slot between its neighbors.
type Block struct {
	kind  blockKind
	spans []span // blockParagraph, blockQuote
	text  string // blockCode: verbatim, newlines preserved; blockRemoved: the marker
	lang  string // blockCode: guessed language, empty when unrecognized

	// hlOut caches the renderer-produced chroma output — the one exception
	// to no-ANSI-before-the-renderer, because tokenizing every block again
	// on every resize step is measurable jank on large threads.
	hlOut  string
	hlDone bool
}

// span is a run of paragraph text with one semantic role. Tokenizers split
// plain spans into finer ones, so a span's role never changes once assigned.
type span struct {
	text   string
	format spanFormat
	href   string // spanLink: OSC 8 target
}
