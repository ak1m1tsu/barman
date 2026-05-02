package command

import (
	"sync"
	"time"
)

// RateLimiter enforces a per-user per-command cooldown in memory.
// It is safe for concurrent use. Pass nil to Registry to disable rate limiting.
type RateLimiter struct {
	mu       sync.Mutex
	last     map[string]time.Time
	cooldown time.Duration
}

// NewRateLimiter returns a RateLimiter that allows each user to invoke
// the same command at most once per cooldown duration.
func NewRateLimiter(cooldown time.Duration) *RateLimiter {
	return &RateLimiter{
		last:     make(map[string]time.Time),
		cooldown: cooldown,
	}
}

// Allow returns true and 0 when the user may invoke the command.
// Returns false and the remaining wait duration when still on cooldown.
func (rl *RateLimiter) Allow(userID, command string) (bool, time.Duration) {
	key := userID + ":" + command

	rl.mu.Lock()
	defer rl.mu.Unlock()

	last, ok := rl.last[key]
	if ok {
		if elapsed := time.Since(last); elapsed < rl.cooldown {
			return false, rl.cooldown - elapsed
		}
	}

	rl.last[key] = time.Now()
	return true, 0
}
