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

func fetchPage(ctx context.Context, url string, parsedURL *nurl.URL) (body []byte, contentType string, finalURL *nurl.URL, err error) {
	client := resty.New()

	defer func() { _ = client.Close() }()

	client.SetTimeout(fetchTimeout)
	client.SetRetryCount(retryCount)
	client.SetHeader("User-Agent", version.Name+"/"+version.Version)
	client.SetLogger(discardLogger{})

	resp, err := client.R().SetContext(ctx).Get(url)
	if err != nil {
		if ctx.Err() != nil {
			return nil, "", nil, ctx.Err()
		}

		return nil, "", nil, fmt.Errorf("could not fetch URL: %w", err)
	}

	if resp.StatusCode() >= 400 {
		return nil, "", nil, fmt.Errorf("server returned status %d for %s", resp.StatusCode(), parsedURL.Host)
	}

	// A redirect can land on another host or path; relative references, site
	// rules and the image Referer must see where the page actually came from,
	// not where the chain started.
	finalURL = parsedURL
	if r := resp.RawResponse; r != nil && r.Request != nil && r.Request.URL != nil {
		finalURL = r.Request.URL
	}

	return resp.Bytes(), resp.Header().Get("Content-Type"), finalURL, nil
}

// fetchArticle retrieves the page reader mode will parse, preferring a known
// full-text mirror of the URL when one exists. The returned URL is the one
// actually fetched — redirects followed — so relative references resolve
// against the right base.
func fetchArticle(ctx context.Context, url string, parsedURL *nurl.URL) ([]byte, string, *nurl.URL, error) {
	if body, contentType, mirror := fetchFullText(ctx, parsedURL); mirror != nil {
		return body, contentType, mirror, nil
	}

	if ctx.Err() != nil {
		return nil, "", nil, ctx.Err()
	}

	return fetchPage(ctx, url, parsedURL)
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

	body, contentType, finalURL, err := fetchPage(ctx, fullText, fullTextParsed)
	if err != nil {
		return nil, "", nil
	}

	return body, contentType, finalURL
}

// documentBaseURL returns the URL the page's relative references resolve
// against: the first <base href> when the document declares one — browsers
// honor it, so such pages write asset paths against it rather than against
// their own location — and pageURL otherwise. Site rules and the image
// Referer are about where the page is, not how it resolves, and keep using
// the page URL.
func documentBaseURL(body []byte, pageURL *nurl.URL) *nurl.URL {
	z := html.NewTokenizer(bytes.NewReader(body))

	for {
		tokenType := z.Next()
		if tokenType == html.ErrorToken {
			return pageURL
		}

		if tokenType != html.StartTagToken && tokenType != html.SelfClosingTagToken {
			continue
		}

		name, hasAttr := z.TagName()
		if !bytes.Equal(name, []byte("base")) {
			continue
		}

		for hasAttr {
			var key, val []byte

			key, val, hasAttr = z.TagAttr()

			if string(key) != "href" {
				continue
			}

			ref, err := nurl.Parse(strings.TrimSpace(string(val)))
			if err != nil {
				return pageURL
			}

			if base := pageURL.ResolveReference(ref); base.Scheme == "http" || base.Scheme == "https" {
				return base
			}

			return pageURL
		}
	}
}

func extractReadable(body []byte, base *nurl.URL) (*html.Node, string, error) {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, "", fmt.Errorf("could not parse page from %s: %w", base.Host, err)
	}

	// MediaWiki markup needs normalizing before readability runs, while the
	// class names that identify it are still present; ARIA-described graphics
	// likewise, before readability deletes their text-less subtrees; code
	// languages likewise, before the classes declaring them are stripped;
	// pre internals likewise, before their spans and links read as chrome;
	// LaTeXML tables likewise, before a class marker reads as page chrome;
	// unconverted LaTeXML pictures likewise, before their class is stripped
	// and the raw source inside reads as prose; BBC's data-block sections
	// likewise, before their wrapper chains make readability score each
	// section alone and keep only the first.
	normalizeDataBlocks(doc)
	normalizeMediaWiki(doc)
	normalizeRoleImages(doc)
	preserveCodeLang(doc)
	preservePreContent(doc)
	normalizeLatexmlTables(doc)
	dropLatexmlRawPictures(doc)

	parser := readability.NewParser()

	parser.ClassesToPreserve = append(parser.ClassesToPreserve, latexmlPreservedClasses...)
	parser.ClassesToPreserve = append(parser.ClassesToPreserve, infoboxPreservedClasses...)

	a, err := parser.ParseAndMutate(doc, base)
	if err != nil {
		return nil, "", fmt.Errorf("could not parse article from %s: %w", base.Host, err)
	}

	if a.Node == nil {
		return nil, "", fmt.Errorf("could not extract readable content from %s", base.Host)
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
