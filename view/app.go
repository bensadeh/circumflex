package view

import (
	"image/color"
	"time"

	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/favorites"
	"github.com/bensadeh/circumflex/history"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/settings"
	"github.com/bensadeh/circumflex/view/list"
	"github.com/bensadeh/circumflex/view/pane"

	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
)

const (
	statusBarEdgeWidth = 5
	statusMessageShort = 2 * time.Second
	statusMessageLong  = 3 * time.Second
)

type model struct {
	screen  screen
	fetch   fetchState
	prompt  prompt
	started bool

	status statusBar
	width  int
	height int
	termBG color.Color // terminal background, reported once at startup; nil if unanswered

	list *list.Model

	// detail is the open comment section or reader view, nil while browsing.
	// Its nil-ness is the single source of truth for "a story is open";
	// screen still says which view it is. Both render into the detail pane,
	// which is the whole terminal when the wide layout is off.
	detail pane.View

	history   history.History
	config    *settings.Config
	service   hn.Service
	favorites *favorites.Favorites
	cat       *categories.Categories
	keymap    keyMap

	helpViewport viewport.Model

	memorialErr error
	browserErr  error

	// Cached styles for hot-path rendering.
	statusLeftStyle lipgloss.Style
	statusMidStyle  lipgloss.Style
	statusEndStyle  lipgloss.Style
}

func newModel(config *settings.Config, cat *categories.Categories, fav *favorites.Favorites, width, height int, service hn.Service, hist history.History) *model {
	m := &model{
		screen: screenList,
		width:  width,
		height: height,
		status: statusBar{spinner: newSpinner()},

		history:   hist,
		config:    config,
		service:   service,
		favorites: fav,
		cat:       cat,
		keymap:    defaultKeyMap(),

		statusLeftStyle: lipgloss.NewStyle().Inline(true).Width(statusBarEdgeWidth).MaxWidth(statusBarEdgeWidth),
		statusMidStyle:  lipgloss.NewStyle().Inline(true).Align(lipgloss.Center),
		statusEndStyle:  lipgloss.NewStyle().Inline(true).Width(statusBarEdgeWidth).Align(lipgloss.Center),
	}

	m.list = list.New(config, cat, hist)
	m.updatePagination()

	return m
}

func (m *model) setSize(width, height int) {
	m.width = width
	m.height = height
	m.updatePagination()
}

// updatePagination gives the list pane its width and the rows left between
// the header and the status bar, repaginating to fit.
func (m *model) updatePagination() {
	f := m.frame()
	m.list.Resize(f.ListWidth(), f.PaneContentHeight())
}

// listFrame collects the per-render facts the list pane cannot know itself.
func (m *model) listFrame() list.Frame {
	f := list.Frame{
		DetailOpen:    m.detail != nil || m.screen == screenHelp,
		DetailLoading: m.detailLoading(),
	}

	switch m.prompt {
	case promptAddFavorite:
		f.Selection = list.SelectionAddFavorite
	case promptRemoveFavorite:
		f.Selection = list.SelectionRemoveFavorite
	case promptNone:
		// Normal selection highlight.
	}

	return f
}

func (m *model) syncFavorites() {
	m.list.SetItems(categories.Favorites, favItemsToStories(m.favorites.Items()))
}

func favItemsToStories(items []*favorites.Item) []*hn.Story {
	stories := make([]*hn.Story, len(items))

	for i, it := range items {
		stories[i] = &hn.Story{
			ID:            it.ID,
			Title:         it.Title,
			Points:        it.Points,
			Author:        it.Author,
			Time:          it.Time,
			URL:           it.URL,
			Domain:        it.Domain,
			CommentsCount: it.CommentsCount,
		}
	}

	return stories
}
