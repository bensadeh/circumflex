package view

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/comment"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/view/message"

	tea "charm.land/bubbletea/v2"
	xansi "github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

// The OSC progress sequences the app writes on fetch would flood the test
// output thousands of times over across the geometry sweep.
func TestMain(m *testing.M) {
	progressOut = io.Discard

	os.Exit(m.Run())
}

// The geometry contract, enforced across a sweep of terminal sizes, wide-view
// configurations, and open surfaces:
//
//  1. Rendering never panics, down to a 1×1 terminal.
//  2. No line is ever wider than the terminal. Bubble Tea does not crop, so
//     an overwide line wraps and corrupts every row below it.
//  3. The wide layout fills the terminal exactly: every line exactly as wide
//     as the screen, exactly one line per row. The row-by-row pane join
//     depends on it.
//
// New surfaces (a pane, a footer, an overlay) get covered by adding them to
// geometrySurfaces.

var (
	// Degenerate sizes, the wide-view floor, the default wide threshold, and
	// off-by-one neighbors on both sides of each boundary.
	geometryWidths  = []int{1, 2, 3, 5, 10, 39, 40, 41, 79, 80, 120, 179, 180, 181, 250}
	geometryHeights = []int{1, 2, 3, 4, 5, 10, 24, 50}

	// Wide-view settings: the default threshold, "always" (0), "never".
	geometryWideMinWidths = []int{180, 0, math.MaxInt}
)

type geometrySurface struct {
	name string
	open func(m *model) *model
}

var geometrySurfaces = []geometrySurface{
	{"browsing", func(m *model) *model { return m }},
	{"comments", func(m *model) *model {
		m, _ = m.Update(message.CommentTreeDataReady{Thread: geometryThread(), FetchID: m.fetchID})

		return m
	}},
	{"reader", func(m *model) *model {
		m, _ = m.Update(message.ArticleReady{Parsed: geometryArticle(), Title: "A Story Title Long Enough To Truncate", FetchID: m.fetchID})

		return m
	}},
	{"help", func(m *model) *model {
		m.screen = screenHelp

		return m
	}},
	{"loading", func(m *model) *model {
		m, _ = m.Update(keyMsg("enter"))

		return m
	}},
	{"loadingreader", func(m *model) *model {
		m, _ = m.Update(keyMsg("space"))

		return m
	}},
	{"loaderror", func(m *model) *model {
		m, _ = m.Update(keyMsg("enter"))
		m, _ = m.Update(message.CommentTreeDataReady{
			Err:     errors.New("a load error message long enough to wrap in the narrowest detail pane"),
			FetchID: m.fetchID,
		})

		return m
	}},
	// J/K from an open story: the narrow layout keeps the story on screen and
	// overlays fetch feedback on its bottom row.
	{"adjacentloading", func(m *model) *model {
		m, _ = m.Update(message.CommentTreeDataReady{Thread: geometryThread(), FetchID: m.fetchID})
		m, _ = m.Update(message.OpenAdjacentStory{Direction: 1})

		return m
	}},
	{"adjacenterror", func(m *model) *model {
		m, _ = m.Update(message.CommentTreeDataReady{Thread: geometryThread(), FetchID: m.fetchID})
		m, _ = m.Update(message.OpenAdjacentStory{Direction: 1})
		m, _ = m.Update(message.CommentTreeDataReady{
			Err:     errors.New("dial tcp: lookup " + strings.Repeat("a-very-long-hostname.example.com.", 8) + ": no such host"),
			FetchID: m.fetchID,
		})

		return m
	}},
}

