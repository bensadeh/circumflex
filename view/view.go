// Package view is the application coordinator: it owns the state machine,
// routes messages, orchestrates fetches and lays out the panes. The story
// list, comment section and reader render under its direction.
package view

import (
	"fmt"
	"os"

	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/favorites"
	"github.com/bensadeh/circumflex/graphics"
	"github.com/bensadeh/circumflex/hn/provider"
	"github.com/bensadeh/circumflex/settings"
	"github.com/bensadeh/circumflex/view/pane"

	tea "charm.land/bubbletea/v2"
)

// teaModel adapts model's pointer-based Update to the tea.Model interface.
type teaModel struct {
	m *model
}

func (t teaModel) Init() tea.Cmd {
	// The background feeds image transparency in reader mode, the foreground
	// its URL selector's separator row; terminals that do not answer simply
	// never deliver the messages.
	return tea.Batch(tea.RequestBackgroundColor, tea.RequestForegroundColor,
		pane.DetectStyledUnderline(), pane.DetectGraphics())
}

func (t teaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	_, cmd := t.m.Update(msg)

	return t, cmd
}

func (t teaModel) View() tea.View {
	frame := t.m.View()

	// An open search prompt marks its cursor cell in the frame; the real
	// terminal cursor parks there, in the terminal's own color — a steady
	// vertical bar, not a blinking block.
	var cursor *tea.Cursor

	if x, y, cleaned, ok := pane.ExtractPromptCursor(frame); ok {
		frame = cleaned
		cursor = tea.NewCursor(x, y)
		cursor.Shape = tea.CursorBar
		cursor.Blink = false
	}

	v := tea.NewView(frame)
	v.AltScreen = true
	v.Cursor = cursor
	v.WindowTitle = t.m.windowTitle()

	return v
}

func Run(config *settings.Config, cat *categories.Categories) error {
	fav, err := favorites.New(settings.FavoritesPath(), settings.LegacyFavoritesPath())
	if err != nil {
		return err
	}

	hist, err := newHistory(config.DebugMode || config.DebugFallible, config.DoNotMarkSubmissionsAsRead)
	if err != nil {
		return err
	}

	service := provider.NewService(config.DebugMode, config.DebugFallible)
	m := newModel(config, cat, fav, 0, 0, service, hist)

	p := tea.NewProgram(teaModel{m: m})

	restoreTitle := pane.SaveWindowTitle()
	settleProgress := pane.WireProgress(p)
	stopGraphics := pane.WireGraphics(p)

	finalModel, err := p.Run()

	settleProgress()
	stopGraphics()
	restoreTitle()

	// Transmitted images survive the program in the terminal's memory;
	// release them now that no frame flush can interleave with the write.
	if seq := graphics.CleanupSeq(); seq != "" {
		_, _ = fmt.Fprint(os.Stdout, seq)
	}

	if err != nil {
		return err
	}

	if fm, ok := finalModel.(teaModel); ok {
		if memErr := fm.m.memorialErr; memErr != nil {
			fmt.Fprintf(os.Stderr, "circumflex: could not check HN memorial status: %v\n", memErr)
		}

		if browserErr := fm.m.browserErr; browserErr != nil {
			fmt.Fprintf(os.Stderr, "circumflex: could not open browser: %v\n", browserErr)
		}
	}

	return nil
}
