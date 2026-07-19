package view

import (
	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/view/message"
	"github.com/bensadeh/circumflex/view/pane"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// errorView is the detail view a failed story load leaves in the pane. It
// counts as an open view like the comment section and reader: the list keeps
// its reading marker on the story that failed, J/K move on to its neighbors,
// and quit returns to the front page. There is nothing to scroll, so every
// other key is ignored.
type errorView struct {
	message   string
	title     string
	nerdFonts bool
	keymap    pane.CommonKeyMap
	metaBlock func(paneWidth int) string // the loading pane's placeholder, kept so the box doesn't flash away
	width     int
	height    int
}

func newErrorView(msg, title string, nerdFonts bool, metaBlock func(paneWidth int) string, width, height int) *errorView {
	return &errorView{
		message:   msg,
		title:     title,
		nerdFonts: nerdFonts,
		keymap:    pane.DefaultCommonKeyMap(),
		metaBlock: metaBlock,
		width:     width,
		height:    height,
	}
}

func (v *errorView) Init() tea.Cmd { return nil }

func (v *errorView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, v.keymap.Quit):
			return func() tea.Msg { return message.ErrorViewQuit{} }

		case key.Matches(msg, v.keymap.NextStory):
			return message.OpenAdjacentStoryCmd(1)

		case key.Matches(msg, v.keymap.PrevStory):
			return message.OpenAdjacentStoryCmd(-1)

		case key.Matches(msg, v.keymap.ToggleWide):
			return message.ToggleWideLayoutCmd()
		}
	}

	return nil
}

// View keeps the failed story's title in the header, unbolded like the
// loading pane's — the full weight belongs to stories that actually opened —
// and the meta block placeholder in its spot, so the transition from loading
// to error moves nothing.
func (v *errorView) View() string {
	// Width alone only breaks at spaces; an unbroken token wider than the
	// pane (a URL, a hostname) needs the hard wrap.
	wrapWidth := max(1, v.width-2*layout.HeaderLeftMargin)
	wrapped := lipgloss.NewStyle().
		Width(wrapWidth).
		Align(lipgloss.Center).
		Render(lipgloss.Wrap(v.message, wrapWidth, ""))

	body := placeholderBody(v.metaBlock(v.width), wrapped, v.width, max(0, v.height-layout.PaneChromeHeight))

	return pane.LoadingTitleHeader(v.title, v.nerdFonts, layout.HeaderLeftMargin, v.width) +
		"\n" + body + "\n" + pane.FooterSeparator(v.width) + "\n"
}
