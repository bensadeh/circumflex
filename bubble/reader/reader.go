package reader

import (
	"clx/bubble/list/message"
	"clx/constants"
	"clx/style"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Model is the Bubble Tea model for the built-in reader view.
type Model struct {
	viewport viewport.Model
	keymap   KeyMap

	headerLines    []int  // line indices containing ■ (section headers)
	title          string // article title for the fixed header
	contentLines   int    // actual content lines (excluding bottom padding)
	screenWidth    int
	viewportHeight int
	standalone     bool // when true, quit sends tea.Quit instead of ReaderViewQuitMsg
}

const (
	headerHeight = 2 // title + overline separator
	footerHeight = 2 // underline separator + keybinding hints
)

// New creates a new reader view model.
func New(content, title string, width, height int) *Model {
	vpHeight := height - headerHeight - footerHeight

	vp := viewport.New(
		viewport.WithWidth(width),
		viewport.WithHeight(vpHeight),
	)

	vp.KeyMap = viewport.DefaultKeyMap()
	vp.KeyMap.HalfPageDown.SetEnabled(false)
	vp.KeyMap.HalfPageUp.SetEnabled(false)
	vp.KeyMap.PageDown.SetEnabled(false)
	vp.KeyMap.PageUp.SetEnabled(false)
	vp.MouseWheelEnabled = false

	lines := strings.Split(content, "\n")
	contentLineCount := len(lines)

	if contentLineCount > 0 && lines[contentLineCount-1] == "" {
		contentLineCount--
	}

	// Scan for header lines (lines containing the ■ block character).
	var headers []int

	for i, line := range lines {
		if strings.Contains(line, constants.Block) {
			headers = append(headers, i)
		}
	}

	// Add bottom padding so G scrolls the last content line to the bottom.
	padding := strings.Repeat("\n", vpHeight)
	padded := strings.TrimRight(content, "\n") + padding

	vp.SetContent(padded)

	m := &Model{
		viewport:       vp,
		keymap:         defaultKeyMap(),
		headerLines:    headers,
		title:          title,
		contentLines:   contentLineCount,
		screenWidth:    width,
		viewportHeight: vpHeight,
	}

	return m
}

// Init returns nil; no initial commands needed.
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the reader view.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, m.keymap.Quit) {
			if m.standalone {
				return tea.Quit
			}

			return func() tea.Msg { return message.ReaderViewQuitMsg{} }
		}

		if key.Matches(msg, m.keymap.GotoTop) {
			m.viewport.GotoTop()

			return nil
		}

		if key.Matches(msg, m.keymap.GotoBottom) {
			m.gotoBottom()

			return nil
		}

		if key.Matches(msg, m.keymap.HalfPageDown) {
			m.halfPageDown()

			return nil
		}

		if key.Matches(msg, m.keymap.HalfPageUp) {
			m.halfPageUp()

			return nil
		}

		if key.Matches(msg, m.keymap.PageDown) {
			m.pageDown()

			return nil
		}

		if key.Matches(msg, m.keymap.PageUp) {
			m.pageUp()

			return nil
		}

		if key.Matches(msg, m.keymap.NextHeader) {
			m.jumpToHeader(1)

			return nil
		}

		if key.Matches(msg, m.keymap.PrevHeader) {
			m.jumpToHeader(-1)

			return nil
		}

	case tea.MouseWheelMsg:
		delta := m.viewport.MouseWheelDelta
		maxOffset := max(0, m.contentLines-m.viewportHeight)

		switch msg.Button {
		case tea.MouseWheelDown:
			m.viewport.SetYOffset(min(m.viewport.YOffset()+delta, maxOffset))
		case tea.MouseWheelUp:
			m.viewport.SetYOffset(max(0, m.viewport.YOffset()-delta))
		}

		return nil

	case tea.WindowSizeMsg:
		m.screenWidth = msg.Width
		m.viewportHeight = msg.Height - headerHeight - footerHeight
		m.viewport.SetWidth(msg.Width)
		m.viewport.SetHeight(m.viewportHeight)

		return nil
	}

	before := m.viewport.YOffset()

	var cmd tea.Cmd

	m.viewport, cmd = m.viewport.Update(msg)
	m.clampScroll(before)

	return cmd
}

// View renders the reader view.
func (m *Model) View() string {
	return m.headerView() + "\n" + m.viewport.View() + "\n" + m.footerSeparator() + "\n" + m.modeIndicator()
}

func (m *Model) headerView() string {
	leftMargin := strings.Repeat(" ", constants.ReaderViewLeftMargin)
	title := leftMargin + m.title
	separator := strings.Repeat("‾", m.screenWidth)

	return title + "\n" + separator
}

func (m *Model) footerSeparator() string {
	underscore := lipgloss.NewStyle().Underline(true).Render(" ")

	return strings.Repeat(underscore, m.screenWidth)
}

func (m *Model) modeIndicator() string {
	return style.ModeIndicator(style.Logo("{", "≡", "}"), []style.Binding{
		{Key: "n/N", Desc: "next/prev section"},
	})
}

// clampScroll prevents scrolling down past the last content line while still
// allowing upward scrolling from a position beyond the clamp point (e.g. after
// an n/N header jump).
func (m *Model) clampScroll(before int) {
	maxOffset := max(0, m.contentLines-m.viewportHeight)
	after := m.viewport.YOffset()

	if after > before && after > maxOffset {
		m.viewport.SetYOffset(max(before, maxOffset))
	}
}

func (m *Model) halfPageDown() {
	halfPage := m.viewportHeight / 2
	maxOffset := max(0, m.contentLines-m.viewportHeight)
	m.viewport.SetYOffset(min(m.viewport.YOffset()+halfPage, maxOffset))
}

func (m *Model) halfPageUp() {
	halfPage := m.viewportHeight / 2
	m.viewport.SetYOffset(max(0, m.viewport.YOffset()-halfPage))
}

func (m *Model) pageDown() {
	maxOffset := max(0, m.contentLines-m.viewportHeight)
	m.viewport.SetYOffset(min(m.viewport.YOffset()+m.viewportHeight, maxOffset))
}

func (m *Model) pageUp() {
	m.viewport.SetYOffset(max(0, m.viewport.YOffset()-m.viewportHeight))
}

func (m *Model) gotoBottom() {
	m.viewport.SetYOffset(max(0, m.contentLines-m.viewportHeight))
}

func (m *Model) jumpToHeader(direction int) {
	if len(m.headerLines) == 0 {
		return
	}

	yOffset := m.viewport.YOffset()

	if direction > 0 {
		for _, line := range m.headerLines {
			if line > yOffset+1 {
				m.viewport.SetYOffset(line)

				return
			}
		}
	} else {
		for i := len(m.headerLines) - 1; i >= 0; i-- {
			if m.headerLines[i] < yOffset {
				m.viewport.SetYOffset(m.headerLines[i])

				return
			}
		}

		if yOffset > 0 {
			m.viewport.SetYOffset(0)
		}
	}
}

// standaloneModel wraps Model to implement tea.Model for standalone use.
type standaloneModel struct {
	inner *Model
}

func (s standaloneModel) Init() tea.Cmd {
	return s.inner.Init()
}

func (s standaloneModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmd := s.inner.Update(msg)

	return s, cmd
}

func (s standaloneModel) View() tea.View {
	v := tea.NewView(s.inner.View())
	v.AltScreen = true

	return v
}

// Run launches the reader as a standalone Bubble Tea program.
func Run(content, title string) error {
	m := New(content, title, 0, 0)
	m.standalone = true

	p := tea.NewProgram(standaloneModel{inner: m})
	_, err := p.Run()

	return err
}
