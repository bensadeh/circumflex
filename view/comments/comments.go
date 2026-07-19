package comments

import (
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
	"github.com/bensadeh/circumflex/view/message"
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
	linkTrail     []message.TrailEntry

	prerendered []renderedComment

	lineMetrics []lineMetrics // indexed by flat index

	searchMatches []commentMatch // all matches in document order, hidden ones included
	searchCurrent int
}

const scrollPadding = 2 // breathing room above/below when scrolling to a comment

func New(thread *comment.Thread, lastVisited int64, commentWidth, indent int, enableNerdFonts bool, width, height int) *Model {
	km := defaultKeyMap()

	// Viewport handles j/k in scroll mode (toggled off in navigate mode).
	vp := pane.NewViewport(width, height-layout.PaneChromeHeight)

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
	hdr := buildCommentHeader(sf, rootBlocks, newComments, enableNerdFonts, clampedWidth) + "\n"

	rc := renderContext{
		header:          hdr,
		rootBlocks:      rootBlocks,
		originalPoster:  thread.Author,
		firstCommentID:  comment.FirstCommentID(thread.Comments),
		commentWidth:    commentWidth,
		indent:          indent,
		enableNerdFonts: enableNerdFonts,
		paneWidth:       width,
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
		Scroller:      pane.Scroller{Viewport: vp, SearchCommittedIcon: nerdfonts.CommentSearchCommitted},
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

// DisableAppKeys removes the bindings that need the surrounding app, for
// standalone use where there is no story list and no split layout.
func (m *Model) DisableAppKeys() {
	m.keymap.DisableAppKeys()
}

// SetLinkTrail marks this thread as reached by following a link inside an
// article: trail is the chain of pages behind it, so quit steps back through
// them instead of closing the detail pane, and the title row carries the
// depth badge.
func (m *Model) SetLinkTrail(trail []message.TrailEntry) {
	if len(trail) == 0 {
		return
	}

	m.linkTrail = trail
	m.rebuildTitleHeader()
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

		widthChanged := msg.Width != m.rc.paneWidth

		m.rc.paneWidth = msg.Width
		m.Viewport.SetWidth(msg.Width)
		m.Viewport.SetHeight(max(0, msg.Height-layout.PaneChromeHeight))

		// A height-only resize changes no wrapping: the header, prerendered
		// comments, and match positions all stand. Only the bottom padding
		// tracks the viewport height.
		if !widthChanged {
			m.RefreshPadding()
			m.restoreScreenPosition(anchorIdx, screenPos)

			return nil
		}

		cw := layout.CommentColumnWidth(msg.Width, m.rc.commentWidth)
		m.rc.header = buildCommentHeader(m.rc.story, m.rc.rootBlocks, m.rc.newComments, m.rc.enableNerdFonts, cw) + "\n"

		m.rebuildTitleHeader()
		m.prerendered = prerenderComments(m.rc, m.flat)

		// The rewrap moved text within comments, so comment-relative match
		// positions are recomputed before the rebuild resolves them.
		if query := m.ActiveQuery(); query != "" {
			m.searchMatches = m.findAllMatches(query)
			m.searchCurrent = min(m.searchCurrent, max(0, len(m.searchMatches)-1))
		}

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
			m.Viewport.Height(),
		)

		return header.HelpHeader("Comment Section", m.rc.paneWidth) + "\n" +
			content + "\n" +
			pane.FooterSeparator(m.rc.paneWidth) + "\n" +
			help.Footer(layout.CommentSectionLeftMargin, contentWidth, m.rc.enableNerdFonts)
	}

	content := scrollbar.Attach(m.DecorateView(m.Viewport.View()), m.rc.paneWidth, m.ContentLines, m.Viewport.Height(), m.Viewport.YOffset())

	return m.titleHeader + "\n" + content + "\n" + pane.FooterSeparator(m.rc.paneWidth) + "\n" + m.modeIndicator()
}

func (m *Model) rebuildTitleHeader() {
	if len(m.linkTrail) > 0 {
		rightEdge := layout.CommentSectionLeftMargin + layout.CommentColumnWidth(m.rc.paneWidth, m.rc.commentWidth)
		badge := pane.DepthBadge(len(m.linkTrail))

		m.titleHeader = pane.TitleHeaderWithBadge(m.title, badge, m.rc.enableNerdFonts, layout.CommentSectionLeftMargin, rightEdge, m.rc.paneWidth)

		return
	}

	m.titleHeader = pane.TitleHeader(m.title, m.rc.enableNerdFonts, layout.CommentSectionLeftMargin, m.rc.paneWidth)
}

// updateViewport rebuilds the viewport content from the current fold state:
// a concatenation of pre-rendered lines. Called on structural changes only
// (collapse, expand, reveal, resize) — focus moves and search updates go
// through syncDecorations instead, which costs nothing per document line.
func (m *Model) updateViewport() {
	lines, metrics := renderFromFlat(m.rc, m.flat, m.visible, m.prerendered)
	m.lineMetrics = metrics
	m.SetLines(lines)
	m.syncDecorations()
}

