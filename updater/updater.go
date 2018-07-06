package updater

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var aeonFormat = "IE%03d: %s"
var aeonPattern = regexp.MustCompile(`^IE(\d{3}): `)

// Update iterates through an account's tweets as far back as necessary to
// discover the last posted aeon, then decides whether or not to post a new
// aeon based off of the next aeon's target time.
//
// now is injected as a parameter for better testability. It's safe to pass
// this as time.Now in most cases.
//
// Update returns an integer representing the ID of the aeon that was posted if
// one was posted, or -1 otherwise. An error is returned if there was a problem
// communicating with Twitter's API.
func Update(api TwitterAPI, aeons []*Aeon, now time.Time) (int, error) {
	var id int
	var ok bool

	it := api.ListTweets()

	fmt.Printf("Iterating backward through tweets\n")

	// Keep in mind that we expect our API to return tweets in reverse order
	// (i.e., newest first). Many assumptions are built into this code to take
	// advantage of that.
	for it.Next() {
		tweet := it.Value()

		id, ok = extractAeonID(tweet.Message)
		if ok {
			fmt.Printf("Found aeon ID: %v\n", id)
			break
		}
	}

	if it.Err() != nil {
		return -1, it.Err()
	}

	var nextAeonID int
	if ok {
		// Pick the next in the series
		nextAeonID = id + 1
	} else {
		// If ok is false, we never extracted an aeon ID, which means that this
		// program has never posted before. Pick the first aeon ID in the
		// series.
		nextAeonID = 0
	}

	fmt.Printf("Next aeon ID: %v\n", nextAeonID)

	if nextAeonID >= len(aeons) {
		fmt.Printf("There is no next aeon; this program is done\n")
		return -1, nil
	}

	// Check if the Aeon is ready to be posted
	aeon := aeons[nextAeonID]

	if aeon.Target.After(now) {
		fmt.Printf("Aeon not ready, target: %v\n", aeon.Target)
		return -1, nil
	}

	tweet, err := api.PostTweet(formatAeon(nextAeonID, aeon.Message))
	if err != nil {
		return -1, err
	}

	fmt.Printf("Posted tweet: %+v\n", tweet)
	return nextAeonID, nil
}

func extractAeonID(content string) (int, bool) {
	matches := aeonPattern.FindAllStringSubmatch(content, 1)

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

func formatAeon(id int, message string) string {
	return fmt.Sprintf(aeonFormat, id, message)
}
