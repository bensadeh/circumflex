package timeago

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRelativeTime(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name     string
		offset   time.Duration
		expected string
	}{
		{"just now", 10 * time.Second, "a few seconds ago"},
		{"boundary below 45s", 44 * time.Second, "a few seconds ago"},
		{"a minute ago", 50 * time.Second, "a minute ago"},
		{"boundary below 90s", 89 * time.Second, "a minute ago"},
		{"just past the minutes threshold rounds up", 100 * time.Second, "2 minutes ago"},
		{"2 minutes ago", 2 * time.Minute, "2 minutes ago"},
		{"15 minutes ago", 15 * time.Minute, "15 minutes ago"},
		{"44 minutes ago", 44 * time.Minute, "44 minutes ago"},
		{"an hour ago", 50 * time.Minute, "an hour ago"},
		{"boundary below 90m", 89 * time.Minute, "an hour ago"},
		{"just past the hours threshold rounds up", 91 * time.Minute, "2 hours ago"},
		{"2 hours ago", 2 * time.Hour, "2 hours ago"},
		{"12 hours ago", 12 * time.Hour, "12 hours ago"},
		{"21 hours ago", 21 * time.Hour, "21 hours ago"},
		{"a day ago", 30 * time.Hour, "a day ago"},
		{"just past the days threshold rounds up", 37 * time.Hour, "2 days ago"},
		{"2 days ago", 2 * 24 * time.Hour, "2 days ago"},
		{"25 days ago", 25 * 24 * time.Hour, "25 days ago"},
		{"a month ago", 40 * 24 * time.Hour, "a month ago"},
		{"just past the months threshold rounds up", 46 * 24 * time.Hour, "2 months ago"},
		{"2 months ago", 60 * 24 * time.Hour, "2 months ago"},
		{"11 months ago", 340 * 24 * time.Hour, "11 months ago"},
		{"a year ago", 400 * 24 * time.Hour, "a year ago"},
		{"boundary below 548d", 547 * 24 * time.Hour, "a year ago"},
		{"just past the years threshold rounds up", 548 * 24 * time.Hour, "2 years ago"},
		{"2 years ago", 730 * 24 * time.Hour, "2 years ago"},
		{"5 years ago", 5 * 365 * 24 * time.Hour, "5 years ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			unixTime := now.Add(-tt.offset).Unix()
			assert.Equal(t, tt.expected, RelativeTime(unixTime))
		})
	}
}
