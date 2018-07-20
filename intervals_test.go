package main

import (
	"testing"

	"github.com/brandur/perpetual/updater"
	assert "github.com/stretchr/testify/require"
)

const maxTweetLength = 280

// Makes sure that all configured intervals are below the maximum length of a
// tweet.
func TestIntervalLengths(t *testing.T) {
	for _, interval := range intervals {
		length := len(updater.FormatInterval(0, interval.Message))
		assert.True(t,
			length < maxTweetLength,
			"Interval message is too long for a tweet: %s (%v characters)",
			interval.Message,
			length,
		)
	}
}
