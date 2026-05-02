package command_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ak1m1tsu/barman/internal/adapter/command"
)

func TestRateLimiter_Allow(t *testing.T) {
	t.Run("allows first invocation", func(t *testing.T) {
		rl := command.NewRateLimiter(5 * time.Second)
		ok, rem, violations := rl.Allow("user1", "ping")
		require.True(t, ok)
		assert.Zero(t, rem)
		assert.Zero(t, violations)
	})

	t.Run("blocks second invocation within cooldown", func(t *testing.T) {
		rl := command.NewRateLimiter(5 * time.Second)
		rl.Allow("user1", "ping")

		ok, rem, violations := rl.Allow("user1", "ping")
		assert.False(t, ok)
		assert.Positive(t, rem)
		assert.LessOrEqual(t, rem, 5*time.Second)
		assert.Equal(t, 1, violations)
	})

	t.Run("violation count increments on each blocked call", func(t *testing.T) {
		rl := command.NewRateLimiter(5 * time.Second)
		rl.Allow("user1", "ping")

		for i := 1; i <= command.SpamThreshold+2; i++ {
			_, _, v := rl.Allow("user1", "ping")
			assert.Equal(t, i, v)
		}
	})

	t.Run("violation count resets after cooldown expires", func(t *testing.T) {
		rl := command.NewRateLimiter(10 * time.Millisecond)
		rl.Allow("user1", "ping")
		rl.Allow("user1", "ping") // violation = 1

		time.Sleep(15 * time.Millisecond)

		ok, _, violations := rl.Allow("user1", "ping")
		require.True(t, ok)
		assert.Zero(t, violations)
	})

	t.Run("allows invocation after cooldown expires", func(t *testing.T) {
		rl := command.NewRateLimiter(10 * time.Millisecond)
		rl.Allow("user1", "ping")

		time.Sleep(15 * time.Millisecond)

		ok, rem, _ := rl.Allow("user1", "ping")
		require.True(t, ok)
		assert.Zero(t, rem)
	})

	t.Run("different commands for same user are independent", func(t *testing.T) {
		rl := command.NewRateLimiter(5 * time.Second)
		rl.Allow("user1", "ping")

		ok, _, _ := rl.Allow("user1", "help")
		assert.True(t, ok, "a different command should not be blocked")
	})

	t.Run("different users for same command are independent", func(t *testing.T) {
		rl := command.NewRateLimiter(5 * time.Second)
		rl.Allow("user1", "ping")

		ok, _, _ := rl.Allow("user2", "ping")
		assert.True(t, ok, "a different user should not be blocked")
	})
}
