// Package website reads the Hacker News pages that the Firebase API does not
// expose. Active Threads is one of them: it ranks the stories with the
// liveliest discussions right now and exists only as a rendered page.
//
// Only the item IDs are read off the page. Firebase still serves the stories
// behind them, so a scraped category arrives through the same pipeline — and
// with the same fields, sanitizing and dead-item filtering — as every other
// one.
package website

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/bensadeh/circumflex/version"

	"golang.org/x/net/html"
	"resty.dev/v3"
)

const (
	defaultBaseURL = "https://news.ycombinator.com"
	activePath     = "/active"
	httpTimeout    = 10 * time.Second
	retryCount     = 3
	retryWaitTime  = 200 * time.Millisecond
	retryMaxWait   = 2 * time.Second
	// The page is one screen of 30 submissions (~35 KB today); the cap leaves
	// generous headroom while still bounding a pathologically large response.
	responseLimit = 4 << 20
)

// errNoStories keeps a markup change from looking like an empty category:
// the page always lists submissions, so parsing none means the rows moved
// rather than that Hacker News went quiet.
var errNoStories = errors.New("found no stories on the active threads page")

// discardLogger silences resty's internal logging so that WARN/ERROR
// messages on context cancellation don't corrupt the TUI.
type discardLogger struct{}

func (discardLogger) Errorf(string, ...any) {}
func (discardLogger) Warnf(string, ...any)  {}
func (discardLogger) Debugf(string, ...any) {}

type Service struct {
	client  *resty.Client
	baseURL string
}

func NewService() *Service {
	client := resty.New()
	client.SetTimeout(httpTimeout)
	client.SetRedirectPolicy(resty.RedirectNoPolicy())
	client.SetHeader("User-Agent", version.Name+"/"+version.Version)
	client.SetRetryCount(retryCount)
	client.SetRetryWaitTime(retryWaitTime)
	client.SetRetryMaxWaitTime(retryMaxWait)
	client.SetResponseBodyLimit(responseLimit)
	client.AddRetryConditions(func(resp *resty.Response, _ error) bool {
		return resp != nil && resp.StatusCode() >= http.StatusInternalServerError
	})
	client.SetLogger(discardLogger{})

	return &Service{client: client, baseURL: defaultBaseURL}
}

// FetchActiveStoryIDs returns the IDs of the stories listed on the Active
// Threads page, in the order the page ranks them.
func (s *Service) FetchActiveStoryIDs(ctx context.Context) ([]int, error) {
	resp, err := s.client.R().SetContext(ctx).Get(s.baseURL + activePath)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("could not fetch active threads, server returned status %d %s",
			resp.StatusCode(), http.StatusText(resp.StatusCode()))
	}

	ids, err := parseStoryIDs(bytes.NewReader(resp.Bytes()))
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return nil, errNoStories
	}

	return ids, nil
}

// parseStoryIDs pulls the item IDs out of the submission rows. A story on a
// Hacker News listing opens with
//
//	<tr class="athing submission" id="49717558">
//
// followed by the rows carrying its subtext and the spacer below it, neither
// of which is an athing — so the class is what separates the stories from the
// rest of the table.
func parseStoryIDs(body io.Reader) ([]int, error) {
	tokenizer := html.NewTokenizer(body)

	var ids []int

	for {
		if tokenizer.Next() == html.ErrorToken {
			if errors.Is(tokenizer.Err(), io.EOF) {
				return ids, nil
			}

			return nil, fmt.Errorf("reading the active threads page: %w", tokenizer.Err())
		}

		name, hasAttributes := tokenizer.TagName()
		if !hasAttributes || string(name) != "tr" {
			continue
		}

		if id, ok := submissionID(tokenizer); ok {
			ids = append(ids, id)
		}
	}
}

// submissionID returns the ID of the story the current <tr> opens, or false
// if the row is not a submission. Attributes are read to the end either way:
// a half-read tag leaves the tokenizer out of step with the next one.
func submissionID(tokenizer *html.Tokenizer) (int, bool) {
	var (
		id           int
		isSubmission bool
		hasID        bool
	)

	for {
		key, value, moreAttributes := tokenizer.TagAttr()

		switch string(key) {
		case "class":
			isSubmission = slices.Contains(strings.Fields(string(value)), "athing")
		case "id":
			parsed, err := strconv.Atoi(string(value))
			id, hasID = parsed, err == nil && parsed > 0
		}

		if !moreAttributes {
			break
		}
	}

	return id, isSubmission && hasID
}
