package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/brandur/perpetual/updater"
	"github.com/dghubble/oauth1"
)

// Event is an event to be passed into the AWS Lambda handler.
type Event struct {
}

// HandleRequest is the target to be invoked by AWS Lambda.
func HandleRequest(ctx context.Context, event Event) (string, error) {
	consumerKey, err := mustEnv("CONSUMER_KEY")
	if err != nil {
		return "", err
	}
	consumerSecret, err := mustEnv("CONSUMER_SECRET")
	if err != nil {
		return "", err
	}
	accessToken, err := mustEnv("ACCESS_TOKEN")
	if err != nil {
		return "", err
	}
	accessTokenSecret, err := mustEnv("ACCESS_TOKEN_SECRET")
	if err != nil {
		return "", err
	}
	screenName, err := mustEnv("SCREEN_NAME")
	if err != nil {
		return "", err
	}

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	api := &updater.LiveTwitterAPI{
		HTTPClient: httpClient,
		ScreenName: screenName,
	}

	_, err = updater.Update(api, intervals, time.Now())
	if err != nil {
		return "", err
	}

	return "Successfully ran check", nil
}

func main() {
	lambda.Start(HandleRequest)
}

//
// Private
//

// hundredYears represents a hundred years of time. This is used as a hack to
// get to ten thousand years, and is a convenient large block of time to add
// up which still fits within time.Duration's maximum size of ~290 years.
const hundredYears = time.Hour * 24 * 365 * 100

// See `intervals.go`
var intervals []*updater.Interval

// time.Parse won't parse a 5-digit years, so we need a little hackiness to get
// us to ten thousand. Unfortunately, this is nowhere near as clean as other
// times because we know that leap years, etc. won't be handled well. This is
// probably made okay because it's unlikely that we ever get there.
func addTenThousandYears(t time.Time) time.Time {
	// 100 * 100 = 10,000
	for i := 0; i < 100; i++ {
		t.Add(hundredYears)
	}
	return t
}

func mustEnv(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", fmt.Errorf("need env key: %s", val)
	}
	return val, nil
}
