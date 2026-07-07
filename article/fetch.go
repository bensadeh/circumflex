package article

import (
	"bytes"
	"context"
	"fmt"
	nurl "net/url"
	"strings"
	"time"

	"github.com/bensadeh/circumflex/version"

	"codeberg.org/readeck/go-readability/v2"
	"golang.org/x/net/html"
	"resty.dev/v3"
)

const (
	fetchTimeout = 10 * time.Second
	retryCount   = 1
)

// discardLogger silences resty's internal logging so that WARN/ERROR
// messages on context cancellation don't corrupt the TUI.
type discardLogger struct{}

func (discardLogger) Errorf(string, ...any) {}
func (discardLogger) Warnf(string, ...any)  {}
func (discardLogger) Debugf(string, ...any) {}

func fetchPage(ctx context.Context, url string, parsedURL *nurl.URL) (body []byte, contentType string, err error) {
	client := resty.New()

	defer func() { _ = client.Close() }()

	client.SetTimeout(fetchTimeout)
	client.SetRetryCount(retryCount)
	client.SetHeader("User-Agent", version.Name+"/"+version.Version)
	client.SetLogger(discardLogger{})

	resp, err := client.R().SetContext(ctx).Get(url)
	if err != nil {
		if ctx.Err() != nil {
			return nil, "", ctx.Err()
		}

		return nil, "", fmt.Errorf("could not fetch URL: %w", err)
	}

	if resp.StatusCode() >= 400 {
		return nil, "", fmt.Errorf("server returned status %d for %s", resp.StatusCode(), parsedURL.Host)
	}

	return resp.Bytes(), resp.Header().Get("Content-Type"), nil
}

func extractReadable(body []byte, parsedURL *nurl.URL) (*html.Node, error) {
	a, err := readability.FromReader(bytes.NewReader(body), parsedURL)
	if err != nil {
		return nil, fmt.Errorf("could not parse article from %s: %w", parsedURL.Host, err)
	}

	if a.Node == nil {
		return nil, fmt.Errorf("could not extract readable content from %s", parsedURL.Host)
	}

	return a.Node, nil
}

// isPlainText sniffs the body as well as the header: some servers label HTML
// as text/plain, and rendering markup verbatim would be worse than reflowing.
func isPlainText(contentType string, body []byte) bool {
	if !strings.HasPrefix(contentType, "text/plain") {
		return false
	}

	head := strings.ToLower(string(body[:min(len(body), 256)]))

	return !strings.Contains(head, "<!doctype html") && !strings.Contains(head, "<html")
}
