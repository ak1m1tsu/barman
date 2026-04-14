package cooldown

import (
	"context"
	"time"
)

// Repository persists per-user bot-reaction cooldown timestamps.
type Repository interface {
	// GetUsedAt returns the last time the user triggered a bot reaction.
	// Returns zero time and nil error if the user has no record.
	GetUsedAt(ctx context.Context, userID string) (time.Time, error)

	// SetUsedAt records the current time for the user.
	SetUsedAt(ctx context.Context, userID string, t time.Time) error
}
