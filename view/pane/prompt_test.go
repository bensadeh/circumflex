package pane

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

func typeText(p *TextPrompt, text string) {
	for _, r := range text {
		p.HandleKey(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
}

func TestTextPrompt_TypeAndCommit(t *testing.T) {
	var p TextPrompt

	p.Start()
	assert.True(t, p.Active())

	typeText(&p, "gpu")
	assert.Equal(t, "gpu", p.Text())

	result := p.HandleKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.Equal(t, PromptCommitted, result)
	assert.False(t, p.Active())
	assert.Equal(t, "gpu", p.Text(), "committed text stays readable")
}

func TestTextPrompt_CommitEmptyCancels(t *testing.T) {
	var p TextPrompt

	p.Start()

	result := p.HandleKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.Equal(t, PromptCanceled, result)
	assert.False(t, p.Active())
}

func TestTextPrompt_EscapeCancels(t *testing.T) {
	var p TextPrompt

	p.Start()
	typeText(&p, "abc")

	result := p.HandleKey(tea.KeyPressMsg{Code: tea.KeyEscape})
	assert.Equal(t, PromptCanceled, result)
	assert.False(t, p.Active())
	assert.Empty(t, p.Text())
}

func TestTextPrompt_BackspaceDeletesRune(t *testing.T) {
	var p TextPrompt

	p.Start()
	typeText(&p, "café")

	result := p.HandleKey(tea.KeyPressMsg{Code: tea.KeyBackspace})
	assert.Equal(t, PromptPending, result)
	assert.Equal(t, "caf", p.Text(), "backspace removes the whole rune")
}

func TestTextPrompt_BackspacePastEmptyCancels(t *testing.T) {
	var p TextPrompt

	p.Start()

	result := p.HandleKey(tea.KeyPressMsg{Code: tea.KeyBackspace})
	assert.Equal(t, PromptCanceled, result)
	assert.False(t, p.Active())
}

func TestTextPrompt_IgnoresModifiedKeys(t *testing.T) {
	var p TextPrompt

	p.Start()
	p.HandleKey(tea.KeyPressMsg{Code: 'a', Text: "a", Mod: tea.ModCtrl})

	assert.Empty(t, p.Text())
}

func TestExtractPromptCursor(t *testing.T) {
	frame := "header\ncontent\n  " + PromptLabel("ab", false)

	x, y, cleaned, ok := ExtractPromptCursor(frame)

	assert.True(t, ok)
	assert.Equal(t, 2, y, "cursor sits on the last line")
	assert.Equal(t, 5, x, "two cells margin, the sigil, two typed cells")
	assert.NotContains(t, cleaned, promptCursorMarker, "the marker cell blanks out")
	assert.Contains(t, cleaned, "ab ", "a space takes the marker's cell")
}

func TestExtractPromptCursor_NoPromptOpen(t *testing.T) {
	frame := "header\ncontent\nfooter"

	_, _, cleaned, ok := ExtractPromptCursor(frame)

	assert.False(t, ok)
	assert.Equal(t, frame, cleaned)
}

// A marker rune typed into the query must not steal the cursor: the real
// cursor is always the appended, last occurrence.
func TestExtractPromptCursor_MarkerInInput(t *testing.T) {
	frame := "content\n" + PromptLabel("a"+promptCursorMarker+"b", false)

	x, _, _, ok := ExtractPromptCursor(frame)

	assert.True(t, ok)
	assert.Equal(t, 4, x, "sigil + three input cells put the cursor at cell 4")
}
