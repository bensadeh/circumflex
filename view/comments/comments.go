package comments

import (
	"fmt"
	"strings"

	"github.com/bensadeh/circumflex/comment"
	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/help"
	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/meta"
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
		Content:       thread.Content,
	}

	hdr := buildCommentHeader(sf, enableNerdFonts, newComments, clampedWidth) + "\n"

	rc := renderContext{
		header:          hdr,
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
		m.rc.header = buildCommentHeader(m.rc.story, m.rc.enableNerdFonts, m.rc.newComments, cw) + "\n"

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
		content := help.FitToHeight(
			help.CommentHelpScreen(m.rc.paneWidth, m.rc.enableNerdFonts, m.keymap.NextStory.Enabled()),
			m.rc.viewportHeight,
		)

		return header.HelpHeader("Comment Section", m.rc.paneWidth) + "\n" +
			content + "\n" +
			pane.FooterSeparator(m.rc.paneWidth) + "\n" +
			help.Footer(m.rc.paneWidth)
	}

	content := scrollbar.Attach(m.Viewport.View(), m.rc.paneWidth, m.ContentLines, m.rc.viewportHeight, m.Viewport.YOffset())

	return m.titleHeader + "\n" + content + "\n" + pane.FooterSeparator(m.rc.paneWidth) + "\n" + m.modeIndicator()
}

func (m *Model) rebuildTitleHeader() {
	m.titleHeader = pane.TitleHeader(m.title, m.rc.enableNerdFonts, layout.CommentSectionLeftMargin, m.rc.paneWidth)
}

// updateViewport re-renders the viewport content with the current focus state.
// This is cheap: it concatenates pre-rendered strings, picking the focused
// header variant for the focused comment.
func (m *Model) updateViewport() {
	focusedFlatIdx := -1
	if m.mode == modeNavigate && m.focusedIdx >= 0 && m.focusedIdx < len(m.visible) {
		focusedFlatIdx = m.visible[m.focusedIdx]
	}

	content, contentLines, metrics := renderFromFlat(m.rc, m.flat, m.visible, m.prerendered, focusedFlatIdx)
	m.ContentLines = contentLines
	m.lineMetrics = metrics
	m.Viewport.SetContent(content)
}

func (m *Model) openStoryInBrowser() tea.Cmd {
	return pane.OpenStoryInBrowser(m.rc.story.URL, m.rc.story.ID)
}

func (m *Model) openCommentsInBrowser() tea.Cmd {
	return pane.OpenCommentsInBrowser(m.rc.story.ID)
}

func (m *Model) modeIndicator() string {
	var label string

	switch m.mode {
	case modeRead:
		label = "  ☰ " + style.Faint("read")
	case modeNavigate:
		label = "  … " + style.Faint(" nav")
	}

	diSlot := 0
	if m.maxDepth > 0 {
		diSlot = 1 + 1 + len(fmt.Sprintf("%d", m.maxDepth)) // " ⋮" + digits
	}

	commentWidth := layout.CommentColumnWidth(m.rc.paneWidth, m.rc.commentWidth)
	totalWidth := layout.CommentSectionLeftMargin + commentWidth
	padding := max(1, totalWidth-lipgloss.Width(label)-diSlot)

	result := label + strings.Repeat(" ", padding)

	if diSlot > 0 {
		di := ""
		if m.mode == modeRead {
			di = m.depthIndicator()
		}

		if di != "" {
			result += di + strings.Repeat(" ", max(0, diSlot-lipgloss.Width(di)))
		} else {
			result += strings.Repeat(" ", diSlot)
		}
	}

	return xansi.Truncate(result, m.rc.paneWidth, "")
}

func (m *Model) depthIndicator() string {
	level := m.expandedDepth
	numStr := fmt.Sprintf("%d", level)

	cycle := style.IndentCycle()

	if level == 0 {
		return ""
	}

	if len(cycle) == 0 {
		return "\u22ee" + style.Faint(numStr)
	}

	colorFn := cycle[(level-1)%len(cycle)]

	return "\u22ee" + colorFn(numStr)
}

func (m *Model) rebuildContent() {
	m.visible = computeVisible(m.flat)
	m.updateViewport()
}

func buildCommentHeader(s storyFields, enableNerdFonts bool, newComments int, width int) string {
	return meta.CommentSection(meta.Data{
		URL:           s.URL,
		Domain:        s.Domain,
		Author:        s.Author,
		TimeAgo:       s.TimeAgo,
		ID:            s.ID,
		Points:        s.Points,
		CommentsCount: s.CommentsCount,
		NewComments:   newComments,
		RootComment:   renderRootComment(s.Content, meta.ContentWidth(width), enableNerdFonts),
		NerdFonts:     enableNerdFonts,
	}).Render(width)
}

func renderRootComment(c string, contentWidth int, enableNerdFonts bool) string {
	if c == "" {
		return ""
	}

	rendered := comment.Render(c, contentWidth, contentWidth, enableNerdFonts, nil)

	return lipgloss.Wrap(rendered, contentWidth, "")
}
