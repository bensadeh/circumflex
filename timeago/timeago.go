package timeago

import (
	"fmt"
	"time"
)

// RelativeTime converts a Unix timestamp to a human-readable relative time
// string like "2 hours ago" or "a day ago".
func RelativeTime(unixTime int64) string {
	d := time.Since(time.Unix(unixTime, 0))
	seconds := int(d.Seconds())

	switch {
	case seconds < 45:
		return "a few seconds ago"
	case seconds < 90:
		return "a minute ago"
	case seconds < 45*60:
		return fmt.Sprintf("%d minutes ago", int(d.Minutes()))
	case seconds < 90*60:
		return "an hour ago"
	case seconds < 22*3600:
		return fmt.Sprintf("%d hours ago", int(d.Hours()))
	case seconds < 36*3600:
		return "a day ago"
	case seconds < 26*24*3600:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	case seconds < 45*24*3600:
		return "a month ago"
	case seconds < 345*24*3600:
		return fmt.Sprintf("%d months ago", int(d.Hours()/(24*30)))
	case seconds < 545*24*3600:
		return "a year ago"
	default:
		return fmt.Sprintf("%d years ago", int(d.Hours()/(24*365)))
	}
}
