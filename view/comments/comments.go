package comments

import (
	"fmt"
	"strings"

	"github.com/bensadeh/circumflex/comment"
	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/help"
	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/meta"
	"github.com/bensadeh/circumflex/nerdfonts"
	"github.com/bensadeh/circumflex/scrollbar"
	"github.com/bensadeh/circumflex/style"
	"github.com/bensadeh/circumflex/timeago"
	"github.com/bensadeh/circumflex/view/pane"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

type Model struct {
	pane.Scroller

	keymap keyMap
	mode   mode

	flat          []flatComment
	visible       []int // indices into flat
	focusedIdx    int   // index into visible (-1 = no focus, scroll mode)
	expandedDepth int
	maxDepth      int
	rc            renderContext
	title         string
	titleHeader   string
	showHelp      bool

	prerendered []renderedComment

	lineMetrics []lineMetrics // indexed by flat index
}

const scrollPadding = 2 // breathing room above/below when scrolling to a comment

func New(thread *comment.Thread, lastVisited int64, commentWidth, indent int, enableNerdFonts bool, width, height int) *Model {
	km := defaultKeyMap()

	// Viewport handles j/k in scroll mode (toggled off in navigate mode).
	// h/l are always handled by us (collapse/expand), so disable them on viewport.
	vp := pane.NewViewport(width, height-layout.PaneChromeHeight)
	vp.KeyMap.Left.SetEnabled(false)
	vp.KeyMap.Right.SetEnabled(false)

	flat := flatten(thread)

	newComments := comment.NewCommentsCount(thread, lastVisited)
	clampedWidth := layout.CommentColumnWidth(width, commentWidth)

	sf := storyFields{
		URL:           thread.URL,
		Domain:        thread.Domain,
		Author:        thread.Author,
		TimeAgo:       timeago.RelativeTime(thread.Time),
		ID:            thread.ID,
		CommentsCount: thread.CommentsCount,
		Points:        thread.Points,
	}

	rootBlocks := comment.Parse(thread.Content)
	hdr := buildCommentHeader(sf, rootBlocks, enableNerdFonts, clampedWidth) + "\n"

	rc := renderContext{
		header:          hdr,
		rootBlocks:      rootBlocks,
		originalPoster:  thread.Author,
		firstCommentID:  comment.FirstCommentID(thread.Comments),
		commentWidth:    commentWidth,
		indent:          indent,
		enableNerdFonts: enableNerdFonts,
		paneWidth:       width,
		viewportHeight:  max(0, height-layout.PaneChromeHeight),
		lastVisited:     lastVisited,
		story:           sf,
		newComments:     newComments,
	}

	md := 0
	for _, fc := range flat {
		if fc.Depth > md {
			md = fc.Depth
		}
	}

	m := Model{
		Scroller:      pane.Scroller{Viewport: vp},
		keymap:        km,
		mode:          modeRead,
		flat:          flat,
		focusedIdx:    -1,
		expandedDepth: 0, // initial: only top-level visible
		maxDepth:      md,
		title:         thread.Title,
		prerendered:   prerenderComments(rc, flat),
		rc:            rc,
	}

	m.rebuildTitleHeader()
	m.rebuildContent()

	return &m
}

// DisableStoryNavigation removes the J/K adjacent-story bindings, for
// standalone use where there is no story list to move through.
func (m *Model) DisableStoryNavigation() {
	m.keymap.DisableStoryNavigation()
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	case tea.MouseWheelMsg:
		if m.showHelp {
			return nil
		}

		m.HandleMouseWheel(msg)

		return nil

	case tea.WindowSizeMsg:
		anchorIdx := m.anchorComment()
		screenPos := m.screenPosition(anchorIdx)

		m.rc.paneWidth = msg.Width
		m.rc.viewportHeight = max(0, msg.Height-layout.PaneChromeHeight)
		m.Viewport.SetWidth(msg.Width)
		m.Viewport.SetHeight(m.rc.viewportHeight)

		cw := layout.CommentColumnWidth(msg.Width, m.rc.commentWidth)
		m.rc.header = buildCommentHeader(m.rc.story, m.rc.rootBlocks, m.rc.enableNerdFonts, cw) + "\n"

		m.rebuildTitleHeader()
		m.prerendered = prerenderComments(m.rc, m.flat)
		m.rebuildContent()
		m.restoreScreenPosition(anchorIdx, screenPos)

		return nil
	}

	return nil
}

func (m *Model) View() string {
	if m.showHelp {
		contentWidth := layout.CommentColumnWidth(m.rc.paneWidth, m.rc.commentWidth)
		content := help.FitToHeight(
			help.CommentHelpScreen(layout.CommentSectionLeftMargin, contentWidth, m.rc.enableNerdFonts, m.keymap.NextStory.Enabled()),
			m.rc.viewportHeight,
		)

		return header.HelpHeader("Comment Section", m.rc.paneWidth) + "\n" +
			content + "\n" +
			pane.FooterSeparator(m.rc.paneWidth) + "\n" +
			help.Footer(layout.CommentSectionLeftMargin, contentWidth, m.rc.enableNerdFonts)
	}

	content := scrollbar.Attach(m.Viewport.View(), m.rc.paneWidth, m.ContentLines, m.rc.viewportHeight, m.Viewport.YOffset())

	return m.titleHeader + "\n" + content + "\n" + pane.FooterSeparator(m.rc.paneWidth) + "\n" + m.modeIndicator()
}

