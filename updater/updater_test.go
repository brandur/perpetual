package updater

import (
	"fmt"
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
)

func TestIntervalPattern(t *testing.T) {
	assert.True(t, intervalPattern.MatchString("LHI000: "))
	assert.True(t, intervalPattern.MatchString("LHI001: "))
	assert.True(t, intervalPattern.MatchString("LHI245: "))

	assert.False(t, intervalPattern.MatchString("LHI0x1: "))
	assert.False(t, intervalPattern.MatchString(" LHI001: "))
	assert.False(t, intervalPattern.MatchString("LHI001 should be coming soon!"))
	assert.False(t, intervalPattern.MatchString("just a normal string"))
}

func TestExtractIntervalID(t *testing.T) {
	{
		id, ok := extractIntervalID("LHI000: ")
		assert.Equal(t, 0, id)
		assert.True(t, ok)
	}

	{
		id, ok := extractIntervalID("LHI001: ")
		assert.Equal(t, 1, id)
		assert.True(t, ok)
	}

	{
		id, ok := extractIntervalID("LHI245: ")
		assert.Equal(t, 245, id)
		assert.True(t, ok)
	}

	{
		id, ok := extractIntervalID("just a normal string")
		assert.Equal(t, -1, id)
		assert.False(t, ok)
	}
}

func TestFormatInterval(t *testing.T) {
	assert.Equal(t, "LHI000: hello", FormatInterval(0, "hello"))
	assert.Equal(t, "LHI001: hello, there", FormatInterval(1, "hello, there"))
	assert.Equal(t, "LHI999: goodbye", FormatInterval(999, "goodbye"))
}

