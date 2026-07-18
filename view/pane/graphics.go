package pane

import (
	"github.com/bensadeh/circumflex/graphics"

	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
)

// DetectGraphics probes the terminal for the Kitty graphics protocol and its
// cell pixel size, enabling high-resolution reader images. Terminals that
// answer confirm through HandleGraphicsReport; no answer leaves the
// half-block art. Terminals the probe would misfire on return nil.
func DetectGraphics() tea.Cmd {
	if !graphics.ShouldQuery() {
		return nil
	}

	return tea.Raw(graphics.QuerySeq())
}

// QueryCellSize re-asks for the cell pixel size on a resize, so a font-size
// change re-derives image geometry. Nothing to keep honest before the
// graphics probe succeeded.
func QueryCellSize() tea.Cmd {
	if !graphics.Enabled() {
		return nil
	}

	return tea.Raw(graphics.CellSizeQuerySeq())
}

// HandleGraphicsReport records a terminal's answer to the graphics probes
// and reports whether the state changed — the caller then repaints an open
// reader so its images pick up the new mode or geometry.
func HandleGraphicsReport(msg tea.Msg) bool {
	switch msg := msg.(type) {
	case uv.KittyGraphicsEvent:
		return graphics.IsQueryReply(msg.Options.ID) && graphics.Enable()

	case uv.CellSizeEvent:
		return graphics.SetCellSize(msg.Width, msg.Height)
	}

	return false
}
