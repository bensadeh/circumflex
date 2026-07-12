package comments

import (
	"image/color"
	"strings"

	"github.com/bensadeh/circumflex/comment"
	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/style"
)

// renderContext holds the stable parameters for rendering a comment view.
// These values don't change between rebuilds unless the window is resized
// or the model is re-created.
type renderContext struct {
	header          string          // pre-computed meta block (raw, before margin wrapping)
	rootBlocks      []comment.Block // parsed self-text, re-rendered on resize
	originalPoster  string
	firstCommentID  int
	commentWidth    int
	indent          int
	enableNerdFonts bool
	paneWidth       int
	viewportHeight  int
	lastVisited     int64
	story           storyFields // scalar fields needed for header rebuild on resize
	newComments     int
}

// storyFields holds the thread metadata needed to rebuild the comment header
// on window resize, without retaining the full comment tree.
type storyFields struct {
	URL           string
	Domain        string
	Author        string
	TimeAgo       string
	ID            int
	CommentsCount int
	Points        int
}

// renderedComment holds the pre-rendered lines for a single flat comment.
// Built once by prerenderComments and reused by renderFromFlat on every
// collapse/expand, avoiding the expensive syntax/wrapping pipeline.
// Header and content are stored separately so that focus highlighting can
// swap in a pre-rendered focused header without re-running the expensive
// content pipeline. Rebuilt on window resize.
type renderedComment struct {
	sep              []string // separator (before the comment body)
	header           []string // header (author + label + time) with margins
	headerFocused    []string // same header but with author highlighted
	content          []string // comment content with indentation and margins
	repliesCollapsed []string // replies indicator when collapsed (nil if no descendants)
}

// splitLines converts a newline-terminated rendered block into its lines.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}

	lines := strings.Split(s, "\n")
	if last := len(lines) - 1; lines[last] == "" {
		lines = lines[:last]
	}

	return lines
}

// prerenderComments renders every comment in flat upfront, so that subsequent
// collapse/expand operations only concatenate pre-rendered strings.
func prerenderComments(rc renderContext, flat []flatComment) []renderedComment {
	leftMargin := strings.Repeat(" ", layout.CommentSectionLeftMargin)
	contentWidth := layout.CommentContentWidth(rc.paneWidth)
	commentWidth := min(contentWidth, rc.commentWidth)

	rendered := make([]renderedComment, len(flat))

	for i := range flat {
		fc := &flat[i]
		out := &rendered[i]

		sep := comment.Separator(fc.Depth, commentWidth, fc.Comment.ID, rc.firstCommentID)
		if sep != "" {
			out.sep = splitLines(style.PrefixLines(sep, leftMargin))
		}

		// Reserve 1 col inside commentWidth for the colored indent symbol when
		// the comment is nested (depth >= 1). Top-level comments have no symbol.
		symbolCols := 0
		if fc.Depth > 0 {
			symbolCols = 1
		}

		// Fold symbolCols into the floor so adjustedCommentWidth >= MinCommentWidth
		// even at the plateau.
		indentCols := comment.EffectiveIndentColumns(fc.Depth, rc.indent, commentWidth, layout.MinCommentWidth+symbolCols)
		depthIndent := strings.Repeat(" ", indentCols)

		availableWidth := contentWidth - indentCols
		adjustedCommentWidth := commentWidth - indentCols - symbolCols

		pad := leftMargin + depthIndent

		header := comment.Header(&fc.Comment, fc.Depth, rc.originalPoster, fc.TopLevelAuthor, rc.lastVisited, rc.enableNerdFonts, false)
		out.header = splitLines(style.PrefixLines(header, pad))

		focusedHeader := comment.Header(&fc.Comment, fc.Depth, rc.originalPoster, fc.TopLevelAuthor, rc.lastVisited, rc.enableNerdFonts, true)
		out.headerFocused = splitLines(style.PrefixLines(focusedHeader, pad))

		var fg color.Color
		if comment.IsMod(fc.Comment.Author) {
			fg = style.CommentModFg()
		}

		content := comment.RenderContent(fc.Blocks, fc.Depth, comment.RenderOptions{
			CommentWidth: adjustedCommentWidth,
			ScreenWidth:  availableWidth,
			NerdFonts:    rc.enableNerdFonts,
			Fg:           fg,
		})
		out.content = splitLines(style.PrefixLines(content+"\n", pad))

		// Pre-render replies indicator (only shown when collapsed). The indicator
		// sits at the first hidden reply's author column so that toggling
		// collapse/expand swaps content at the same left edge. Children are always
		// at depth >= 1, so the floor accounts for the child's 1-col symbol. The
		// trailing +1 matches the header's 1-col offset where the ▎ would sit.
		if fc.DescendantCount > 0 {
			childIndentCols := comment.EffectiveIndentColumns(fc.Depth+1, rc.indent, commentWidth, layout.MinCommentWidth+1)
			indicatorIndent := strings.Repeat(" ", childIndentCols+1)

			collapsed := comment.RepliesIndicator(fc.DescendantCount, indicatorIndent, true)
			out.repliesCollapsed = splitLines(style.PrefixLines(collapsed, leftMargin))
		}
	}

	return rendered
}

// renderFromFlat builds the full comment view content from the flat comment
// list, respecting fold state. It returns the content lines and line metrics
// indexed by flat index for navigation.
//
// focusedFlatIdx selects which comment gets the focused header variant;
// pass -1 when no comment is focused (scroll mode).
//
// The pre-rendered slice (indexed by flat index, built by prerenderComments)
// avoids re-running the expensive syntax-highlighting and text-wrapping
// pipeline on every collapse/expand or focus change.
func renderFromFlat(rc renderContext, flat []flatComment, visible []int, prerendered []renderedComment, focusedFlatIdx int) ([]string, []lineMetrics) {
	lines := splitLines(rc.header)
	lines = append(lines, "")

	metrics := make([]lineMetrics, len(flat))

	for _, flatIdx := range visible {
		fc := flat[flatIdx]
		pre := &prerendered[flatIdx]

		sepStart := len(lines)
		lines = append(lines, pre.sep...)

		startLine := len(lines)

		if flatIdx == focusedFlatIdx {
			lines = append(lines, pre.headerFocused...)
		} else {
			lines = append(lines, pre.header...)
		}

		lines = append(lines, pre.content...)

		if fc.DescendantCount > 0 && fc.Collapsed {
			lines = append(lines, pre.repliesCollapsed...)
		}

		metrics[flatIdx] = lineMetrics{
			SepStart:  sepStart,
			StartLine: startLine,
			LineCount: len(lines) - startLine,
		}
	}

	return lines, metrics
}