func TestUpdate(t *testing.T) {
	now := time.Now()

	// We have a check to make sure that intervals are after tweets that are
	// listed from Twitter, so just make sure any of the messages we preload
	// were posted in the past.
	past := now.Add(-1 * time.Second)

	// A special error case: if there were no posted intervals and the last
	// tweet was after the beginning of our intervals, we don't post because it
	// can't be sure whether we did already or not.
	{
		future := now.Add(1 * time.Second)

		_, err := Update(
			&mockTwitterAPI{tweets: []*Tweet{
				{CreatedAt: future, Message: "this is a tweet"},
			}},
			[]*Interval{
				{Target: now, Message: "Interval 000"},
			},
			now,
		)
		assert.Error(t, fmt.Errorf(
			"Last available tweet is after beginning of intervals; can't be sure "+
				"if we've already posted or not so electing not to",
		), err)
	}

	// Posts a first interval given none existing
	{
		id, err := Update(
			&mockTwitterAPI{tweets: []*Tweet{
				{CreatedAt: past, Message: "this is a tweet"},
				{CreatedAt: past, Message: "tweet"},
				{CreatedAt: past, Message: "first tweet"},
			}},
			[]*Interval{
				{Target: now, Message: "Interval 000"},
			},
			now,
		)
		assert.NoError(t, err)
		assert.Equal(t, 0, id)
	}

	// Posts nothing if the interval is already posted
	{
		id, err := Update(
			&mockTwitterAPI{tweets: []*Tweet{
				{CreatedAt: past, Message: "this is a tweet"},
				{CreatedAt: past, Message: "LHI000: Interval 000"},
				{CreatedAt: past, Message: "tweet"},
				{CreatedAt: past, Message: "first tweet"},
			}},
			[]*Interval{
				{Target: now, Message: "Interval 000"},
			},
			now,
		)
		assert.NoError(t, err)
		assert.Equal(t, -1, id)
	}

	// Posts nothing if the interval is already posted and future interval is not ready
	{
		assert.False(t, now.After(now.Add(2*time.Minute)))

		id, err := Update(
			&mockTwitterAPI{tweets: []*Tweet{
				{CreatedAt: past, Message: "this is a tweet"},
				{CreatedAt: past, Message: "LHI000: Interval 000"},
				{CreatedAt: past, Message: "tweet"},
				{CreatedAt: past, Message: "first tweet"},
			}},
			[]*Interval{
				{Target: now, Message: "Interval 000"},
				{Target: now.Add(2 * time.Minute), Message: "Interval 001"},
			},
			now,
		)
		assert.NoError(t, err)
		assert.Equal(t, -1, id)
	}

	// Posts a second interval given one existing
	{
		id, err := Update(
			&mockTwitterAPI{tweets: []*Tweet{
				{CreatedAt: past, Message: "this is a tweet"},
				{CreatedAt: past, Message: "LHI000: Interval 000"},
				{CreatedAt: past, Message: "tweet"},
				{CreatedAt: past, Message: "first tweet"},
			}},
			[]*Interval{
				{Target: now, Message: "Interval 000"},
				{Target: now, Message: "Interval 001"},
			},
			now,
		)
		assert.NoError(t, err)
		assert.Equal(t, 1, id)
	}

	// Posts nothing if both intervals are already posted
	{
		id, err := Update(
			&mockTwitterAPI{tweets: []*Tweet{
				{CreatedAt: past, Message: "LHI001: Interval 001"},
				{CreatedAt: past, Message: "this is a tweet"},
				{CreatedAt: past, Message: "LHI000: Interval 000"},
				{CreatedAt: past, Message: "tweet"},
				{CreatedAt: past, Message: "first tweet"},
			}},
			[]*Interval{
				{Target: now, Message: "Interval 000"},
				{Target: now, Message: "Interval 001"},
			},
			now,
		)
		assert.NoError(t, err)
		assert.Equal(t, -1, id)
	}

	// Tests a full ladder of intervals. This one will probably be harder to debug,
	// so hopefully any real problems get caught by one of the simple cases
	// above.
	{
		tweets := []*Tweet{
			{CreatedAt: past, Message: "first tweet"},
		}

		intervals := []*Interval{
			{Target: now.Add(0 * time.Minute), Message: "Interval 000"},
			{Target: now.Add(1 * time.Minute), Message: "Interval 001"},
			{Target: now.Add(2 * time.Minute), Message: "Interval 002"},
			{Target: now.Add(3 * time.Minute), Message: "Interval 003"},
			{Target: now.Add(4 * time.Minute), Message: "Interval 004"},
			{Target: now.Add(5 * time.Minute), Message: "Interval 005"},
			{Target: now.Add(6 * time.Minute), Message: "Interval 006"},
			{Target: now.Add(7 * time.Minute), Message: "Interval 007"},
			{Target: now.Add(8 * time.Minute), Message: "Interval 008"},
			{Target: now.Add(9 * time.Minute), Message: "Interval 009"},
		}

		for i := 0; i < 10; i++ {
			targetNow := now.Add(time.Duration(i) * time.Minute)

			// At one second before target, make sure nothing gets posted
			{
				id, err := Update(
					&mockTwitterAPI{tweets: tweets},
					intervals,
					targetNow.Add(-1*time.Second),
				)
				assert.NoError(t, err)
				assert.Equal(t, -1, id)
			}

			// At one second after target, make sure we do post
			{
				id, err := Update(
					&mockTwitterAPI{tweets: tweets},
					intervals,
					targetNow.Add(1*time.Second),
				)
				assert.NoError(t, err)
				assert.Equal(t, i, id)
			}

			// Add this interval to the mock API (note it gets *prepended* because
			// tweets are iterated in reverse chronological order) so that the
			// next iteration of the loop will behave as expected
			tweets = append([]*Tweet{
				{CreatedAt: past, Message: FormatInterval(i, intervals[i].Message)},
			}, tweets...)

			// Test a duplicate operation: now that our message is in the list,
			// nothing should get posted
			{
				id, err := Update(
					&mockTwitterAPI{tweets: tweets},
					intervals,
					targetNow.Add(2*time.Second),
				)
				assert.NoError(t, err)
				assert.Equal(t, -1, id)
			}
		}
	}
}
