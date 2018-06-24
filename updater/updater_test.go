package updater

import (
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
)

func TestAeonPattern(t *testing.T) {
	assert.True(t, aeonPattern.MatchString("IE000: "))
	assert.True(t, aeonPattern.MatchString("IE001: "))
	assert.True(t, aeonPattern.MatchString("IE245: "))

	assert.False(t, aeonPattern.MatchString("IE0x1: "))
	assert.False(t, aeonPattern.MatchString(" IE001: "))
	assert.False(t, aeonPattern.MatchString("IE001 should be coming soon!"))
	assert.False(t, aeonPattern.MatchString("just a normal string"))
}

func TestExtractAeonID(t *testing.T) {
	{
		id, ok := extractAeonID("IE000: ")
		assert.Equal(t, 0, id)
		assert.True(t, ok)
	}

	{
		id, ok := extractAeonID("IE001: ")
		assert.Equal(t, 1, id)
		assert.True(t, ok)
	}

	{
		id, ok := extractAeonID("IE245: ")
		assert.Equal(t, 245, id)
		assert.True(t, ok)
	}

	{
		id, ok := extractAeonID("just a normal string")
		assert.Equal(t, -1, id)
		assert.False(t, ok)
	}
}

func TestFormatAeon(t *testing.T) {
	assert.Equal(t, "IE000: hello", formatAeon(0, "hello"))
	assert.Equal(t, "IE001: hello, there", formatAeon(1, "hello, there"))
	assert.Equal(t, "IE999: goodbye", formatAeon(999, "goodbye"))
}

func TestUpdate(t *testing.T) {
	now := time.Now()

	// Posts a first aeon given none existing
	{
		id, err := Update(
			&mockTwitterAPI{messages: []string{
				"this is a tweet",
				"tweet",
				"first tweet",
			}},
			[]*Aeon{
				{Target: now, Message: "Aeon 000"},
			},
			now,
		)
		assert.NoError(t, err)
		assert.Equal(t, 0, id)
	}

	// Posts nothing if the aeon is already posted
	{
		id, err := Update(
			&mockTwitterAPI{messages: []string{
				"this is a tweet",
				"IE000: Aeon 000",
				"tweet",
				"first tweet",
			}},
			[]*Aeon{
				{Target: now, Message: "Aeon 000"},
			},
			now,
		)
		assert.NoError(t, err)
		assert.Equal(t, -1, id)
	}

	// Posts nothing if the aeon is already posted and future aeon is not ready
	{
		assert.False(t, now.After(now.Add(2*time.Minute)))

		id, err := Update(
			&mockTwitterAPI{messages: []string{
				"this is a tweet",
				"IE000: Aeon 000",
				"tweet",
				"first tweet",
			}},
			[]*Aeon{
				{Target: now, Message: "Aeon 000"},
				{Target: now.Add(2 * time.Minute), Message: "Aeon 001"},
			},
			now,
		)
		assert.NoError(t, err)
		assert.Equal(t, -1, id)
	}

	// Posts a second aeon given one existing
	{
		id, err := Update(
			&mockTwitterAPI{messages: []string{
				"this is a tweet",
				"IE000: Aeon 000",
				"tweet",
				"first tweet",
			}},
			[]*Aeon{
				{Target: now, Message: "Aeon 000"},
				{Target: now, Message: "Aeon 001"},
			},
			now,
		)
		assert.NoError(t, err)
		assert.Equal(t, 1, id)
	}

	// Posts nothing if both aeons are already posted
	{
		id, err := Update(
			&mockTwitterAPI{messages: []string{
				"IE001: Aeon 001",
				"this is a tweet",
				"IE000: Aeon 000",
				"tweet",
				"first tweet",
			}},
			[]*Aeon{
				{Target: now, Message: "Aeon 000"},
				{Target: now, Message: "Aeon 001"},
			},
			now,
		)
		assert.NoError(t, err)
		assert.Equal(t, -1, id)
	}

	// Tests a full ladder of aeons. This one will probably be harder to debug,
	// so hopefully any real problems get caught by one of the simple cases
	// above.
	{
		messages := []string{
			"first tweet",
		}

		aeons := []*Aeon{
			{Target: now.Add(0 * time.Minute), Message: "Aeon 000"},
			{Target: now.Add(1 * time.Minute), Message: "Aeon 001"},
			{Target: now.Add(2 * time.Minute), Message: "Aeon 002"},
			{Target: now.Add(3 * time.Minute), Message: "Aeon 003"},
			{Target: now.Add(4 * time.Minute), Message: "Aeon 004"},
			{Target: now.Add(5 * time.Minute), Message: "Aeon 005"},
			{Target: now.Add(6 * time.Minute), Message: "Aeon 006"},
			{Target: now.Add(7 * time.Minute), Message: "Aeon 007"},
			{Target: now.Add(8 * time.Minute), Message: "Aeon 008"},
			{Target: now.Add(9 * time.Minute), Message: "Aeon 009"},
		}

		for i := 0; i < 10; i++ {
			targetNow := now.Add(time.Duration(i) * time.Minute)

			// At one second before target, make sure nothing gets posted
			{
				id, err := Update(
					&mockTwitterAPI{messages: messages},
					aeons,
					targetNow.Add(-1*time.Second),
				)
				assert.NoError(t, err)
				assert.Equal(t, -1, id)
			}

			// At one second after target, make sure we do post
			{
				id, err := Update(
					&mockTwitterAPI{messages: messages},
					aeons,
					targetNow.Add(1*time.Second),
				)
				assert.NoError(t, err)
				assert.Equal(t, i, id)
			}

			// Add this aeon to the mock API (note it gets *prepended* because
			// tweets are iterated in reverse chronological order) so that the
			// next iteration of the loop will behave as expected
			messages = append([]string{formatAeon(i, aeons[i].Message)}, messages...)

			// Test a duplicate operation: now that our message is in the list,
			// nothing should get posted
			{
				id, err := Update(
					&mockTwitterAPI{messages: messages},
					aeons,
					targetNow.Add(2*time.Second),
				)
				assert.NoError(t, err)
				assert.Equal(t, -1, id)
			}
		}
	}
}
