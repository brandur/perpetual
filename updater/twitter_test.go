package updater

import (
	"fmt"
)

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

func (a *mockTwitterAPI) PostTweet(message string) (Tweet, error) {
	fmt.Printf("Posting tweet: %v\n", message)
	return Tweet{Message: message}, nil
}
