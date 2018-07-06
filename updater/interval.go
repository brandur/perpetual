package updater

import (
	"time"
)

const timeLayout = "Jan 2 15:04:00 MST 2006"

// Interval is a threshold in time that we're measuring across. This program wakes
// up and posts the message of one after it crosses the target time.
type Interval struct {
	// Message is the content to tweet for this interval.
	Message string

	// Target is the target time for the interval to be posted. This is measured
	// directly as a time instead of a duration (which would be easier for
	// testing/readability) to avoid problems with timezones, leap years, etc.
	// We let the underlying time library do the math for us.
	Target time.Time
}

// MustParseTime is similar to time.Parse but panics if value wasn't parseable.
// This is useful in our case because all input is predetermined.
func MustParseTime(value string) time.Time {
	t, err := time.Parse(timeLayout, value)
	if err != nil {
		panic(err)
	}
	return t
}
