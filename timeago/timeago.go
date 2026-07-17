package timeago

import (
	"fmt"
	"math"
	"time"
)

const (
	secondsPerMinute = 60
	secondsPerHour   = 3600
	secondsPerDay    = 24 * secondsPerHour
	hoursPerDay      = 24
	hoursPerMonth    = hoursPerDay * 30
	hoursPerYear     = hoursPerDay * 365

	fewSecondsThreshold = 45
	oneMinuteThreshold  = 90
	minutesThreshold    = 45 * secondsPerMinute
	oneHourThreshold    = 90 * secondsPerMinute
	hoursThreshold      = 22 * secondsPerHour
	oneDayThreshold     = 36 * secondsPerHour
	daysThreshold       = 26 * secondsPerDay
	oneMonthThreshold   = 45 * secondsPerDay
	monthsThreshold     = 345 * secondsPerDay
	// 548 days is the first value that rounds to 2 years; anything lower
	// would reach the plural branch while still rounding to 1.
	oneYearThreshold = 548 * secondsPerDay
)

// RelativeTime converts a Unix timestamp to a human-readable relative time
// string like "2 hours ago" or "a day ago". Plural counts round to the
// nearest unit rather than truncate: each plural branch starts halfway into
// its unit (36 hours, 45 days, ...), where truncation would produce
// "1 days ago".
func RelativeTime(unixTime int64) string {
	d := time.Since(time.Unix(unixTime, 0))
	seconds := int(d.Seconds())

	count := func(units float64) int { return int(math.Round(units)) }

	switch {
	case seconds < fewSecondsThreshold:
		return "a few seconds ago"
	case seconds < oneMinuteThreshold:
		return "a minute ago"
	case seconds < minutesThreshold:
		return fmt.Sprintf("%d minutes ago", count(d.Minutes()))
	case seconds < oneHourThreshold:
		return "an hour ago"
	case seconds < hoursThreshold:
		return fmt.Sprintf("%d hours ago", count(d.Hours()))
	case seconds < oneDayThreshold:
		return "a day ago"
	case seconds < daysThreshold:
		return fmt.Sprintf("%d days ago", count(d.Hours()/hoursPerDay))
	case seconds < oneMonthThreshold:
		return "a month ago"
	case seconds < monthsThreshold:
		return fmt.Sprintf("%d months ago", count(d.Hours()/hoursPerMonth))
	case seconds < oneYearThreshold:
		return "a year ago"
	default:
		return fmt.Sprintf("%d years ago", count(d.Hours()/hoursPerYear))
	}
}
