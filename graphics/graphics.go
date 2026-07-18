// Package graphics tracks the terminal's Kitty graphics support and builds
// the escape sequences behind reader mode's high-resolution images. Pixels
// travel out-of-band as APC transmissions; on screen an image is ordinary
// styled text — Unicode placeholder cells the terminal composites pixels
// onto — so scrolling and cell diffing never see the difference.
package graphics

import (
	"encoding/base64"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	xansi "github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/kitty"
)

// queryID marks the capability probe so its echo is recognizable among
// graphics responses; transmissions are sent with q=2 and never answer.
const queryID = 31

var (
	enabled atomic.Bool

	// The terminal cell's pixel dimensions, for sizing an image's cell
	// grid to its aspect ratio; zero until the terminal reports them.
	cellWidth, cellHeight atomic.Int32

	mu        sync.Mutex
	nextID    int
	allocated = map[int]struct{}{}
)

// Enable records that the terminal answered the Kitty graphics probe. It
// reports whether this was news, so the caller knows a repaint is due.
func Enable() bool {
	return !enabled.Swap(true)
}

func Enabled() bool {
	return enabled.Load()
}

// SetCellSize records the terminal cell's pixel dimensions and reports
// whether they changed. Degenerate reports (a multiplexer answering zeros)
// are ignored.
func SetCellSize(width, height int) bool {
	if width <= 1 || height <= 1 || width > 1<<14 || height > 1<<14 {
		return false
	}

	oldWidth := cellWidth.Swap(int32(width))
	oldHeight := cellHeight.Swap(int32(height))

	return oldWidth != int32(width) || oldHeight != int32(height)
}

func CellSize() (width, height int) {
	return int(cellWidth.Load()), int(cellHeight.Load())
}

// AllocID claims a terminal-global image ID. IDs stay within one byte so
// placeholder cells can carry them as an indexed foreground color, immune
// to color-profile downsampling; the wrap at 255 recycles IDs from articles
// long scrolled away.
func AllocID() int {
	mu.Lock()
	defer mu.Unlock()

	nextID++
	if nextID > 255 {
		nextID = 1
	}

	allocated[nextID] = struct{}{}

	return nextID
}

// ShouldQuery reports whether the capability probe is safe and worthwhile.
// Apple's Terminal.app prints APC sequences to the screen instead of
// consuming them; WezTerm answers the probe but does not draw Unicode
// placeholders; GNU screen swallows the passthrough entirely.
func ShouldQuery() bool {
	termProgram := os.Getenv("TERM_PROGRAM")
	if strings.Contains(termProgram, "Apple") || termProgram == "WezTerm" {
		return false
	}

	if strings.HasPrefix(os.Getenv("TERM"), "screen") && os.Getenv("TMUX") == "" {
		return false
	}

	return true
}

// QuerySeq is the startup probe: the cell pixel size request, then a dummy
// one-pixel query the terminal answers only if it speaks the Kitty graphics
// protocol. The cell size goes first because terminals answer in input order:
// its report lands before the answer that turns graphics on, so the enabling
// repaint sizes image grids from real cell metrics instead of transmitting
// at the fallback shape and correcting a moment later. Terminals that
// recognize neither consume both silently, so no answer simply leaves the
// feature off.
func QuerySeq() string {
	return CellSizeQuerySeq() +
		wrap(xansi.KittyGraphics([]byte("AAAA"),
			"i="+strconv.Itoa(queryID), "s=1", "v=1", "a=q", "t=d", "f=24"))
}

// IsQueryReply reports whether a graphics response echoes the probe.
func IsQueryReply(id int) bool {
	return id == queryID
}

// CellSizeQuerySeq asks for the cell's pixel dimensions (XTWINOPS 16), so a
// font-size change keeps image geometry honest across resizes.
func CellSizeQuerySeq() string {
	return xansi.WindowOp(16)
}

// TransmitSeq transfers a PNG to the terminal and creates its virtual
// placement: the image will show wherever placeholder cells carrying id
// appear, scaled to cols x rows cells. The payload is chunked to the
// protocol's limit; q=2 keeps the terminal from answering.
//
// The transmission opens by deleting the id: IDs recycle across sessions and
// wrap at 255 within one, and the terminal may still hold an image a crashed
// or killed session never cleaned up. Deleting first means no placement or
// image from another life can survive alongside the fresh transmission;
// deleting a nonexistent id is a no-op.
func TransmitSeq(id int, png []byte, cols, rows int) string {
	payload := base64.StdEncoding.EncodeToString(png)

	idOpt := "i=" + strconv.Itoa(id)
	deleteSeq := wrap(xansi.KittyGraphics(nil, "a=d", "d=I", idOpt, "q=2"))

	baseOpts := []string{
		"a=T", "f=100", "t=d", "q=2", "U=1", idOpt,
		"c=" + strconv.Itoa(cols),
		"r=" + strconv.Itoa(rows),
	}

	if len(payload) <= kitty.MaxChunkSize {
		return deleteSeq + wrap(xansi.KittyGraphics([]byte(payload), baseOpts...))
	}

	var sb strings.Builder

	sb.WriteString(deleteSeq)

	for off := 0; off < len(payload); off += kitty.MaxChunkSize {
		end := min(off+kitty.MaxChunkSize, len(payload))

		var opts []string

		switch {
		case off == 0:
			opts = append(slices.Clone(baseOpts), "m=1")

		case end == len(payload):
			opts = []string{"q=2", "m=0"}

		default:
			opts = []string{"q=2", "m=1"}
		}

		sb.WriteString(wrap(xansi.KittyGraphics([]byte(payload[off:end]), opts...)))
	}

	return sb.String()
}

// PlacementSeq resizes an already-transmitted image to a new cell grid: the
// old placement is dropped and a fresh virtual one created, no pixel data
// re-sent. This is the whole cost of a terminal resize.
func PlacementSeq(id, cols, rows int) string {
	idOpt := "i=" + strconv.Itoa(id)

	return wrap(xansi.KittyGraphics(nil, "a=d", "d=i", idOpt, "q=2")) +
		wrap(xansi.KittyGraphics(nil, "a=p", "U=1", idOpt, "q=2",
			"c="+strconv.Itoa(cols),
			"r="+strconv.Itoa(rows)))
}

// CleanupSeq deletes every image this session transmitted, freeing the
// terminal's image memory on exit. Empty when the terminal never enabled
// graphics — nothing was transmitted, and a terminal that failed the probe
// should not receive APC sequences at all.
func CleanupSeq() string {
	if !Enabled() {
		return ""
	}

	mu.Lock()
	defer mu.Unlock()

	var sb strings.Builder

	for _, id := range slices.Sorted(maps.Keys(allocated)) {
		sb.WriteString(wrap(xansi.KittyGraphics(nil, "a=d", "d=I", "i="+strconv.Itoa(id), "q=2")))
	}

	return sb.String()
}

// wrap tunnels a sequence through tmux, which would otherwise swallow the
// APC. The placeholder cells themselves are plain text and need no wrapping.
func wrap(seq string) string {
	if os.Getenv("TMUX") == "" {
		return seq
	}

	return xansi.TmuxPassthrough(seq)
}
