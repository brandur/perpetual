package updater

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var intervalFormat = "LHI%03d: %s"
var intervalPattern = regexp.MustCompile(`^LHI(\d{3}): `)

// Update iterates through an account's tweets as far back as necessary to
// discover the last posted interval, then decides whether or not to post a new
// interval based off of the next interval's target time.
//
// now is injected as a parameter for better testability. It's safe to pass
// this as time.Now in most cases.
//
// Update returns an integer representing the ID of the interval that was posted if
// one was posted, or -1 otherwise. An error is returned if there was a problem
// communicating with Twitter's API.
func Update(api TwitterAPI, intervals []*Interval, now time.Time) (int, error) {
	var id int
	var ok bool

	it := api.ListTweets()

	fmt.Printf("Iterating backward through tweets\n")

	// Keep in mind that we expect our API to return tweets in reverse order
	// (i.e., newest first). Many assumptions are built into this code to take
	// advantage of that.
	for it.Next() {
		tweet := it.Value()

		id, ok = extractIntervalID(tweet.Message)
		if ok {
			fmt.Printf("Found interval ID: %v\n", id)
			break
		}
	}

	if it.Err() != nil {
		return -1, it.Err()
	}

	var nextIntervalID int
	if ok {
		// Pick the next in the series
		nextIntervalID = id + 1
	} else {
		// If ok is false, we never extracted an interval ID, which means that this
		// program has never posted before. Pick the first interval ID in the
		// series.
		nextIntervalID = 0
	}

	fmt.Printf("Next interval ID: %v\n", nextIntervalID)

	if nextIntervalID >= len(intervals) {
		fmt.Printf("There is no next interval; this program is done\n")
		return -1, nil
	}

	// Check if the Interval is ready to be posted
	interval := intervals[nextIntervalID]

	if interval.Target.After(now) {
		fmt.Printf("Interval not ready, target: %v\n", interval.Target)
		return -1, nil
	}

	tweet, err := api.PostTweet(formatInterval(nextIntervalID, interval.Message))
	if err != nil {
		return -1, err
	}

	fmt.Printf("Posted tweet: %+v\n", tweet)
	return nextIntervalID, nil
}

func extractIntervalID(content string) (int, bool) {
	matches := intervalPattern.FindAllStringSubmatch(content, 1)

	if len(matches) == 0 {
		return -1, false
	}

	// See: https://godoc.org/regexp#Regexp.FindAllStringSubmatch
	idString := matches[0][1]

	id, err := strconv.Atoi(idString)

	// We know we already matched on a number, and don't expect any kind of
	// malicious input, so we never expect this not to work.
	if err != nil {
		panic(err)
	}

	return id, true
}

func formatInterval(id int, message string) string {
	return fmt.Sprintf(intervalFormat, id, message)
}
