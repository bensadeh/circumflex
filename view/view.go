// Package view is the application coordinator: it owns the state machine,
// routes messages, orchestrates fetches and lays out the panes. The story
// list, comment section and reader render under its direction.
package view

import (
	"fmt"
	"os"

	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/favorites"
	"github.com/bensadeh/circumflex/hn/provider"
	"github.com/bensadeh/circumflex/settings"

	tea "charm.land/bubbletea/v2"
)

// teaModel adapts model's pointer-based Update to the tea.Model interface.
type teaModel struct {
	m *model
}

func (t teaModel) Init() tea.Cmd {
	// The response feeds image transparency in reader mode; terminals that
	// do not answer simply never deliver the message.
	return tea.RequestBackgroundColor
}

func (t teaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	_, cmd := t.m.Update(msg)

	return t, cmd
}

func (t teaModel) View() tea.View {
	v := tea.NewView(t.m.View())
	v.AltScreen = true

	return v
}

func Run(config *settings.Config, cat *categories.Categories) error {
	fav, err := favorites.New(settings.FavoritesPath())
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

	// Progress sequences ride the program's message loop (see emitProgress).
	// The forwarder goroutine exists because Send would deadlock called from
	// Update itself; once the program stops, Send is a no-op.
	ch := make(chan string, 64)
	forwarderDone := make(chan struct{})
	progressCh = ch

	go func() {
		for {
			select {
			case seq := <-ch:
				p.Send(tea.RawMsg{Msg: seq})
			case <-forwarderDone:
				return
			}
		}
	}()

	finalModel, err := p.Run()

	close(forwarderDone)

	// The program has stopped writing, so a direct write can no longer
	// interleave with a frame. Clearing here rather than in the quit paths
	// guarantees the indicator never outlives the app, whatever the exit.
	_, _ = fmt.Fprint(os.Stderr, progressClearSeq)

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
