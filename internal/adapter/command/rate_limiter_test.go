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
		ok, rem := rl.Allow("user1", "ping")
		require.True(t, ok)
		assert.Zero(t, rem)
	})

	t.Run("blocks second invocation within cooldown", func(t *testing.T) {
		rl := command.NewRateLimiter(5 * time.Second)
		rl.Allow("user1", "ping")

		ok, rem := rl.Allow("user1", "ping")
		assert.False(t, ok)
		assert.Positive(t, rem)
		assert.LessOrEqual(t, rem, 5*time.Second)
	})

	t.Run("allows invocation after cooldown expires", func(t *testing.T) {
		rl := command.NewRateLimiter(10 * time.Millisecond)
		rl.Allow("user1", "ping")

		time.Sleep(15 * time.Millisecond)

		ok, rem := rl.Allow("user1", "ping")
		require.True(t, ok)
		assert.Zero(t, rem)
	})

	t.Run("different commands for same user are independent", func(t *testing.T) {
		rl := command.NewRateLimiter(5 * time.Second)
		rl.Allow("user1", "ping")

		ok, _ := rl.Allow("user1", "help")
		assert.True(t, ok, "a different command should not be blocked")
	})

	t.Run("different users for same command are independent", func(t *testing.T) {
		rl := command.NewRateLimiter(5 * time.Second)
		rl.Allow("user1", "ping")

		ok, _ := rl.Allow("user2", "ping")
		assert.True(t, ok, "a different user should not be blocked")
	})
}
