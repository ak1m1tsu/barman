package command

import (
	"fmt"
	"sync"
	"time"
)

// SpamThreshold is the number of consecutive blocked attempts after which
// the bot switches to the anti-spam message.
const SpamThreshold = 5

// RateLimiter enforces a per-user per-command cooldown in memory.
// It is safe for concurrent use. Pass nil to Registry to disable rate limiting.
type RateLimiter struct {
	mu         sync.Mutex
	last       map[string]time.Time
	violations map[string]int
	cooldown   time.Duration
}

// NewRateLimiter returns a RateLimiter that allows each user to invoke
// the same command at most once per cooldown duration.
func NewRateLimiter(cooldown time.Duration) *RateLimiter {
	return &RateLimiter{
		last:       make(map[string]time.Time),
		violations: make(map[string]int),
		cooldown:   cooldown,
	}
}

// RateLimitMessage returns the reply to send when a user is rate-limited.
// After SpamThreshold consecutive blocked attempts the message becomes harsher.
func RateLimitMessage(violations int, remaining time.Duration) string {
	if violations >= SpamThreshold {
		return "Не спамь долбаеб тупорылый"
	}
	secs := int(remaining.Round(time.Second).Seconds())
	return fmt.Sprintf("⏳ Подождите **%d сек.** перед следующей командой.", secs)
}

// Allow returns (true, 0, 0) when the user may invoke the command.
// Returns (false, remaining, violations) when still on cooldown, where
// violations is the count of consecutive blocked attempts for this key.
// The violation counter resets to 0 whenever the user is allowed through.
func (rl *RateLimiter) Allow(userID, command string) (bool, time.Duration, int) {
	key := userID + ":" + command

	rl.mu.Lock()
	defer rl.mu.Unlock()

	last, ok := rl.last[key]
	if ok {
		if elapsed := time.Since(last); elapsed < rl.cooldown {
			rl.violations[key]++
			return false, rl.cooldown - elapsed, rl.violations[key]
		}
	}

	rl.last[key] = time.Now()
	rl.violations[key] = 0
	return true, 0, 0
}
