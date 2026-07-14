package pane

import (
	"context"
	"time"
)

// FetchToken is a fetch's identity: the context its commands must fetch under
// and the id their results must carry. Only Begin hands one out, so a fetch
// command can only be built for the fetch actually starting. A token held
// across Update cycles goes stale — results built from it fail the Finish
// guard and are dropped.
type FetchToken struct {
	Ctx context.Context //nolint:containedctx // fetch-scoped, never stored beyond the Update cycle
	ID  uint64
}

// FetchGuard owns the identity and cancellation of the single fetch a shell
// allows in flight. Begin invalidates any predecessor and hands out the new
// fetch's token; Finish drops stale results; Abort cancels and moves to a
// new era, so a result the fetch managed to deliver before the cancel took
// hold can never match.
type FetchGuard struct {
	ctx    context.Context //nolint:containedctx // single active fetch, accessed only from the Update goroutine
	cancel context.CancelFunc

	// id is the current fetch era, bumped by Begin and Abort. Results and
	// era-scoped side effects carry the era they were made in, so anything
	// stamped with an older one is ignored.
	id     uint64
	active bool
}

func (g *FetchGuard) InFlight() bool { return g.active }

// CurrentID stamps era-scoped side effects made outside Begin (e.g. an
// error-display timeout) so they can be ignored once a newer fetch begins.
func (g *FetchGuard) CurrentID() uint64 { return g.id }

// Begin starts a fetch's lifecycle, invalidating any predecessor. A zero
// timeout means the fetch runs until finished or aborted.
func (g *FetchGuard) Begin(timeout time.Duration) FetchToken {
	if g.cancel != nil {
		g.cancel()
	}

	g.id++
	g.active = true

	if timeout > 0 {
		g.ctx, g.cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		g.ctx, g.cancel = context.WithCancel(context.Background())
	}

	return FetchToken{Ctx: g.ctx, ID: g.id}
}

// Finish ends the in-flight fetch if id belongs to it; false is a stale
// result the caller must drop.
func (g *FetchGuard) Finish(id uint64) bool {
	if !g.active || id != g.id {
		return false
	}

	if g.cancel != nil {
		g.cancel()
	}

	g.active = false

	return true
}

// Abort cancels the in-flight fetch; false means nothing was in flight.
func (g *FetchGuard) Abort() bool {
	if !g.active {
		return false
	}

	if g.cancel != nil {
		g.cancel()
	}

	g.id++
	g.active = false

	return true
}
