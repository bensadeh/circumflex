package article

import (
	"regexp"
	"slices"
	"strings"
)

// Rules match against unstyled block text, before any rendering. All entries
// whose domain matches the page's host are merged, so shared fragments and
// site-specific additions compose. stopAt* truncates the article from the
// first match on; dropBlock* removes single blocks; dropInline edits span text.
type siteRules struct {
	domains []string

	stopAtHeading         []string
	stopAtBlockEquals     []string
	stopAtBlockContaining []string

	dropBlockEquals     []string
	dropBlockContaining []string
	dropBlockMatching   []*regexp.Regexp

	dropInline []*regexp.Regexp
}

// domainMatches reports whether hostname is domain itself or a subdomain of
// it. The site rules and the domain blocklist share this one definition.
func domainMatches(hostname, domain string) bool {
	return hostname == domain || strings.HasSuffix(hostname, "."+domain)
}

func (rs siteRules) matches(hostname string) bool {
	for _, domain := range rs.domains {
		if domainMatches(hostname, domain) {
			return true
		}
	}

	return false
}

func (rs siteRules) merge(other siteRules) siteRules {
	rs.stopAtHeading = append(rs.stopAtHeading, other.stopAtHeading...)
	rs.stopAtBlockEquals = append(rs.stopAtBlockEquals, other.stopAtBlockEquals...)
	rs.stopAtBlockContaining = append(rs.stopAtBlockContaining, other.stopAtBlockContaining...)
	rs.dropBlockEquals = append(rs.dropBlockEquals, other.dropBlockEquals...)
	rs.dropBlockContaining = append(rs.dropBlockContaining, other.dropBlockContaining...)
	rs.dropBlockMatching = append(rs.dropBlockMatching, other.dropBlockMatching...)
	rs.dropInline = append(rs.dropInline, other.dropInline...)

	return rs
}

var (
	reWikipediaRef  = regexp.MustCompile(`\[(\d+|edit)\]`)
	reCommentsCount = regexp.MustCompile(`^\d+ Comments$`)
	reArxivFormats  = regexp.MustCompile(`^View PDF`)
)

var allSiteRules = []siteRules{
	{
		// These only appear on the abstract page, the fallback for papers
		// without a full-text HTML rendering.
		domains:           []string{"arxiv.org"},
		stopAtHeading:     []string{"Submission history"},
		dropBlockMatching: []*regexp.Regexp{reArxivFormats},
	},
	{
		domains:       []string{"wikipedia.org"},
		stopAtHeading: []string{"References", "Footnotes", "See also", "Notes", "External links"},
		dropBlockEquals: []string{
			"From Wikipedia, the free encyclopedia",
			"Official website",
		},
		dropBlockContaining: []string{
			"Archived from the original",
			"Edit this at Wikidata",
		},
		dropInline: []*regexp.Regexp{reWikipediaRef},
	},
	{
		domains: []string{"nytimes.com"},
		dropBlockContaining: []string{
			"Credit…",
			"This is a developing story. Check back for updates.",
		},
		dropBlockEquals: []string{"Credit", "Image"},
	},
	{
		domains: []string{"economist.com"},
		dropBlockContaining: []string{
			"Listen to this story",
			"Your browser does not support the ",
			"Listen on the go",
			"Get The Economist app and play articles",
			"Play in app",
			"Enjoy more audio and podcasts on iOS or Android",
		},
		stopAtBlockContaining: []string{
			"This article appeared in the",
			"For more coverage of ",
		},
	},
	{
		domains: []string{"bbc.com", "bbc.co.uk"},
		stopAtBlockEquals: []string{
			"--",
			"You may also be interested in:",
		},
		dropBlockContaining: []string{"(Image credit: "},
	},
	{
		domains:             []string{"tomshardware.com"},
		dropBlockContaining: []string{"(Image credit: "},
	},
	{
		domains:             []string{"cnn.com"},
		dropBlockContaining: []string{"Credit: "},
	},
	{
		domains: []string{"arstechnica.com"},
		dropBlockContaining: []string{
			"This story originally appeared on ",
			"Credit: ",
			"Listing image for first story",
		},
		dropBlockMatching: []*regexp.Regexp{reCommentsCount},
	},
	{
		domains:       []string{"macrumors.com"},
		stopAtHeading: []string{"Top Stories", "Related Stories"},
	},
	{
		domains: []string{"wired.com", "wired.co.uk"},
		dropBlockContaining: []string{
			"Read more: ",
			"Do you use social media regularly? Take our short survey.",
		},
		stopAtHeading: []string{"More Great WIRED Stories"},
	},
	{
		domains:             []string{"theguardian.com"},
		dropBlockContaining: []string{"Photograph:"},
	},
	{
		domains: []string{"axios.com"},
		dropBlockContaining: []string{
			"Sign up for our daily briefing",
			"Catch up on the day's biggest business stories",
			"Stay on top of the latest market trends",
			"Sports news worthy of your time",
			"Tech news worthy of your time",
			"Get the inside stories",
			"Axios on your phone",
			"Catch up on coronavirus stories and special reports",
			"Want a daily digest of the top ",
			"Get a daily digest of the most important stories ",
			"Download for free.",
			"Sign up for free.",
			"Make your busy days simpler with Axios AM/PM",
			"Subscribe to Axios Closer",
			"Get breaking news",
			"Sign up for Axios",
			"Stay up-to-date on the most important and interesting",
		},
	},
	{
		domains: []string{"9to5mac.com"},
		dropBlockContaining: []string{
			"We use income earning auto affiliate links.",
			"Check out 9to5Mac on YouTube for more Apple news:",
		},
		stopAtHeading: []string{"About the Author"},
	},
	{
		domains:             []string{"smithsonianmag.com"},
		dropBlockContaining: []string{"smithsonianmag.com"},
		stopAtHeading:       []string{"Like this article?"},
	},
	{
		domains: []string{"cnet.com"},
		dropBlockContaining: []string{
			"Read more:",
			"Stay up-to-date on the latest news",
		},
	},
}

