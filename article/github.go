package article

import (
	"bytes"
	"encoding/json"
	nurl "net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bensadeh/circumflex/ansi"

	"golang.org/x/net/html"
)

// GitHub serves issues as a React app: the server-rendered HTML carries only
// a fragment of the issue body, so readability finds no author and no
// comments. The full conversation ships inside the page anyway, as the app's
// embedded JSON payload, and issue pages parse from that instead — the issue
// body, then each comment, each boxed under its author. Long threads embed
// only the first window of timeline items; a trailing marker links to the
// rest. A page without a readable payload falls back to ordinary extraction,
// so the worst case stays what readability finds.

var reGitHubIssuePath = regexp.MustCompile(`^/([^/]+)/([^/]+)/issues/(\d+)/?$`)

func isGitHubIssue(u *nurl.URL) bool {
	return domainMatches(u.Hostname(), "github.com") && reGitHubIssuePath.MatchString(u.Path)
}

// issueOwner reads the repo owner off the URL — the page needs no asking.
func issueOwner(u *nurl.URL) string {
	m := reGitHubIssuePath.FindStringSubmatch(u.Path)
	if m == nil {
		return ""
	}

	return m[1]
}

type ghActor struct {
	Login string `json:"login"`
}

type ghTimelineItem struct {
	Typename  string  `json:"__typename"`
	Author    ghActor `json:"author"`
	Body      string  `json:"body"`
	CreatedAt string  `json:"createdAt"`
}

type ghIssue struct {
	Title              string  `json:"title"`
	Body               string  `json:"body"`
	State              string  `json:"state"`
	CreatedAt          string  `json:"createdAt"`
	Author             ghActor `json:"author"`
	FrontTimelineItems struct {
		TotalCount int `json:"totalCount"`
		Edges      []struct {
			Node ghTimelineItem `json:"node"`
		} `json:"edges"`
	} `json:"frontTimelineItems"`
}

type ghEmbeddedData struct {
	Payload struct {
		PreloadedQueries []struct {
			Result struct {
				Data struct {
					Repository struct {
						Issue *ghIssue `json:"issue"`
					} `json:"repository"`
				} `json:"data"`
			} `json:"result"`
		} `json:"preloadedQueries"`
	} `json:"payload"`
}

// parseGitHubIssueBlocks returns nil blocks when the page carries no readable
// payload — a moved page, a login wall, a redesign — and the caller falls
// back to readability extraction.
func parseGitHubIssueBlocks(body []byte, base *nurl.URL) ([]block, string) {
	raw := embeddedAppJSON(body)
	if raw == nil {
		return nil, ""
	}

	var embedded ghEmbeddedData
	if err := json.Unmarshal(raw, &embedded); err != nil {
		return nil, ""
	}

	var issue *ghIssue

	for _, query := range embedded.Payload.PreloadedQueries {
		if i := query.Result.Data.Repository.Issue; i != nil && i.Title != "" {
			issue = i

			break
		}
	}

	if issue == nil {
		return nil, ""
	}

	// The repo owner in the URL is the whole maintainer story: an org's
	// members go untagged, and that is the accepted price of asking nobody.
	owner := issueOwner(base)
	maintains := func(a ghActor) bool {
		return a.Login != "" && strings.EqualFold(a.Login, owner)
	}

	// The state rides the thread's own box; the repo and issue number stay
	// off the page — the meta bar's URL row already spells them out.
	first := issueComment(issue.Author, issue.CreatedAt, issue.Body, true, maintains(issue.Author), base)
	first.state = issueStateLabel(issue.State)

	blocks := []block{first}

	for _, edge := range issue.FrontTimelineItems.Edges {
		node := edge.Node
		if node.Typename != "IssueComment" || strings.TrimSpace(node.Body) == "" {
			continue
		}

		blocks = append(blocks, issueComment(node.Author, node.CreatedAt, node.Body, node.Author.Login == issue.Author.Login, maintains(node.Author), base))
	}

	if remaining := issue.FrontTimelineItems.TotalCount - len(issue.FrontTimelineItems.Edges); remaining > 0 {
		blocks = append(blocks, block{
			kind:  blockMore,
			spans: []span{{text: strconv.Itoa(remaining) + " more on GitHub", href: base.String()}},
		})
	}

	return blocks, issue.Title
}

func issueComment(author ghActor, createdAt, body string, op, maintainer bool, base *nurl.URL) block {
	children, err := markdownToBlocks([]byte(body), base)
	if err != nil {
		children = parseTextBlocks(body)
	}

	if len(children) == 0 {
		children = []block{{kind: blockParagraph, spans: []span{{text: "No description provided.", format: formatItalic}}}}
	}

	// A deleted account renders as the ghost user, the way GitHub itself
	// shows one.
	login := ansi.Field(author.Login)
	if login == "" {
		login = "ghost"
	}

	return block{
		kind:       blockComment,
		author:     login,
		when:       issueDate(createdAt),
		op:         op,
		maintainer: maintainer,
		children:   children,
	}
}

// issueDate is absolute where the comment views show relative time: issues
// reach the reader years after the fact, and "Apr 2016" places a thread the
// way "10 years ago" cannot.
func issueDate(createdAt string) string {
	t, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return ""
	}

	return t.Format("Jan 2, 2006")
}

func issueStateLabel(state string) string {
	if state == "" {
		return ""
	}

	return strings.ToLower(state)
}

// embeddedAppJSON returns the react-app's embedded data payload — raw JSON,
// which GitHub escapes inside its own strings rather than as HTML entities —
// or nil when the page carries none.
func embeddedAppJSON(body []byte) []byte {
	z := html.NewTokenizer(bytes.NewReader(body))
	inPayload := false

	for {
		tokenType := z.Next()

		if tokenType == html.ErrorToken {
			return nil
		}

		if tokenType == html.TextToken && inPayload {
			return z.Text()
		}

		if tokenType != html.StartTagToken {
			inPayload = false

			continue
		}

		name, hasAttr := z.TagName()
		inPayload = false

		if !bytes.Equal(name, []byte("script")) {
			continue
		}

		for hasAttr {
			var key, val []byte

			key, val, hasAttr = z.TagAttr()

			if string(key) == "data-target" && string(val) == "react-app.embeddedData" {
				inPayload = true
			}
		}
	}
}
