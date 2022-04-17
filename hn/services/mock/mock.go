package mock

import (
	"clx/item"
	"time"
)

type Service struct{}

func (s Service) Init(_ int) {
}

func (s Service) FetchStories(_ int, _ int) []*item.Item {
	return []*item.Item{
		{
			Title:         "Lorem ipsum dolor sit amet",
			Points:        31,
			ID:            31060183,
			User:          "alfa",
			Time:          time.Now().Add(-time.Minute * 2).Unix(),
			Domain:        "stackoverflow.com",
			CommentsCount: 61,
		},
		{
			Title:         "Aliquam mauris massa, rhoncus nec lectus eu",
			Points:        17,
			ID:            31060183,
			User:          "beta",
			Time:          time.Now().Add(-time.Minute * 46).Unix(),
			Domain:        "pkg.go.dev",
			CommentsCount: 27,
		},
		{
			Title:         "Show HN: consectetur adipiscing elit",
			Points:        331,
			User:          "gamma",
			Time:          time.Now().Add(-time.Hour * 3).Unix(),
			Domain:        "github.com",
			CommentsCount: 145,
		},
		{
			Title:         "Interdum et malesuada fames ac ante [PDF]",
			Points:        40,
			User:          "delta",
			Time:          time.Now().Add(-time.Hour * 5).Unix(),
			Domain:        "objective-see.com",
			CommentsCount: 58,
		},
		{
			Title:         "Donec sed orci aliquam lorem mattis consequat",
			Points:        417,
			User:          "epsilon",
			Time:          time.Now().Add(-time.Minute * 68).Unix(),
			Domain:        "bloomberg.com",
			CommentsCount: 41,
		},
		{
			Title:         "Ask HN: Vestibulum ante ipsum primis in faucibus?",
			Points:        40,
			User:          "zeta",
			Time:          time.Now().Add(-time.Minute * 27).Unix(),
			Domain:        "",
			CommentsCount: 77,
		},
		{
			Title:         "Orci luctus et ultrices posuere cubilia curae",
			Points:        37,
			User:          "eta",
			Time:          time.Now().Add(-time.Minute * 32).Unix(),
			Domain:        "apple.com",
			CommentsCount: 89,
		},
		{
			Title:         "Tell HN: Donec eget sapien viverra, viverra lacus in",
			Points:        16,
			User:          "theta",
			Time:          time.Now().Add(-time.Hour * 3).Unix(),
			Domain:        "en.wikipedia.org",
			CommentsCount: 91,
		},
		{
			Title:         "Vivamus rhoncus sit amet tellus eget blandit",
			Points:        7,
			User:          "iota",
			Time:          time.Now().Add(-time.Hour * 16).Unix(),
			Domain:        "xbox.com",
			CommentsCount: 98,
		},
		{
			Title:         "Fusce venenatis laoreet interdum (2003)",
			Points:        503,
			User:          "kappa",
			Time:          time.Now().Add(-time.Hour * 15).Unix(),
			Domain:        "huffingtonpost.com",
			CommentsCount: 125,
		},
		{
			Title:         "Etiam tincidunt, ex fermentum iaculis placerat",
			Points:        48,
			User:          "lambda",
			Time:          time.Now().Add(-time.Hour * 8).Unix(),
			Domain:        "bbc.co.uk",
			CommentsCount: 94,
		},
		{
			Title:         "Aenean sit amet erat eu quam vehicula fringilla",
			Points:        110,
			User:          "mu",
			Time:          time.Now().Add(-time.Minute * 400).Unix(),
			Domain:        "sites.google.com",
			CommentsCount: 14,
		},
		{
			Title:         "Launch HN: Cras consectetur (YC W05) - Nam vitae massa leo",
			Points:        9,
			User:          "nu",
			Time:          time.Now().Add(-time.Minute * 120).Unix(),
			Domain:        "",
			CommentsCount: 103,
		},
		{
			Title:         "Sed ut perspiciatis, unde omnis iste natus error",
			Points:        0,
			User:          "xi",
			Time:          time.Now().Add(-time.Minute * 17).Unix(),
			Domain:        "wired.com",
			CommentsCount: 75,
		},
		{
			Title:         "Nemo enim ipsam voluptatem, quia voluptas sit",
			Points:        66,
			User:          "omicron",
			Time:          time.Now().Add(-time.Hour).Unix(),
			Domain:        "macrumors.com",
			CommentsCount: 148,
		},
		{
			Title:         "Aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos",
			Points:        65,
			User:          "pi",
			Time:          time.Now().Add(-time.Hour).Unix(),
			Domain:        "arxiv.org",
			CommentsCount: 135,
		},
		{
			Title:         "Ut enim ad minima veniam",
			Points:        27,
			User:          "rho",
			Time:          time.Now().Add(-time.Hour).Unix(),
			Domain:        "twitter.com",
			CommentsCount: 81,
		},
		{
			Title:         "Nam libero tempore, cum soluta nobis est eligendi optio",
			Points:        84,
			User:          "sigma",
			Time:          time.Now().Add(-time.Hour).Unix(),
			Domain:        "wsj.com",
			CommentsCount: 90,
		},
		{
			Title:         "Cumque nihil impedit, quo minus id, quod maxime placeat",
			Points:        6,
			User:          "tau",
			Time:          time.Now().Add(-time.Hour).Unix(),
			Domain:        "medium.com",
			CommentsCount: 18,
		},
		{
			Title:         "Emporibus autem quibusdam et aut officiis debitis aut rerum",
			Points:        150,
			User:          "upsilon",
			Time:          time.Now().Add(-time.Hour).Unix(),
			Domain:        "npr.com",
			CommentsCount: 114,
		},
		{
			Title:         "Necessitatibus saepe eveniet, ut et voluptates repudiandae",
			Points:        135,
			User:          "phi",
			Time:          time.Now().Add(-time.Hour).Unix(),
			Domain:        "nature.com",
			CommentsCount: 118,
		},
		{
			Title:         "Quis autem vel eum iure reprehenderit",
			Points:        97,
			User:          "chi",
			Time:          time.Now().Add(-time.Hour).Unix(),
			Domain:        "venturebeat.com",
			CommentsCount: 66,
		},
		{
			Title:         "Perferendis doloribus asperiores repellat",
			Points:        50,
			User:          "psi",
			Time:          time.Now().Add(-time.Hour).Unix(),
			Domain:        "techcrunch.com",
			CommentsCount: 107,
		},
		{
			Title:         "Obcaecati cupiditate non provident,",
			Points:        68,
			User:          "omega",
			Time:          time.Now().Add(-time.Hour).Unix(),
			Domain:        "newyorker.com",
			CommentsCount: 109,
		},
		{
			Title:         "Quae ab illo inventore veritatis et quasi architecto",
			Points:        115,
			User:          "alfa",
			Time:          time.Now().Add(-time.Hour).Unix(),
			Domain:        "nytimes.com",
			CommentsCount: 62,
		},
		{
			Title:         "Qui ratione voluptatem sequi nesciunt",
			Points:        102,
			User:          "beta",
			Time:          time.Now().Add(-time.Hour).Unix(),
			Domain:        "archive.org",
			CommentsCount: 34,
		},
		{
			Title:         "Nisi ut aliquid ex ea commodi consequatur?",
			Points:        74,
			User:          "gamma",
			Time:          time.Now().Add(-time.Hour).Unix(),
			Domain:        "sciencedirect.com",
			CommentsCount: 139,
		},
		{
			Title:         "Temporibus autem quibusdam et aut officiis debitis",
			Points:        47,
			User:          "delta",
			Time:          time.Now().Add(-time.Hour).Unix(),
			Domain:        "smithsonianmag.com",
			CommentsCount: 143,
		},
		{
			Title:         "Id est laborum et dolorum fuga",
			Points:        90,
			User:          "epsilon",
			Time:          time.Now().Add(-time.Hour).Unix(),
			Domain:        "vice.com",
			CommentsCount: 48,
		},
		{
			Title:         "Omnis voluptas assumenda est, omnis dolor repellendus",
			Points:        22,
			User:          "zeta",
			Time:          time.Now().Add(-time.Hour).Unix(),
			Domain:        "abc.com",
			CommentsCount: 122,
		},
	}
}

