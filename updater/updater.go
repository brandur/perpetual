package updater

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// LHI: Long Heartbeat Interval
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
	var lastTweet *Tweet
	var ok bool

	it := api.ListTweets()

	fmt.Printf("Iterating backward through tweets\n")

	// Keep in mind that we expect our API to return tweets in reverse order
	// (i.e., newest first). Many assumptions are built into this code to take
	// advantage of that.
	for it.Next() {
		lastTweet = it.Value()

		id, ok = extractIntervalID(lastTweet.Message)
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
		// A special case: the Twitter API has a fundamental limitation in that
		// it will not return every tweet. At some point when you go too far
		// back, it gives up and gives you nothing more.
		//
		// This puts us in an awkward place. It's possible that for our longer
		// intervals there's been a lot of time and many tweets posted between
		// the last interval and now. The last interval was posted, but we
		// can't see it because it's outside of the set of tweets that Twitter
		// will return. If we posted in this state, it's possible that we'd
		// accidentally just restart the whole interval sequence all over and
		// double-post everything.
		//
		// To avoid that failure case, we just error and do nothing. This does
		// put us in a reasonably likely case of failing to post far future
		// intervals because of this limitation, but there's little we can do
		// to rectify that.
		if lastTweet != nil && lastTweet.CreatedAt.After(intervals[0].Target) {
			return -1, fmt.Errorf(
				"Last available tweet is after beginning of intervals; can't be sure " +
					"if we've already posted or not so electing not to",
			)
		}

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
