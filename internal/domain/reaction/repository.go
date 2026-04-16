package reaction

import "context"

// StatsRepository persists global reaction usage counters.
type StatsRepository interface {
	// Increment adds 1 to the counter for the given reaction type.
	Increment(ctx context.Context, reactionType string) error

	// GetAll returns a map of reaction type → total usage count.
	// Types with no recorded usage are not included.
	GetAll(ctx context.Context) (map[string]int64, error)
}
