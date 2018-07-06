package updater

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

//
// Common interface/types
//

// Tweet represents a tweet returned from Twitter's API.
type Tweet struct {
	ID      uint64
	Message string
}

// TweetIterator iterates through a list of tweets that are returned from
// Twitter's API in a way that accounts for pagination.
//
// And initial call to Next should be made, and a decision as to whether to
// continue should depend on whether it returns true or false. If true, a call
// to Value can be used to get the iterator's current value before continuing
// iteration.
//
// Its iteration order is always expected to be in reverse chronological order.
// Its implementations guarantee this.
type TweetIterator interface {
	Err() error
	Next() bool
	Value() Tweet
}

// TwitterAPI is a subset of the implementation of Twitter's API needed for the
// purposes of this project.
type TwitterAPI interface {
	ListTweets() TweetIterator
	PostTweet(message string) (*Tweet, error)
}

//
// Live implementation
//

// LiveTweetIterator is a tweet iterator for the live Twitter API.
type LiveTweetIterator struct {
	// A pointer back to the API instance that generated this iterator.
	api *LiveTwitterAPI

	// The set of tweets that were retrieved on the last page. After we reach
	// the end of these, we'll need to ask for another page.
	currentTweets []*Tweet

	// A flag that we set once we've definitely reached the end of iteration.
	done bool

	// An error that the iterator encountered (if it encountered one).
	err error

	// The last ID of the last page of messages (we use this to calculate
	// `max_id` the next time we fetch a page).
	lastID uint64

	// Our position within the current page (in currentTweets).
	position int
}

// liveTweet is a tweet that we decoded in a response from the Twitter API.
type liveTweet struct {
	ID   uint64 `json:"id"`
	Text string `json:"text"`
}

// Err gets an error set on the iterator.
func (it *LiveTweetIterator) Err() error {
	return it.err
}

// Next moves the iterator to its next value.
func (it *LiveTweetIterator) Next() bool {
	if it.done {
		return false
	}

	// If we still have tweets left to consume on this page, do that
	if it.position != -1 && it.position < len(it.currentTweets)-1 {
		it.position++
		return true
	}

	fmt.Printf("\nRequesting next page (max ID = %v)\n\n", it.lastID)

	req, err := it.api.newAuthorizedRequest(
		"GET", "https://api.twitter.com/1.1/statuses/user_timeline.json")
	if err != nil {
		it.err = err
		return false
	}

	query := req.URL.Query()
	query.Add("count", "200") // 200 is the largest page allowed
	query.Add("exclude_replies", "true")
	query.Add("include_rts", "false")
	query.Add("screen_name", it.api.ScreenName)
	query.Add("trim_user", "true")

	// If this isn't the first page, ask for the next sequence by subtracting
	// one from the last ID of the last page that we processed.
	if it.lastID != 0 {
		query.Add("max_id", strconv.FormatUint(it.lastID-1, 10))

		// Also, sleep one second if this isn't the first request so we don't
		// hit a rate limit. This particular Twitter API allows one request per
		// second.
		time.Sleep(1 * time.Second)
	}

	var tweets []*liveTweet
	err = it.api.encodeAndExecuteRequest(req, query, &tweets)
	if err != nil {
		it.err = err
		return false
	}

	if len(tweets) < 1 {
		it.done = true
		return false
	}

	it.currentTweets = make([]*Tweet, len(tweets))
	for i, v := range tweets {
		it.currentTweets[i] = &Tweet{ID: v.ID, Message: v.Text}
	}

	// Set the page's last ID so we know where to start on the next iteration
	it.lastID = tweets[len(tweets)-1].ID

	// Reset the cursor to the beginning of the page
	it.position = 0

	return true
}

// Value gets the value of the current element that the iterator is pointing
// to.
func (it *LiveTweetIterator) Value() Tweet {
	if it.err != nil {
		panic("Iterator encountered an error; access it using Err")
	}

	if it.position == -1 {
		panic("Must call Next on iterator before a call to Value is allowed")
	}

	return *it.currentTweets[it.position]
}

// LiveTwitterAPI is an API implementation for the live Twitter API.
type LiveTwitterAPI struct {
	// AccessToken is the base64-encoded access token used to authorize against
	// the Twitter API.
	AccessToken string

	// ScreenName is the Twitter screen name that will be read from and posted to.
	ScreenName string
}

// ListTweets returns an iterator for the configured account's live tweets.
func (a *LiveTwitterAPI) ListTweets() TweetIterator {
	return &LiveTweetIterator{api: a, lastID: 0, position: -1}
}

// PostTweet posts a tweet to the configured account.
func (a *LiveTwitterAPI) PostTweet(message string) (*Tweet, error) {
	req, err := a.newAuthorizedRequest("POST", "https://api.twitter.com/1.1/statuses/update.json")
	if err != nil {
		return nil, err
	}

	fmt.Printf("Posting tweet: %v\n", message)

	query := req.URL.Query()
	query.Add("status", message)

	var tweet *liveTweet
	err = a.encodeAndExecuteRequest(req, query, &tweet)
	if err != nil {
		return nil, err
	}

	return &Tweet{ID: tweet.ID, Message: tweet.Text}, nil
}

func (a *LiveTwitterAPI) encodeAndExecuteRequest(
	req *http.Request, query url.Values, v interface{}) error {

	req.URL.RawQuery = query.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"Improper response from the Twitter API (status: %v): %s",
			resp.Status,
			string(data))
	}

	return json.Unmarshal(data, v)
}

func (a *LiveTwitterAPI) newAuthorizedRequest(method, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+a.AccessToken)

	return req, nil
}
