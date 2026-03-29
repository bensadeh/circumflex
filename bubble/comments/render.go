package comments

import (
	"clx/comment"
	"clx/constants"
	"clx/settings"
	"strings"

	text "github.com/MichaelMure/go-term-text"
)

// renderContext holds the stable parameters for rendering a comment view.
// These values don't change between rebuilds unless the window is resized
// or the model is re-created.
type renderContext struct {
	header         string // pre-computed meta block (raw, before margin wrapping)
	originalPoster string
	firstCommentID int
	config         *settings.Config
	screenWidth    int
	viewportHeight int
	lastVisited    int64
}

// cachedComment stores the pre-rendered output for a single flat comment.
// Populated lazily by renderFromFlat and invalidated on window resize.
type cachedComment struct {
	valid     bool
	sep       string // rendered separator (before the comment body)
	sepLines  int
	body      string // rendered comment body with indentation and margins
	bodyLines int
	fold      string // fold indicator (empty if no descendants)
	foldLines int
}

// renderFromFlat builds the full comment view content from the flat comment
// list, respecting fold state. It returns the rendered content, the number of
// content lines, and line metrics indexed by flat index for navigation.
// This function has no knowledge of focus state — focus highlighting is
// applied separately in View().
//
// The cache slice (indexed by flat index) avoids re-running the expensive
// syntax-highlighting and text-wrapping pipeline on every collapse/expand.
func renderFromFlat(rc renderContext, flat []FlatComment, visible []int, cache []cachedComment) (string, int, []LineMetrics) {
	leftMargin := strings.Repeat(" ", constants.CommentSectionLeftMargin)
	contentWidth := rc.screenWidth - constants.CommentSectionLeftMargin

	// Indent the pre-computed header with left margin.
	header, _ := text.WrapWithPad(rc.header, rc.screenWidth, leftMargin)

	var sb strings.Builder
	sb.WriteString(header)

	lineCount := strings.Count(header, "\n")

	metrics := make([]LineMetrics, len(flat))

	for _, flatIdx := range visible {
		fc := flat[flatIdx]
		cc := &cache[flatIdx]

		if !cc.valid {
			// Separator.
			sep := comment.Separator(fc.Depth, rc.config.CommentWidth, fc.Comment.ID, rc.firstCommentID)
			if sep != "" {
				indentedSep, _ := text.WrapWithPad(sep, rc.screenWidth, leftMargin)
				cc.sep = indentedSep
				cc.sepLines = strings.Count(indentedSep, "\n")
			}

			// Render the comment body.
			depthIndent := comment.IndentString(fc.Depth)
			depthIndentLen := len(depthIndent)
			availableWidth := contentWidth - depthIndentLen
			adjustedCommentWidth := rc.config.CommentWidth - fc.Depth

			rendered := comment.RenderBody(&fc.Comment, fc.Depth, rc.config, rc.originalPoster, fc.TopLevelAuthor,
				adjustedCommentWidth, availableWidth, rc.lastVisited)

			withDepth, _ := text.WrapWithPad(rendered, contentWidth, depthIndent)
			withMargin, _ := text.WrapWithPad(withDepth+"\n", rc.screenWidth, leftMargin)
			cc.body = withMargin
			cc.bodyLines = strings.Count(withMargin, "\n")

			// Pre-render fold indicator.
			if fc.DescendantCount > 0 {
				indicator := comment.FoldIndicator(fc.DescendantCount, fc.Depth, adjustedCommentWidth)
				indentedIndicator, _ := text.WrapWithPad(indicator, rc.screenWidth, leftMargin)
				cc.fold = indentedIndicator
				cc.foldLines = strings.Count(indentedIndicator, "\n")
			}

			cc.valid = true
		}

		sb.WriteString(cc.sep)
		lineCount += cc.sepLines

		startLine := lineCount

		sb.WriteString(cc.body)
		lineCount += cc.bodyLines

		// Fold indicator for collapsed comments with children.
		if fc.Collapsed && fc.DescendantCount > 0 {
			sb.WriteString(cc.fold)
			lineCount += cc.foldLines
		}

		metrics[flatIdx] = LineMetrics{
			StartLine: startLine,
			LineCount: lineCount - startLine,
		}
	}

	contentLines := lineCount

	// Add bottom padding so the last comments can be scrolled to the top.
	sb.WriteString(strings.Repeat("\n", rc.viewportHeight))

	return sb.String(), contentLines, metrics
}
