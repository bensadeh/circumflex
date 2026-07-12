package comments

import (
	"fmt"
	"testing"

	"github.com/bensadeh/circumflex/comment"
)

// benchThread approximates a large front-page thread: 100 top-level comments,
// each with a reply chain of 10, bodies long enough to wrap.
func benchThread() *comment.Thread {
	comments := make([]*comment.Comment, 0, 100)
	id := 1

	for i := range 100 {
		var chain *comment.Comment

		for d := 10; d >= 1; d-- {
			c := newComment(id, fmt.Sprintf("author%d", id),
				fmt.Sprintf("reply %d at depth %d — the quick brown fox jumps over the lazy dog, "+
					"pack my box with five dozen liquor jugs, needle in the haystack", id, d))
			id++

			if chain != nil {
				c.Children = []*comment.Comment{chain}
			}

			chain = c
		}

		top := newComment(id, fmt.Sprintf("top%d", i),
			fmt.Sprintf("top-level comment %d with enough words to wrap across a couple of rendered lines "+
				"in the default comment column width, mentioning the needle too", i))
		id++
		top.Children = []*comment.Comment{chain}
		comments = append(comments, top)
	}

	return newThread(comments...)
}

func benchModel(b *testing.B, search bool) *Model {
	b.Helper()

	m := New(benchThread(), 0, 80, 1, false, 130, 45)

	expandAll(m.flat)
	m.rebuildContent()

	if search {
		commitCommentSearch(m, "needle")
	}

	m.toggleMode()

	return m
}

// The focus-move path: must stay off the O(document) content push.
func BenchmarkNavigateKeystroke(b *testing.B) {
	m := benchModel(b, false)

	for i := 0; b.Loop(); i++ {
		if i%40 < 20 {
			m.navigateComment(1)
		} else {
			m.navigateComment(-1)
		}
	}
}

func BenchmarkNavigateKeystrokeWithSearch(b *testing.B) {
	m := benchModel(b, true)

	for i := 0; b.Loop(); i++ {
		if i%40 < 20 {
			m.navigateComment(1)
		} else {
			m.navigateComment(-1)
		}
	}
}

// The structural path (fold, reveal, resize): allowed to cost the content push.
func BenchmarkUpdateViewport(b *testing.B) {
	m := benchModel(b, true)

	for b.Loop() {
		m.updateViewport()
	}
}

func BenchmarkViewWithSearch(b *testing.B) {
	m := benchModel(b, true)

	for b.Loop() {
		_ = m.View()
	}
}