func applySiteRules(blocks []block, hostname string) []block {
	rules, found := rulesForHost(hostname)
	if !found {
		return blocks
	}

	var out []block

	for _, b := range blocks {
		b = dropInline(b, rules.dropInline)
		text := b.plainText()
		trimmed := strings.TrimSpace(text)

		if b.kind == blockHeading && slices.Contains(rules.stopAtHeading, trimmed) {
			break
		}

		if slices.Contains(rules.stopAtBlockEquals, trimmed) {
			break
		}

		if containsAny(text, rules.stopAtBlockContaining) {
			break
		}

		if slices.Contains(rules.dropBlockEquals, trimmed) {
			continue
		}

		if containsAny(text, rules.dropBlockContaining) {
			continue
		}

		if matchesAny(trimmed, rules.dropBlockMatching) {
			continue
		}

		out = append(out, b)
	}

	return out
}

func rulesForHost(hostname string) (siteRules, bool) {
	var merged siteRules

	found := false

	for _, rules := range allSiteRules {
		if rules.matches(hostname) {
			merged = merged.merge(rules)
			found = true
		}
	}

	return merged, found
}

func matchesAny(text string, patterns []*regexp.Regexp) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(text) {
			return true
		}
	}

	return false
}

func containsAny(text string, targets []string) bool {
	for _, target := range targets {
		if strings.Contains(text, target) {
			return true
		}
	}

	return false
}

func dropInline(b block, patterns []*regexp.Regexp) block {
	if len(patterns) == 0 {
		return b
	}

	// Code and verbatim blocks are content, not prose: citation strippers
	// like [\d+] would silently rewrite array indices and the like.
	if b.kind == blockCode || b.kind == blockVerbatim {
		return b
	}

	clean := func(text string) string {
		for _, pattern := range patterns {
			text = pattern.ReplaceAllString(text, "")
		}

		return text
	}

	b.spans = cleanSpans(b.spans, clean)
	b.text = clean(b.text)

	if len(b.items) > 0 {
		items := make([]listItem, len(b.items))
		for i, item := range b.items {
			item.spans = cleanSpans(item.spans, clean)
			items[i] = item
		}

		b.items = items
	}

	if len(b.rows) > 0 {
		rows := make([][]string, len(b.rows))
		for i, row := range b.rows {
			cells := make([]string, len(row))
			for j, cell := range row {
				cells[j] = clean(cell)
			}

			rows[i] = cells
		}

		b.rows = rows
	}

	return b
}

func cleanSpans(spans []span, clean func(string) string) []span {
	if len(spans) == 0 {
		return spans
	}

	out := make([]span, 0, len(spans))

	for _, s := range spans {
		s.text = clean(s.text)
		if s.text != "" {
			out = append(out, s)
		}
	}

	return out
}
