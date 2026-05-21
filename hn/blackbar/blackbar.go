package blackbar

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bensadeh/circumflex/version"
)

// HN renders a black bar above the orange header only when commemorating a
// death. The bar is a <tr> with this background, inserted before the orange
// header row. Detecting the former before the latter is the signal.
const (
	blackMarker  = `bgcolor="#000000"`
	orangeMarker = `bgcolor="#ff6600"`
	// The markers sit ~470 bytes into the page today; this cap covers the
	// whole front page (~35 KB) with generous headroom while still bounding a
	// pathologically large response.
	readLimit = 64 << 10
)

var baseURL = "https://news.ycombinator.com/"

// Detect reports whether the Hacker News memorial black bar is currently
// rendered on the front page. A non-nil error means the status could not be
// determined; callers should treat that as "no bar" rather than surfacing it
// mid-session.
func Detect(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
	if err != nil {
		return false, fmt.Errorf("building request: %w", err)
	}

	req.Header.Set("User-Agent", version.Name+"/"+version.Version)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("fetching front page: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("front page returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, readLimit))
	if err != nil {
		return false, fmt.Errorf("reading front page: %w", err)
	}

	return hasBlackBar(string(body)), nil
}

func hasBlackBar(html string) bool {
	black := strings.Index(html, blackMarker)
	if black == -1 {
		return false
	}

	orange := strings.Index(html, orangeMarker)

	return orange == -1 || black < orange
}
