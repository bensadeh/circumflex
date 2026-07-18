package graphics

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Enabled is process-global and cannot be unset, so the disabled-state
// assertions run before the lifecycle flips it — deliberately serial.
func TestEnableLifecycle(t *testing.T) {
	assert.False(t, Enabled())
	assert.Empty(t, CleanupSeq(), "nothing was transmitted, and an unprobed terminal must receive no APC at all")

	AllocID()

	assert.Empty(t, CleanupSeq(), "allocation alone transmits nothing")

	assert.True(t, Enable(), "the first answer is news")
	assert.False(t, Enable(), "repeats are not")
	assert.True(t, Enabled())

	cleanup := CleanupSeq()
	assert.Contains(t, cleanup, "a=d,d=I,i=1", "transmitted IDs are freed on exit")
}

func TestAllocIDStaysWithinOneByte(t *testing.T) {
	seen := map[int]bool{}

	for range 300 {
		id := AllocID()
		assert.GreaterOrEqual(t, id, 1)
		assert.LessOrEqual(t, id, 255, "IDs ride an indexed foreground color")
		seen[id] = true
	}

	assert.True(t, seen[255], "the counter walks the whole range")
	assert.False(t, seen[0], "0 means no ID in the protocol")
}

func TestSetCellSize(t *testing.T) {
	assert.False(t, SetCellSize(0, 0), "a multiplexer answering zeros is ignored")
	assert.False(t, SetCellSize(1, 1))

	assert.True(t, SetCellSize(10, 21))
	assert.False(t, SetCellSize(10, 21), "unchanged size is not news")
	assert.True(t, SetCellSize(12, 25), "a font-size change is")

	w, h := CellSize()
	assert.Equal(t, 12, w)
	assert.Equal(t, 25, h)
}

func TestShouldQuery(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("TERM", "xterm-ghostty")
	t.Setenv("TMUX", "")
	assert.True(t, ShouldQuery())

	t.Setenv("TERM_PROGRAM", "Apple_Terminal")
	assert.False(t, ShouldQuery(), "Terminal.app prints APC sequences to the screen")

	t.Setenv("TERM_PROGRAM", "WezTerm")
	assert.False(t, ShouldQuery(), "WezTerm answers the probe but draws no placeholders")

	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("TERM", "screen-256color")
	assert.False(t, ShouldQuery(), "GNU screen swallows the passthrough")

	t.Setenv("TMUX", "/tmp/tmux-1000/default,1234,0")
	assert.True(t, ShouldQuery(), "tmux reports a screen TERM but tunnels via passthrough")
}

func TestQuerySeqShape(t *testing.T) {
	t.Setenv("TMUX", "")

	seq := QuerySeq()
	assert.Contains(t, seq, "\x1b_G", "the graphics probe is an APC")
	assert.Contains(t, seq, "a=q", "query action, so nothing is stored")
	assert.Contains(t, seq, "i=31")
	assert.Contains(t, seq, "\x1b[16t", "cell pixel size rides along")
	assert.Less(t, strings.Index(seq, "\x1b[16t"), strings.Index(seq, "\x1b_G"),
		"cell size is asked first, so its report precedes the answer that enables graphics")

	assert.True(t, IsQueryReply(31))
	assert.False(t, IsQueryReply(7), "stray graphics responses are not the probe echo")
}

func TestTransmitSeqSingleChunk(t *testing.T) {
	t.Setenv("TMUX", "")

	seq := TransmitSeq(7, []byte("tiny"), 10, 5)

	assert.Equal(t, 2, strings.Count(seq, "\x1b_G"), "the id purge, then the payload in one piece")
	assert.Less(t, strings.Index(seq, "a=d,d=I,i=7"), strings.Index(seq, "a=T"),
		"the id is purged first: recycled IDs may collide with a dead session's leftovers")
	assert.Contains(t, seq, "a=T", "transmit-and-put creates the placement in one step")
	assert.Contains(t, seq, "f=100", "the payload is a PNG")
	assert.Contains(t, seq, "U=1", "virtual placement: pixels appear only under placeholder cells")
	assert.Contains(t, seq, "c=10,r=5")
	assert.Contains(t, seq, "q=2", "the terminal must not answer transmissions")
	assert.NotContains(t, seq, "m=", "no continuation marker on a single chunk")
}

func TestTransmitSeqChunks(t *testing.T) {
	t.Setenv("TMUX", "")

	seq := TransmitSeq(7, make([]byte, 9000), 10, 5)

	chunks := strings.Count(seq, "\x1b_G")
	assert.Equal(t, 4, chunks, "the id purge, then 9000 bytes base64 to 12000, split at 4096")

	assert.Less(t, strings.Index(seq, "a=d,d=I,i=7"), strings.Index(seq, "a=T"))
	assert.Equal(t, 1, strings.Count(seq, "a=T"), "options ride the first chunk only")
	assert.Equal(t, 2, strings.Count(seq, "m=1"), "all but the last chunk continue")
	assert.Equal(t, 1, strings.Count(seq, "m=0"), "the last chunk closes the stream")
}

func TestTransmitSeqTmuxPassthrough(t *testing.T) {
	t.Setenv("TMUX", "/tmp/tmux-1000/default,1234,0")

	seq := TransmitSeq(7, []byte("tiny"), 10, 5)

	require.True(t, strings.HasPrefix(seq, "\x1bPtmux;"), "tmux swallows bare APC sequences")
	assert.Contains(t, seq, "\x1b\x1b_G", "the inner escape is doubled for the tunnel")
}

func TestPlacementSeq(t *testing.T) {
	t.Setenv("TMUX", "")

	seq := PlacementSeq(9, 20, 8)

	deleteIdx := strings.Index(seq, "a=d,d=i,i=9")
	placeIdx := strings.Index(seq, "a=p")

	require.GreaterOrEqual(t, deleteIdx, 0)
	require.GreaterOrEqual(t, placeIdx, 0)
	assert.Less(t, deleteIdx, placeIdx, "the stale placement drops before the fresh one lands")

	assert.Contains(t, seq, "c=20,r=8")
	assert.NotContains(t, seq, "f=100", "no pixel data travels on a resize")
}
