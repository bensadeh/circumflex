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
	blockDeleted
)

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
	text  string // blockCode: verbatim, newlines preserved
}

// span is a run of paragraph text with one semantic role. Tokenizers split
// plain spans into finer ones, so a span's role never changes once assigned.
type span struct {
	text   string
	format spanFormat
	href   string // spanLink: OSC 8 target
}