// geometryThread nests comments deep enough to hit the indent plateau and
// includes text long enough to wrap at every width.
func geometryThread() *comment.Thread {
	deep := &hn.CommentNode{ID: 10, Author: "alice", Content: strings.Repeat("deeply nested wrapping text ", 8)}
	for i := range 6 {
		deep = &hn.CommentNode{ID: 11 + i, Author: "bob", Content: "reply with a quote:<p><i>&gt; quoted line that should wrap</i>", Children: []*hn.CommentNode{deep}}
	}

	tree := &hn.CommentTree{
		ID:            1,
		Title:         "Show HN: A story title comfortably longer than the narrowest pane",
		Author:        "op",
		Content:       "Root story text with a link <a href=\"https://example.com\">example</a> and <code>inline code</code>.",
		CommentsCount: 9,
		Points:        42,
		URL:           "https://example.com/a/rather/long/path",
		Domain:        "example.com",
		Comments: []*hn.CommentNode{
			deep,
			{ID: 2, Author: "carol", Content: strings.Repeat("top level comment text ", 12)},
		},
	}

	return comment.ToThread(tree)
}

// geometryArticle exercises the reader's full-width paths (code blocks,
// tables) alongside wrapped prose and section headers.
func geometryArticle() *article.Parsed {
	return article.NewParsedFromHTML(`
		<h1>Heading</h1>
		<p>` + strings.Repeat("prose that wraps at any width ", 10) + `</p>
		<h2>Section</h2>
		<pre><code>func main() { fmt.Println("a code line wider than a narrow pane") }</code></pre>
		<table><tr><th>col one</th><th>col two</th></tr><tr><td>alpha</td><td>beta</td></tr></table>
	`)
}

func newGeometryModel(t *testing.T, width, height, wideMinWidth int) *model {
	t.Helper()

	m := newTestModel(t)
	m.config.WideViewMinWidth = wideMinWidth

	m, _ = m.Update(tea.WindowSizeMsg{Width: width, Height: height})

	m.list.SetItems(categories.Top, testItems())
	m.status.StopSpinner()
	m.fetching = false
	m.updatePagination()

	return m
}

func assertGeometry(t *testing.T, m *model) {
	t.Helper()

	lines := strings.Split(m.View(), "\n")

	for i, line := range lines {
		w := xansi.StringWidth(line)
		require.LessOrEqual(t, w, m.width, "line %d is %d cells wide: %q", i, w, xansi.Strip(line))

		if m.isWide() {
			require.Equal(t, m.width, w, "wide layout line %d must fill the width exactly", i)
		}
	}

	if m.isWide() {
		require.Len(t, lines, m.height, "wide layout must fill the height exactly")
	}
}

func TestGeometry_EverySurfaceAtEverySize(t *testing.T) {
	t.Parallel()

	for _, wideMin := range geometryWideMinWidths {
		for _, surface := range geometrySurfaces {
			for _, w := range geometryWidths {
				for _, h := range geometryHeights {
					t.Run(fmt.Sprintf("wideMin=%d/%s/%dx%d", wideMin, surface.name, w, h), func(t *testing.T) {
						t.Parallel()

						m := surface.open(newGeometryModel(t, w, h, wideMin))
						assertGeometry(t, m)
					})
				}
			}
		}
	}
}

// Resizing an open surface must uphold the same contract: the detail views
// rebuild their layout from the synthesized pane-sized WindowSizeMsg, and a
// resize across the wide threshold swaps layouts in place.
func TestGeometry_ResizeAfterOpen(t *testing.T) {
	t.Parallel()

	baseSizes := []struct{ w, h int }{{250, 30}, {60, 10}}

	for _, wideMin := range geometryWideMinWidths {
		for _, surface := range geometrySurfaces {
			for _, base := range baseSizes {
				for _, w := range geometryWidths {
					for _, h := range geometryHeights {
						t.Run(fmt.Sprintf("wideMin=%d/%s/from=%dx%d/to=%dx%d", wideMin, surface.name, base.w, base.h, w, h), func(t *testing.T) {
							t.Parallel()

							m := surface.open(newGeometryModel(t, base.w, base.h, wideMin))
							m, _ = m.Update(tea.WindowSizeMsg{Width: w, Height: h})
							assertGeometry(t, m)

							// Ensure the wide-view floor guard holds after
							// layout.WideViewFloor changes, not just at the
							// sizes in the sweep.
							if m.isWide() {
								require.GreaterOrEqual(t, m.width, layout.WideViewFloor)
							}
						})
					}
				}
			}
		}
	}
}
