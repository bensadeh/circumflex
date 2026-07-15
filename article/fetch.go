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

// fetchArticle retrieves the page reader mode will parse, preferring a known
// full-text mirror of the URL when one exists. The returned URL is the one
// actually fetched, so relative references resolve against the right base.
func fetchArticle(ctx context.Context, url string, parsedURL *nurl.URL) ([]byte, string, *nurl.URL, error) {
	if body, contentType, mirror := fetchFullText(ctx, parsedURL); mirror != nil {
		return body, contentType, mirror, nil
	}

	if ctx.Err() != nil {
		return nil, "", nil, ctx.Err()
	}

	body, contentType, err := fetchPage(ctx, url, parsedURL)

	return body, contentType, parsedURL, err
}

// fetchFullText returns a nil URL when no mirror is known for the page or the
// mirror did not serve it, e.g. an arXiv paper with no HTML conversion.
func fetchFullText(ctx context.Context, parsedURL *nurl.URL) ([]byte, string, *nurl.URL) {
	fullText := fullTextURL(parsedURL)
	if fullText == "" {
		return nil, "", nil
	}

	fullTextParsed, err := nurl.ParseRequestURI(fullText)
	if err != nil {
		return nil, "", nil
	}

	body, contentType, err := fetchPage(ctx, fullText, fullTextParsed)
	if err != nil {
		return nil, "", nil
	}

	return body, contentType, fullTextParsed
}

func extractReadable(body []byte, parsedURL *nurl.URL) (*html.Node, string, error) {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, "", fmt.Errorf("could not parse page from %s: %w", parsedURL.Host, err)
	}

	// MediaWiki markup needs normalizing before readability runs, while the
	// class names that identify it are still present.
	normalizeMediaWiki(doc)

	parser := readability.NewParser()

	parser.ClassesToPreserve = append(parser.ClassesToPreserve, latexmlPreservedClasses...)

	a, err := parser.ParseAndMutate(doc, parsedURL)
	if err != nil {
		return nil, "", fmt.Errorf("could not parse article from %s: %w", parsedURL.Host, err)
	}

	if a.Node == nil {
		return nil, "", fmt.Errorf("could not extract readable content from %s", parsedURL.Host)
	}

	return a.Node, a.Title(), nil
}

// isPlainText sniffs the body as well as the header: some servers label HTML
// as text/plain, and rendering markup verbatim would be worse than reflowing.
func isPlainText(contentType string, body []byte) bool {
	return strings.HasPrefix(contentType, "text/plain") && !looksLikeHTML(body)
}

func looksLikeHTML(body []byte) bool {
	head := strings.ToLower(string(body[:min(len(body), 256)]))

	return strings.Contains(head, "<!doctype html") || strings.Contains(head, "<html")
}