// syncDecorations refreshes the display-time decorations: the focused header
// override and, when a search is active or being typed, the match positions
// re-resolved against the current line metrics.
func (m *Model) syncDecorations() {
	m.SetRowOverrides(m.focusOverrides())

	if m.SearchActive() || m.SearchPrompting() {
		matches, current := m.absoluteMatches()
		m.SetSearchMatches(matches)
		m.SetCurrentMatch(current)
	} else {
		m.SetSearchMatches(nil)
	}
}

// focusOverrides swaps the focused comment's header rows for their focused
// variant. Both variants render the same plain text, so row widths and
// match cell offsets are unaffected.
func (m *Model) focusOverrides() []pane.RowOverride {
	if m.focusedComment() == nil {
		return nil
	}

	flatIdx := m.visible[m.focusedIdx]
	lm := m.lineMetrics[flatIdx]
	focused := m.prerendered[flatIdx].headerFocused

	overrides := make([]pane.RowOverride, len(focused))
	for i, row := range focused {
		overrides[i] = pane.RowOverride{Line: lm.StartLine + i, Content: row}
	}

	return overrides
}

func (m *Model) openStoryInBrowser() tea.Cmd {
	return pane.OpenStoryInBrowser(m.rc.story.URL, m.rc.story.ID)
}

func (m *Model) openCommentsInBrowser() tea.Cmd {
	return pane.OpenCommentsInBrowser(m.rc.story.ID)
}

func (m *Model) modeIndicator() string {
	if search := m.SearchFooterLabel(m.rc.enableNerdFonts); search != "" {
		// The counter takes over the depth gauge's slot on the right:
		// position over the full match list once committed, the live total
		// while the prompt is open.
		counter := pane.MatchCountLabel(m.searchCurrent, len(m.searchMatches))

		if m.SearchPrompting() {
			counter = ""
			if m.ActiveQuery() != "" {
				counter = pane.MatchTotalLabel(len(m.searchMatches))
			}
		}

		commentWidth := layout.CommentColumnWidth(m.rc.paneWidth, m.rc.commentWidth)
		totalWidth := layout.CommentSectionLeftMargin + commentWidth
		result := layout.FooterSections(totalWidth, "  "+search, counter)

		return xansi.Truncate(result, m.rc.paneWidth, "")
	}

	var icon, text string

	switch m.mode {
	case modeRead:
		icon, text = "☰", "Comment Section"
		if m.rc.enableNerdFonts {
			icon = nerdfonts.CommentSection
		}
	case modeNavigate:
		text = "Navigate"

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

	// Pad the gap after the icon so the mode text starts at a fixed
	// column when toggling modes: ☰ measures two cells, the navigate
	// icons and nerd glyphs one (nerd glyphs render wider, spilling
	// into the gap).
	sep := strings.Repeat(" ", 3-xansi.StringWidth(icon))

	label := "  " + icon + sep + style.Faint(text)

	di := ""
	if m.mode == modeRead {
		di = m.depthIndicator()
	}

	// Two sections across the comment column: the mode label at the left
	// margin and the depth gauge ending at the column's right edge — the
	// same edge the meta block and the separator rule share. The comment
	// counts live in the meta block's opening rule, not here.
	commentWidth := layout.CommentColumnWidth(m.rc.paneWidth, m.rc.commentWidth)
	totalWidth := layout.CommentSectionLeftMargin + commentWidth

	result := layout.FooterSections(totalWidth, label, di)

	return xansi.Truncate(result, m.rc.paneWidth, "")
}

// depthIndicator is the footer's expansion gauge: one dot per indent level,
// filled up to the current expansion depth, a dim middle dot beyond it. Each
// filled dot takes its level's indent-cycle color, so the gauge doubles as a
// legend for the gutter markers. At zero expansion the all-dim gauge still
// shows how deep the thread goes.
func (m *Model) depthIndicator() string {
	cycle := style.IndentCycleFaint()

	var b strings.Builder

	for level := 1; level <= m.maxDepth; level++ {
		switch {
		case level > m.expandedDepth:
			b.WriteString(style.Faint("·"))
		case len(cycle) > 0:
			b.WriteString(cycle[(level-1)%len(cycle)]("•"))
		default:
			b.WriteString(style.Faint("•"))
		}
	}

	return b.String()
}

func (m *Model) rebuildContent() {
	m.visible = computeVisible(m.flat)
	m.updateViewport()
}

func buildCommentHeader(s storyFields, rootBlocks []comment.Block, newComments int, enableNerdFonts bool, width int) string {
	block := meta.CommentSection(meta.Data{
		URL:           s.URL,
		Domain:        s.Domain,
		Author:        s.Author,
		TimeAgo:       s.TimeAgo,
		Points:        s.Points,
		CommentsCount: s.CommentsCount,
		NewComments:   newComments,
		RootComment:   renderRootComment(rootBlocks, meta.ContentWidth(width), enableNerdFonts),
		NerdFonts:     enableNerdFonts,
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
