package mock

import (
	"math/rand"
	"time"

	"clx/constants/category"
	"clx/item"
)

type Service struct{}

func (Service) Init(_ int) {
}

func (Service) FetchItems(_ int, cat int) []*item.Item {
	// Uncomment to test the spinner on startup
	if cat != 0 {
		time.Sleep(time.Second * 1)
	}

	items := []*item.Item{
		{
			Title:         "Lorem ipsum dolor sit amet et quasi architecto",
			Points:        31,
			ID:            1,
			User:          "alfa",
			Time:          time.Now().Add(-time.Minute * 2).Unix(),
			Domain:        "stackoverflow.com",
			CommentsCount: 61,
		},
		{
			Title:         "Aliquam mauris massa, rhoncus nec lectus eu",
			Points:        17,
			ID:            2,
			User:          "beta",
			Time:          time.Now().Add(-time.Minute * 46).Unix(),
			Domain:        "pkg.go.dev",
			CommentsCount: 27,
		},
		{
			Title:         "Show HN: consectetur adipiscing elit doris elles",
			Points:        331,
			ID:            3,
			User:          "gamma",
			Time:          time.Now().Add(-time.Hour * 3).Unix(),
			Domain:        "github.com",
			CommentsCount: 145,
		},
		{
			Title:         "Interdum et malesuada fames ac ante merquenta aquito [PDF]",
			Points:        40,
			ID:            4,
			User:          "delta",
			Time:          time.Now().Add(-time.Hour * 5).Unix(),
			Domain:        "objective-see.com",
			CommentsCount: 58,
		},
		{
			Title:         "Donec sed orci aliquam lorem mattis consequat lapin es dom",
			Points:        417,
			ID:            5,
			User:          "epsilon",
			Time:          time.Now().Add(-time.Minute * 68).Unix(),
			Domain:        "bloomberg.com",
			CommentsCount: 0,
		},
		{
			Title:         "Ask HN: Vestibulum ante plero ipsum primis in faucibus?",
			Points:        40,
			ID:            6,
			User:          "zeta",
			Time:          time.Now().Add(-time.Minute * 27).Unix(),
			Domain:        "",
			CommentsCount: 77,
		},
		{
			Title:         "Orci luctus et ultrices posuere cubilia curae (2017)",
			Points:        37,
			ID:            7,
			User:          "eta",
			Time:          time.Now().Add(-time.Minute * 32).Unix(),
			Domain:        "apple.com",
			CommentsCount: 89,
		},
		{
			Title:         "Tell HN: Donec eget sapien viverra, viverra lacus in",
			Points:        16,
			ID:            8,
			User:          "theta",
			Time:          time.Now().Add(-time.Hour * 3).Unix(),
			Domain:        "en.wikipedia.org",
			CommentsCount: 91,
		},
		{
			Title:         "Vivamus rhoncus sit amet tellus eget blandit",
			Points:        7,
			ID:            9,
			User:          "iota",
			Time:          time.Now().Add(-time.Hour * 16).Unix(),
			Domain:        "xbox.com",
			CommentsCount: 98,
		},
		{
			Title:         "Fusce venenatis laoreet interdum (2003)",
			Points:        503,
			ID:            10,
			User:          "kappa",
			Time:          time.Now().Add(-time.Hour * 15).Unix(),
			Domain:        "huffingtonpost.com",
			CommentsCount: 125,
		},
		{
			Title:         "Etiam tincidunt, ex fermentum iaculis placerat",
			Points:        48,
			ID:            11,
			User:          "lambda",
			Time:          time.Now().Add(-time.Hour * 8).Unix(),
			Domain:        "bbc.co.uk",
			CommentsCount: 94,
		},
		{
			Title:         "Aenean sit amet erat eu quam vehicula fringilla",
			Points:        110,
			ID:            12,
			User:          "mu",
			Time:          time.Now().Add(-time.Minute * 400).Unix(),
			Domain:        "sites.google.com",
			CommentsCount: 14,
		},
		{
			Title:         "Launch HN: Cras consectetur (YC W05) - Nam vitae massa leo",
			Points:        9,
			ID:            13,
			User:          "nu",
			Time:          time.Now().Add(-time.Minute * 120).Unix(),
			Domain:        "",
			CommentsCount: 103,
		},
		{
			Title:         "Sed ut perspiciatis, unde omnis iste natus error [video]",
			Points:        26,
			ID:            14,
			User:          "xi",
			Time:          time.Now().Add(-time.Minute * 17).Unix(),
			Domain:        "wired.com",
			CommentsCount: 75,
		},
		{
			Title:         "Nemo enim ipsam voluptatem, quia voluptas sit",
			Points:        66,
			ID:            15,
			User:          "omicron",
			Time:          time.Now().Add(-time.Hour).Unix(),
			Domain:        "macrumors.com",
			CommentsCount: 148,
		},
		{
			Title:         "Aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos",
			Points:        65,
			ID:            16,
			User:          "pi",
			Time:          time.Now().Add(-time.Hour).Unix(),
			Domain:        "arxiv.org",
			CommentsCount: 135,
		},
		{
			Title:         "Ut enim ad minima veniam",
			Points:        27,
			ID:            17,
			User:          "rho",
			Time:          time.Now().Add(-time.Hour).Unix(),
			Domain:        "twitter.com",
			CommentsCount: 81,
		},
		{
			Title:         "Nam libero tempore (YC S16), cum soluta nobis est eligendi optio",
			ID:            18,
			Points:        84,
			User:          "sigma",
			Time:          time.Now().Add(-time.Hour).Unix(),
			Domain:        "wsj.com",
			CommentsCount: 90,
		},
		{
			Title:         "Cumque nihil impedit, quo minus id, quod maxime placeat",
			Points:        6,
			ID:            19,
			User:          "tau",
			Time:          time.Now().Add(-time.Hour).Unix(),
			Domain:        "medium.com",
			CommentsCount: 18,
		},
		{
			Title:         "Emporibus autem quibusdam et aut officiis debitis aut rerum",
			Points:        150,
			ID:            20,
			User:          "upsilon",
			Time:          time.Now().Add(-time.Hour).Unix(),
			Domain:        "npr.com",
			CommentsCount: 114,
		},
		{
			Title:         "Necessitatibus saepe eveniet, ut et voluptates repudiandae",
			Points:        135,
			ID:            21,
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
			Title:         "Obcaecati cupiditate non provident",
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

	// Randomize list to make debugging easier
	if cat != category.FrontPage {
		rand.Shuffle(len(items), func(i, j int) { items[i], items[j] = items[j], items[i] })
	}

	return items
}

func (Service) FetchComments(_ int) *item.Item {
	return &item.Item{
		ID:      32145667,
		Title:   "Mauris commodo odio (YC W05) quis diam fermentum, et suscipit augue pharetra [video]",
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
				Content: "<p>Vivamus elementum auctor congue. Etiam[1] nulla nisl, varius vehicula[2] quam vel, aliquet " +
					"iaculis enim. Donec felis elit[3], sollicitudin viverra velit eget, posuere[4] vestibulum eros. " +
					"<p>[1] <a href=\"https:&#x2F;&#x2F;en.wikipedia.org&#x2F;wiki&#x2F;Boss_key\" rel=\"nofollow\">" +
					"https:&#x2F;&#x2F;en.wikipedia.org&#x2F;wiki&#x2F;Boss_key</a><p>[2] <a href=\"https:&#x2F;&#x2F;www." +
					"ncaa.com&#x2F;march-madness-live&#x2F;boss\" rel=\"nofollow\">https:&#x2F;&#x2F;www.ncaa.com&#x2F;" +
					"march-madness-live&#x2F;boss</a><p>[3] <a href=\"http:&#x2F;&#x2F;pcottle.github.io&#x2F;MSOutlookit&#" +
					"x2F;&#x2F;\" rel=\"nofollow\">http:&#x2F;&#x2F;pcottle.github.io&#x2F;MSOutlookit&#x2F;&#x2F;</a><p>[4]" +
					" <a href=\"https:&#x2F;&#x2F;redditshell.com&#x2F;\" rel=\"nofollow\">https:&#x2F;&#x2F;redditshell.com&#x2F;</a>",
				CommentsCount: 0,
			},
			{
				ID:      5,
				User:    "hamilton",
				TimeAgo: "2 hours ago",
				Level:   0,
				Comments: []*item.Item{
					{
						ID:            41,
						User:          "euler",
						TimeAgo:       "1 hour ago",
						Level:         1,
						Comments:      nil,
						Content:       "<p>&gt; Godot doesn’t fight you when you’re building scenes. Making a scene feels a lot like creating a class using composition, and scenes can even inherit from other scenes (using another scene as the the root node of a scene allows you to inherit from it and override its properties in the editor and in code), allowing you to express patterns you’re intimately familiar with from object-oriented programming.<p>I personally find the approach of nodes everywhere a bit odd.<p><pre><code>  EnemyObject\n    PathfindingObject\n      PathfindingBehavior (attached script)\n    ShootingObject\n      ShootingBehavior (attached script)\n    TalkingObject\n      TalkingBehavior (attached script)\n</code></pre>\nIt kind of feels like it would be nicer to be able to attach a number of scripts to the object that I actually want to control, instead of having Nodes that I don&#x27;t really see much of a use for, apart from them being script containers.</a>",
						CommentsCount: 0,
					},
				},
				Content:       "<p>&gt; Godot doesn’t fight you when you’re building scenes. Making a scene feels a lot like creating a class using composition, and scenes can even inherit from other scenes (using another scene as the the root node of a scene allows you to inherit from it and override its properties in the editor and in code), allowing you to express patterns you’re intimately familiar with from object-oriented programming.<p>I personally find the approach of nodes everywhere a bit odd.<p><pre><code>  EnemyObject\n    PathfindingObject\n      PathfindingBehavior (attached script)\n    ShootingObject\n      ShootingBehavior (attached script)\n    TalkingObject\n      TalkingBehavior (attached script)\n</code></pre>\nIt kind of feels like it would be nicer to be able to attach a number of scripts to the object that I actually want to control, instead of having Nodes that I don&#x27;t really see much of a use for, apart from them being script containers.</a>",
				CommentsCount: 0,
			},
			{
				ID:            4,
				User:          "hamilton",
				TimeAgo:       "1 hour ago",
				Level:         0,
				Comments:      nil,
				Content:       "<p> This comment tests parsing of YC-funded companies: xxxxxxCompany (YC S07)",
				CommentsCount: 0,
			},
		},
		Content: "<p>Lorem ipsum dolor sit amet, " +
			"consectetur adipiscing elit. Integer a augue id elit efficitur tempor sit amet quis lectus.",
		CommentsCount: 57,
	}
}

func (s Service) FetchItem(id int) *item.Item {
	return nil
}