func (s Service) FetchStory(_ int) *item.Item {
	return &item.Item{
		ID:      32145667,
		Title:   "Mauris commodo odio quis diam fermentum, et suscipit augue pharetra",
		Points:  543,
		User:    "riemann",
		TimeAgo: "6 hours ago",
		URL:     "https://en.wikipedia.org/wiki/Riemann_hypothesis",
		Level:   0,
		Domain:  "google.com",
		Comments: []*item.Item{
			{
				ID:      28,
				User:    "euler",
				TimeAgo: "1 hour ago",
				Level:   0,
				Comments: []*item.Item{
					{
						ID:      23,
						User:    "hilbert",
						TimeAgo: "40 minutes ago",
						Level:   1,
						Comments: []*item.Item{
							{
								ID:      33,
								User:    "riemann",
								TimeAgo: "27 minutes ago",
								Level:   2,
								Comments: []*item.Item{
									{
										ID:      33,
										User:    "gauss",
										TimeAgo: "26 minutes ago",
										Level:   3,
										Comments: []*item.Item{
											{
												ID:      33,
												Time:    time.Now().Add(time.Minute).Unix(),
												User:    "cantor",
												TimeAgo: "10 minutes ago",
												Level:   4,
												Comments: []*item.Item{
													{
														ID:      33,
														Time:    time.Now().Add(time.Minute).Unix(),
														User:    "euler",
														TimeAgo: "4 minutes ago",
														Level:   5,
														Content: "Duis accumsan eros sit amet libero facilisis, id placerat tellus auctor.",
													},
												},
												Content: "Curabitur cursus @hilbert in feugiat varius. Donec sit amet " +
													"erat tincidunt, mollis ex vehicula, cursus purus.",
											},
										},
										Content: "<p> > Donec quam tortor <p>Aliquam iaculis, quam ut venenatis gravida, felis risus " +
											"tristique erat, consectetur sodales quam sapien ac neque.",
									},
								},
								Content: "Phasellus ut nulla risus. Ut sed volutpat dui. Donec quam tortor, " +
									"porttitor a ante sed, finibus feugiat risus.",
							},
						},
						Content: "Sed quis lectus quam. Donec `ls -ltr` vitae lorem porttitor, vel dignissim dolor interdum.",
					},
					{
						ID:      33,
						User:    "ramanujan",
						TimeAgo: "27 minutes ago",
						Level:   1,
						Content: "Temporibus autem quibusdam et aut officiis debitis aut rerum necessitatibus saepe " +
							"eveniet, ut et voluptates repudiandae.",
					},
				},
				Content: "Maecenas suscipit aliquet lorem, at semper orci elementum vitae. Ut sit amet ante " +
					"venenatis, molestie sem quis, sagittis felis.",
			},
			{
				ID:       1,
				User:     "wolfgang",
				TimeAgo:  "6 minutes ago",
				Level:    0,
				Comments: nil,
				Content: "<p>Boss keys[1] should be standard for every website.  The March Madness[2] one for example." +
					"  There are some good ones developed for Reddit to look like Outlook[3], shell [4], etc.<p>[1]" +
					" <a href=\"https:&#x2F;&#x2F;en.wikipedia.org&#x2F;wiki&#x2F;Boss_key\" rel=\"nofollow\">" +
					"https:&#x2F;&#x2F;en.wikipedia.org&#x2F;wiki&#x2F;Boss_key</a><p>[2] <a href=\"https:&#x2F;&#x2F;www." +
					"ncaa.com&#x2F;march-madness-live&#x2F;boss\" rel=\"nofollow\">https:&#x2F;&#x2F;www.ncaa.com&#x2F;" +
					"march-madness-live&#x2F;boss</a><p>[3] <a href=\"http:&#x2F;&#x2F;pcottle.github.io&#x2F;MSOutlookit&#" +
					"x2F;&#x2F;\" rel=\"nofollow\">http:&#x2F;&#x2F;pcottle.github.io&#x2F;MSOutlookit&#x2F;&#x2F;</a><p>[4]" +
					" <a href=\"https:&#x2F;&#x2F;redditshell.com&#x2F;\" rel=\"nofollow\">https:&#x2F;&#x2F;redditshell.com&#x2F;</a>",
				CommentsCount: 0,
			},
		},
		Content: "<p>Lorem ipsum dolor sit amet, " +
			"consectetur adipiscing elit. Integer a augue id elit efficitur tempor sit amet quis lectus.",
		CommentsCount: 57,
	}
}
