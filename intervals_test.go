package main

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

const maxTweetLength = 280

// Makes sure that all configured intervals are below the maximum length of a
// tweet.
func TestIntervalLengths(t *testing.T) {
	for _, interval := range intervals {
		assert.True(t, len(interval.Message) < maxTweetLength)
	}
}