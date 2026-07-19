package hn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseItemURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		id   int
		ok   bool
	}{
		{"canonical", "https://news.ycombinator.com/item?id=42", 42, true},
		{"round-trips ItemURL", ItemURL(38519999), 38519999, true},
		{"http scheme", "http://news.ycombinator.com/item?id=7", 7, true},
		{"www prefix", "https://www.news.ycombinator.com/item?id=7", 7, true},
		{"extra query params", "https://news.ycombinator.com/item?id=9&p=2", 9, true},

		{"user page", "https://news.ycombinator.com/user?id=dang", 0, false},
		{"front page", "https://news.ycombinator.com/", 0, false},
		{"from page", "https://news.ycombinator.com/from?site=example.com", 0, false},
		{"missing id", "https://news.ycombinator.com/item", 0, false},
		{"non-numeric id", "https://news.ycombinator.com/item?id=dang", 0, false},
		{"negative id", "https://news.ycombinator.com/item?id=-1", 0, false},
		{"other host", "https://hn.algolia.com/item?id=42", 0, false},
		{"host suffix spoof", "https://news.ycombinator.com.evil.com/item?id=42", 0, false},
		{"no scheme", "news.ycombinator.com/item?id=42", 0, false},
		{"mailto", "mailto:dang@ycombinator.com", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, ok := ParseItemURL(tt.url)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.id, id)
		})
	}
}
