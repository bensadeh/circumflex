package pane

import (
	"strings"
	"unicode/utf8"

	"github.com/bensadeh/circumflex/nerdfonts"
	"github.com/bensadeh/circumflex/style"

	tea "charm.land/bubbletea/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

// This file is the single home of search-input UI: every search surface —
// the front page's Hacker News search and the comment section's and
// reader's in-page search — handles keys through TextPrompt and renders
// through PromptLabel and CommittedSearchLabel. Changing an atom here (the
// sigil, the cursor, the committed styling) changes every view; rendering a
// search prompt any other way is a bug.

// TextPrompt is a one-line text input: the Scroller's search prompt and the
// front page's Hacker News search share its key handling, so the two prompts
// can never drift apart in feel.
type TextPrompt struct {
	active bool
	input  string
}

func (p *TextPrompt) Start() {
	p.active = true
	p.input = ""
}

func (p *TextPrompt) Active() bool { return p.active }

// Text is the typed input; after PromptCommitted it holds the committed text.
func (p *TextPrompt) Text() string { return p.input }

// HandleKey feeds one key press to the open prompt. Printable characters
// append, enter commits the typed text, esc and backspacing past empty
// cancel. Committing an empty prompt also cancels.
func (p *TextPrompt) HandleKey(msg tea.KeyPressMsg) PromptResult {
	switch msg.Code {
	case tea.KeyEscape:
		p.active = false
		p.input = ""

		return PromptCanceled

	case tea.KeyEnter:
		p.active = false

		if p.input == "" {
			return PromptCanceled
		}

		return PromptCommitted

	case tea.KeyBackspace:
		if p.input == "" {
			p.active = false

			return PromptCanceled
		}

		_, size := utf8.DecodeLastRuneInString(p.input)
		p.input = p.input[:len(p.input)-size]

		return PromptPending
	}

	if msg.Text != "" && msg.Mod&^tea.ModShift == 0 {
		p.input += msg.Text
	}

	return PromptPending
}

// promptSigil is the / every search surface opens with.
func promptSigil() string { return style.Faint("/") }

// promptCursorMarker marks where the real terminal cursor belongs: the
// prompt renders it after the typed text, and ExtractPromptCursor trades it
// for the actual cursor at the frame level. Its glyph only shows if a view
// renders a prompt outside the main frame — a beam, so even that failure
// mode reads as a cursor.
const promptCursorMarker = "▏"

// promptCursor is the input cursor, appended after the typed text.
func promptCursor() string { return promptCursorMarker }

// ExtractPromptCursor locates the prompt cursor in a rendered frame, blanks
// its cell and reports the position, for the program to park the real
// terminal cursor there — the cursor the terminal draws in its own
// configured color and blink. Every open search prompt lives in a footer or
// status row, so only the frame's last line is scanned; ok is false when no
// prompt is open. The typed query itself may contain the marker rune — the
// cursor is appended after the input, so the last occurrence is the cursor.
func ExtractPromptCursor(frame string) (x, y int, cleaned string, ok bool) {
	lastNL := strings.LastIndexByte(frame, '\n')
	lastLine := frame[lastNL+1:]

	idx := strings.LastIndex(lastLine, promptCursorMarker)
	if idx < 0 {
		return 0, 0, frame, false
	}

	x = xansi.StringWidth(lastLine[:idx])
	y = strings.Count(frame, "\n")
	cleaned = frame[:lastNL+1] + lastLine[:idx] + " " + lastLine[idx+len(promptCursorMarker):]

	return x, y, cleaned, true
}

// promptZone is everything before the typed text: the faint /, or under
// nerd fonts the magnifier at full strength with extra room after the wide
// glyph.
func promptZone(enableNerdFonts bool) string {
	if enableNerdFonts {
		return nerdfonts.Search + "  "
	}

	return promptSigil()
}

// PromptSigilWidth is the cells the prompt zone occupies before the typed
// text. Views that keep the text column identical between the open prompt
// and the committed query outdent the prompt by the difference.
func PromptSigilWidth(enableNerdFonts bool) int {
	return xansi.StringWidth(promptZone(enableNerdFonts))
}

// PromptLabel renders an open prompt: the sigil zone, the typed text and
// the cursor.
func PromptLabel(input string, enableNerdFonts bool) string {
	return promptZone(enableNerdFonts) + input + promptCursor()
}

// CommittedSearchLabel renders a committed query dimmed behind its sigil.
// Nerd fonts trade the / for the done-searching glyph — icon overrides it,
// empty picks the shared default; views that keep the plain sigil even under
// nerd fonts (the front page) pass enableNerdFonts=false.
func CommittedSearchLabel(query, icon string, enableNerdFonts bool) string {
	prompt := promptSigil()

	if enableNerdFonts {
		if icon == "" {
			icon = nerdfonts.SearchCommitted
		}

		prompt = icon + "  "
	}

	return prompt + style.Faint(query)
}