func (m *Model) rebuildTitleHeader() {
	m.titleHeader = pane.TitleHeader(m.title, m.rc.enableNerdFonts, layout.CommentSectionLeftMargin, m.rc.paneWidth)
}

// updateViewport re-renders the viewport content with the current focus state.
// This is cheap: it concatenates pre-rendered lines, picking the focused
// header variant for the focused comment.
func (m *Model) updateViewport() {
	focusedFlatIdx := -1
	if m.mode == modeNavigate && m.focusedIdx >= 0 && m.focusedIdx < len(m.visible) {
		focusedFlatIdx = m.visible[m.focusedIdx]
	}

	lines, metrics := renderFromFlat(m.rc, m.flat, m.visible, m.prerendered, focusedFlatIdx)
	m.lineMetrics = metrics
	m.SetLines(lines)
}

func (m *Model) openStoryInBrowser() tea.Cmd {
	return pane.OpenStoryInBrowser(m.rc.story.URL, m.rc.story.ID)
}

func (m *Model) openCommentsInBrowser() tea.Cmd {
	return pane.OpenCommentsInBrowser(m.rc.story.ID)
}

func (m *Model) modeIndicator() string {
	var icon, text string

	switch m.mode {
	case modeRead:
		icon, text = "☰", "read"
		if m.rc.enableNerdFonts {
			icon = nerdfonts.Document
		}
	case modeNavigate:
		text = " nav"

		// Tree-view convention: + on a collapsed comment (expandable),
		// − on an expanded one (collapsible), … / draft outline on a leaf.
		icon = "…"
		nfIcon := nerdfonts.CommentDraft

		if fc := m.focusedComment(); fc != nil && fc.DescendantCount > 0 {
			if fc.Collapsed {
				icon, nfIcon = "+", nerdfonts.CommentPlusOutline
			} else {
				icon, nfIcon = "−", nerdfonts.CommentMinusOutline
			}
		}

		if m.rc.enableNerdFonts {
			icon = nfIcon
		}
	}

	// Nerd font glyphs render wider than one cell, so they get extra room.
	sep := " "
	if m.rc.enableNerdFonts {
		sep = "  "
	}

	label := "  " + icon + sep + style.Faint(text)

	di := ""
	if m.mode == modeRead {
		di = m.depthIndicator()
	}

	// Three sections across the comment column: the mode label at the left
	// margin, the depth indicator between, and the counts ending at the
	// column's right edge — the same edge the meta block and the separator
	// rule share.
	commentWidth := layout.CommentColumnWidth(m.rc.paneWidth, m.rc.commentWidth)
	totalWidth := layout.CommentSectionLeftMargin + commentWidth

	result := pane.FooterSections(totalWidth,
		label,
		di,
		commentCountLabel(m.rc.story.CommentsCount, m.rc.newComments, m.rc.enableNerdFonts))

	return xansi.Truncate(result, m.rc.paneWidth, "")
}

// commentCountLabel is the footer's comment tally: total comments and, in
// parentheses, how many arrived since the last visit. The icon keeps full
// strength like the footer's other icons; the counts stay faint — they
// inform, they don't call for attention. In nerd-fonts mode the new-comment
// count also takes the meta new-comments color — a hue shift within the same
// faint register, hinting at fresh activity without shouting.
func commentCountLabel(commentsCount, newComments int, enableNerdFonts bool) string {
	if enableNerdFonts {
		label := style.Faint(fmt.Sprintf("%d", commentsCount))
		if newComments > 0 {
			label += style.Faint(" (") + style.MetaNewCommentsFaint(fmt.Sprintf("%d", newComments)) + style.Faint(")")
		}

		return label + " " + nerdfonts.Comment
	}

	label := fmt.Sprintf("%d comments", commentsCount)
	if newComments > 0 {
		label += fmt.Sprintf(" (%d new)", newComments)
	}

	return style.Faint(label)
}

func (m *Model) depthIndicator() string {
	level := m.expandedDepth
	if level == 0 {
		return ""
	}

	icon := "⋮"
	if level == m.maxDepth {
		icon = "∴"
	}

	numStr := fmt.Sprintf("%d", level)

	cycle := style.IndentCycleFaint()
	if len(cycle) == 0 {
		return icon + " " + style.Faint(numStr)
	}

	colorFn := cycle[(level-1)%len(cycle)]

	return icon + " " + colorFn(numStr)
}

func (m *Model) rebuildContent() {
	m.visible = computeVisible(m.flat)
	m.updateViewport()
}

func buildCommentHeader(s storyFields, rootBlocks []comment.Block, enableNerdFonts bool, width int) string {
	block := meta.CommentSection(meta.Data{
		URL:         s.URL,
		Domain:      s.Domain,
		Author:      s.Author,
		TimeAgo:     s.TimeAgo,
		Points:      s.Points,
		RootComment: renderRootComment(rootBlocks, meta.ContentWidth(width), enableNerdFonts),
		NerdFonts:   enableNerdFonts,
	}).Render(width)

	return style.PrefixLines(block, strings.Repeat(" ", layout.CommentSectionLeftMargin))
}

// renderRootComment renders the story's self-text for the meta block. A
// story without self-text renders empty, which the meta block treats as
// absent.
func renderRootComment(blocks []comment.Block, contentWidth int, enableNerdFonts bool) string {
	rendered := comment.RenderBlocks(blocks, comment.RenderOptions{
		CommentWidth: contentWidth,
		ScreenWidth:  contentWidth,
		NerdFonts:    enableNerdFonts,
	})

	return lipgloss.Wrap(rendered, contentWidth, "")
}
