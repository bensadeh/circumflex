package view

import (
	"testing"

	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/comment"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/version"
	"github.com/bensadeh/circumflex/view/message"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWindowTitle_FrontPageNamesTheApp(t *testing.T) {
	m := newTestModelReady(t)

	assert.Equal(t, version.Name, m.windowTitle())
}

func TestWindowTitle_OpenStory(t *testing.T) {
	m := newTestModelReady(t)
	startTestFetch(m, screenComments)

	thread := &comment.Thread{Story: hn.Story{ID: 1, Title: "First item", CommentsCount: 5}}
	m, _ = m.Update(message.CommentTreeDataReady{Thread: thread, FetchID: m.fetch.currentID()})
	require.NotNil(t, m.detail)

	assert.Equal(t, "First item", m.windowTitle())

	m, _ = m.Update(message.DetailQuit{})
	assert.Equal(t, version.Name, m.windowTitle(), "leaving the story gives the window back to the app")
}

// A story fetch renames the window before its result lands: the detail pane is
// already showing that story's loading state.
func TestWindowTitle_LoadingStory(t *testing.T) {
	m := newTestModelReady(t)

	m, _ = m.Update(keyMsg("enter"))
	require.True(t, m.fetch.detailLoading())

	assert.Equal(t, "First item", m.windowTitle())
}

// A followed link swaps a new page into the detail pane; the window names
// the page on screen, not the selected story behind the trail.
func TestWindowTitle_FollowedLinkNamesThePage(t *testing.T) {
	m := openTestReader(t, newTestModelReady(t))

	_, _ = m.startLinkFetch(0)
	m, _ = m.Update(message.LinkArticleReady{
		Parsed:  testParsedArticle(),
		Title:   "Linked Page",
		URL:     "https://example.com/linked",
		Trail:   []message.TrailEntry{{URL: "https://example.com/1", Story: true}},
		FetchID: m.fetch.currentID(),
	})

	assert.Equal(t, "Linked Page", m.windowTitle())

	// Walking back renames the window to the page restored.
	m, _ = m.Update(message.RestorePage{
		Entry: message.TrailEntry{URL: "https://example.com/1", Title: "First item", Parsed: testParsedArticle(), Story: true},
	})
	assert.Equal(t, "First item", m.windowTitle())
}

func TestWindowTitle_LinkedThreadNamesTheThread(t *testing.T) {
	m := openTestReader(t, newTestModelReady(t))

	_, _ = m.startLinkFetch(0)
	m, _ = m.Update(message.LinkCommentsReady{
		Thread:  &comment.Thread{Story: hn.Story{ID: 9, Title: "Linked thread", CommentsCount: 5}},
		Trail:   []message.TrailEntry{{URL: "https://example.com/1", Story: true}},
		FetchID: m.fetch.currentID(),
	})

	assert.Equal(t, "Linked thread", m.windowTitle())
}

func TestWindowTitle_SearchQuery(t *testing.T) {
	m := newTestModelReady(t)

	m, _ = m.Update(keyMsg("/"))
	for _, r := range "rust" {
		m, _ = m.Update(keyMsg(string(r)))
	}

	m, _ = m.Update(keyMsg("enter"))
	m, _ = m.Update(message.StoriesReady{Category: categories.Search, Index: -1, FetchID: m.fetch.currentID()})

	assert.Equal(t, "search: rust", m.windowTitle())
}

// The window title is a sink: a title that reached the list unstripped must
// not close the OSC sequence early or break the line.
func TestWindowTitle_NeutralizesHostileTitles(t *testing.T) {
	m := newTestModelReady(t)

	hostile := []*hn.Story{{ID: 1, Title: "Evil\x07\x1b]0;pwned\x07 story\nline two"}}

	_, _ = m.startFetch(m.listRollback())
	m, _ = m.Update(message.StoriesReady{Stories: hostile, Category: categories.Top, FetchID: m.fetch.currentID()})

	startTestFetch(m, screenComments)

	thread := &comment.Thread{Story: hn.Story{ID: 1, Title: hostile[0].Title}}
	m, _ = m.Update(message.CommentTreeDataReady{Thread: thread, FetchID: m.fetch.currentID()})

	assert.Equal(t, "Evil story line two", m.windowTitle())
}
