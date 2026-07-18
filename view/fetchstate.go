package view

import (
	"context"
	"time"

	"github.com/bensadeh/circumflex/view/pane"

	tea "charm.land/bubbletea/v2"
)

// fetchToken is a fetch's identity: the context its commands must fetch under
// and the id their results must carry. Only begin hands one out, so a fetch
// command can only be built for the fetch actually starting. A token held
// across Update cycles goes stale — results built from it fail the finish
// guard and are dropped.
type fetchToken struct {
	ctx context.Context //nolint:containedctx // fetch-scoped, never stored beyond the Update cycle
	id  uint64
}

type fetchKind int

const (
	fetchNone   fetchKind = iota
	fetchList             // replaces the front page's stories
	fetchDetail           // loads a story's comments or article
	fetchLink             // loads a link followed from inside an article; the article stays visible
)

// rollbackPoint is what a failed or cancelled fetch restores. The caller
// captures it when the fetch starts — explicitly, so a caller that moves the
// selection first (J/K story navigation, refresh's cursor reset) passes the
// position being left, not the one arriving.
type rollbackPoint struct {
	categoryIndex int
	storyIndex    int // -1 for list fetches: page and cursor carry the restore
	page          int // list fetches: the page to return to
	cursor        int // list fetches: the cursor to return to
}

// fetchState owns the lifecycle of the single fetch the app allows in flight:
// identity, cancellation and the stale-result guard live in the shared
// pane.FetchGuard; what the app adds is which kind of fetch it is, the view
// it opens, and what a failure restores.
type fetchState struct {
	guard pane.FetchGuard
	kind  fetchKind

	target   screen // the view a detail fetch opens; the loading placeholder matches its meta block
	rollback rollbackPoint
}

func (f *fetchState) inFlight() bool      { return f.guard.InFlight() }
func (f *fetchState) detailLoading() bool { return f.guard.InFlight() && f.kind == fetchDetail }
func (f *fetchState) linkLoading() bool   { return f.guard.InFlight() && f.kind == fetchLink }
func (f *fetchState) currentID() uint64   { return f.guard.CurrentID() }

// begin starts a fetch's lifecycle, invalidating any predecessor.
func (f *fetchState) begin(timeout time.Duration, kind fetchKind, target screen, rb rollbackPoint) fetchToken {
	tok := f.guard.Begin(timeout)

	f.kind = kind
	f.target = target
	f.rollback = rb

	return fetchToken{ctx: tok.Ctx, id: tok.ID}
}

// finish ends the in-flight fetch if id belongs to it; ok=false is a stale
// result the caller must drop. The returned rollback is the restore point the
// fetch was started with.
func (f *fetchState) finish(id uint64) (rollbackPoint, bool) {
	if !f.guard.Finish(id) {
		return rollbackPoint{}, false
	}

	return f.rollback, true
}

// abort cancels the in-flight fetch and moves to a new era, so a result the
// fetch managed to deliver before the cancel took hold can never match.
// ok=false means nothing was in flight and there is no rollback to apply.
func (f *fetchState) abort() (rollbackPoint, bool) {
	if !f.guard.Abort() {
		return rollbackPoint{}, false
	}

	return f.rollback, true
}

// startFetch begins a list fetch: the stories replace the front page, and rb
// is the selection to restore if the fetch fails or is cancelled. List
// fetches carry no timeout — cancellation is the user's. The token is only
// valid in this Update cycle: build the fetch command now, or a cancel or
// newer fetch in between would leave the command's results stale.
func (m *model) startFetch(rb rollbackPoint) (fetchToken, tea.Cmd) {
	return m.fetch.begin(0, fetchList, screenList, rb), m.status.StartSpinner()
}

// startDetailFetch begins a fetch of a story's comments or article: the list
// stays in place, dimmed, while the detail loads. target is the view the
// fetch opens, so the loading pane can lay out what that view will draw.
func (m *model) startDetailFetch(timeout time.Duration, target screen, rb rollbackPoint) (fetchToken, tea.Cmd) {
	return m.fetch.begin(timeout, fetchDetail, target, rb), m.status.StartSpinner()
}

// startLinkFetch begins a fetch of a link followed from inside an article.
// The open reader stays on screen until the page arrives — failure surfaces
// as a status message — so the rollback point is the selection as it stands.
func (m *model) startLinkFetch(timeout time.Duration) (fetchToken, tea.Cmd) {
	return m.fetch.begin(timeout, fetchLink, screenReader, m.detailRollback(m.list.Index())), m.status.StartSpinner()
}

// listRollback is the restore point for a fetch that replaces the list: the
// category, page and cursor being left. Capture it before advancing to the
// incoming category — and before any pre-fetch cursor reset.
func (m *model) listRollback() rollbackPoint {
	return rollbackPoint{
		categoryIndex: m.cat.CurrentIndex(),
		storyIndex:    -1,
		page:          m.list.Page(),
		cursor:        m.list.Cursor(),
	}
}

// detailRollback is the restore point for a story fetch: the current category
// plus the story the reading marker moves back to on failure.
func (m *model) detailRollback(storyIndex int) rollbackPoint {
	return rollbackPoint{categoryIndex: m.cat.CurrentIndex(), storyIndex: storyIndex}
}

// rollbackFetch recovers from a failed or cancelled fetch: it restores the
// category selection and unfreezes the list. For a story fetch it moves the
// list selection back to the story that is still open, so the reading
// marker never points at a story the detail view doesn't show. For a list
// fetch it restores the page and cursor, undoing the reset refresh and
// search apply before fetching.
func (m *model) rollbackFetch(rb rollbackPoint) {
	m.cat.SetIndex(rb.categoryIndex)

	if rb.storyIndex >= 0 {
		m.list.SetIndex(rb.storyIndex)
	} else {
		m.list.SetPage(rb.page)
		m.list.SetCursorClamped(rb.cursor)
	}

	m.list.EndTransition()
}

// finishFetch settles the bookkeeping every fetch result shares: the stale
// guard, the terminal progress indicator, the spinner. ok=false is a stale
// result the caller must drop without touching anything.
func (m *model) finishFetch(id uint64, err error) (rollbackPoint, bool) {
	rb, ok := m.fetch.finish(id)
	if !ok {
		return rollbackPoint{}, false
	}

	pane.SyncProgress(err)
	m.status.StopSpinner()

	return rb, true
}

// abortFetchOnQuit kills a fetch racing a detail view's quit — a J/K story
// fetch or a link follow minted a cycle before the quit landed — so its
// result cannot reopen a story the user just left. Quitting is not
// cancelling: no status message.
func (m *model) abortFetchOnQuit() {
	rb, ok := m.fetch.abort()
	if !ok {
		return
	}

	m.rollbackFetch(rb)

	pane.ClearProgress()
	m.status.StopSpinner()
	m.updatePagination()
}

func (m *model) handleCancelFetch() tea.Cmd {
	rb, ok := m.fetch.abort()
	if !ok {
		return nil
	}

	m.rollbackFetch(rb)

	pane.ClearProgress()
	m.status.StopSpinner()
	// The screen stays where the fetch started: canceling a J/K story fetch
	// keeps the open story, canceling a category fetch keeps the front page.
	m.updatePagination()

	return m.status.NewStatusMessageWithDuration(pane.CancelledStatus(), statusMessageShort)
}
