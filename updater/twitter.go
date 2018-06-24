package updater

// Tweet represents a tweet returned from Twitter's API.
type Tweet struct {
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
	Next() bool
	Value() Tweet
}

// TwitterAPI is a subset of the implementation of Twitter's API needed for the
// purposes of this project.
type TwitterAPI interface {
	ListTweets() TweetIterator
	PostTweet(message string) (Tweet, error)
}
