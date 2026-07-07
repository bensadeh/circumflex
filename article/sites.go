package article

import (
	"regexp"
	"slices"
	"strings"
)

// siteRules is the per-domain cleanup vocabulary. Rules operate on the block
// representation before any styling, so matches are against plain text.
type siteRules struct {
	domains []string

	// stopAtHeading truncates the article at a heading with this exact text.
	stopAtHeading []string

	// stopAtBlockContaining truncates the article at the first block
	// containing this text.
	stopAtBlockContaining []string

	// dropBlockEquals removes blocks whose entire text equals this string.
	dropBlockEquals []string

	// dropBlockContaining removes blocks containing this text.
	dropBlockContaining []string

	// dropInline deletes matches from the text of every span.
	dropInline []*regexp.Regexp
}

var reWikipediaRef = regexp.MustCompile(`\[(\d+|edit)\]`)

var allSiteRules = []siteRules{
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
		domains:             []string{"tomshardware.com"},
		dropBlockContaining: []string{"(Image credit: "},
	},
	{
		domains:             []string{"cnn.com"},
		dropBlockContaining: []string{"Credit: "},
	},
	{
		domains:             []string{"arstechnica.com"},
		dropBlockContaining: []string{"This story originally appeared on "},
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

		if b.kind == blockHeading && slices.Contains(rules.stopAtHeading, text) {
			break
		}

		if containsAny(text, rules.stopAtBlockContaining) {
			break
		}

		if slices.Contains(rules.dropBlockEquals, strings.TrimSpace(text)) {
			continue
		}

		if containsAny(text, rules.dropBlockContaining) {
			continue
		}

		out = append(out, b)
	}

	return out
}

func rulesForHost(hostname string) (siteRules, bool) {
	for _, rules := range allSiteRules {
		for _, domain := range rules.domains {
			if hostname == domain || strings.HasSuffix(hostname, "."+domain) {
				return rules, true
			}
		}
	}

	return siteRules{}, false
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
