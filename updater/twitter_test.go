package updater

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/dghubble/oauth1"
	assert "github.com/stretchr/testify/require"
)

//
// Mock Twitter API
//

type mockTweetIterator struct {
	messages []string
	position int
}

func (i *mockTweetIterator) Next() bool {
	i.position++
	if i.position >= len(i.messages) {
		return false
	}
	return true
}

func (i *mockTweetIterator) Err() error {
	return nil
}

func (i *mockTweetIterator) Value() Tweet {
	if i.position == -1 {
		panic("Must call Next on iterator before a call to Value is allowed")
	}

	return Tweet{Message: i.messages[i.position]}
}

type mockTwitterAPI struct {
	messages []string
}

func (a *mockTwitterAPI) ListTweets() TweetIterator {
	return &mockTweetIterator{messages: a.messages, position: -1}
}

func (a *mockTwitterAPI) PostTweet(message string) (*Tweet, error) {
	fmt.Printf("Posting tweet: %v\n", message)
	return &Tweet{Message: message}, nil
}

//
// Test for LiveTwitterAPI
//

func TestLiveTwitterAPI_ListTweets(t *testing.T) {
	t.Skip("Makes live API requests")

	api := getLiveTwitterAPI()

	it := api.ListTweets()
	for it.Next() {
		tweet := it.Value()

		message := tweet.Message
		if len(message) > 50 {
			message = message[0:50]
		}
		message = strings.Replace(message, "\n", "", -1)

		fmt.Printf("	ID: %v | Tweet: %s\n", tweet.ID, message)
	}

	assert.NoError(t, it.Err())
}

func TestLiveTwitterAPI_PostTweet(t *testing.T) {
	t.Skip("Makes live API requests")

	api := getLiveTwitterAPI()

	tweet, err := api.PostTweet("Hello from Perpetual.")
	assert.NoError(t, err)

	fmt.Printf("Posted tweet: %+v\n", tweet)
}

//
// Helpers
//

func getLiveTwitterAPI() TwitterAPI {
	fmt.Printf("Access token = %s\n", os.Getenv("ACCESS_TOKEN"))
	fmt.Printf("Access token secret = %s\n", os.Getenv("ACCESS_TOKEN_SECRET"))

	fmt.Printf("Consumer key = %s\n", os.Getenv("CONSUMER_KEY"))
	fmt.Printf("Consumer secret = %s\n", os.Getenv("CONSUMER_SECRET"))

	fmt.Printf("Screen name = %s\n", os.Getenv("SCREEN_NAME"))

	config := oauth1.NewConfig(os.Getenv("CONSUMER_KEY"), os.Getenv("CONSUMER_SECRET"))
	token := oauth1.NewToken(os.Getenv("ACCESS_TOKEN"), os.Getenv("ACCESS_TOKEN_SECRET"))
	httpClient := config.Client(oauth1.NoContext, token)

	return &LiveTwitterAPI{
		HTTPClient: httpClient,
		ScreenName: os.Getenv("SCREEN_NAME"),
	}
}
