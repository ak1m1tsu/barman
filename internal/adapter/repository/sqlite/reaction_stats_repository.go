package sqlite

import (
	"context"
	"database/sql"
)

// ReactionStatsRepository implements reaction.StatsRepository using SQLite.
type ReactionStatsRepository struct {
	db *sql.DB
}

// NewReactionStatsRepository returns a ReactionStatsRepository backed by the given SQLite connection.
func NewReactionStatsRepository(db *sql.DB) *ReactionStatsRepository {
	return &ReactionStatsRepository{db: db}
}

func (r *ReactionStatsRepository) Increment(ctx context.Context, reactionType string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO reaction_stats (reaction_type, count) VALUES (?, 1)
		 ON CONFLICT(reaction_type) DO UPDATE SET count = count + 1`,
		reactionType,
	)
	return err
}

func (r *ReactionStatsRepository) GetAll(ctx context.Context) (map[string]int64, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT reaction_type, count FROM reaction_stats`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	stats := make(map[string]int64)
	for rows.Next() {
		var reactionType string
		var count int64
		if err := rows.Scan(&reactionType, &count); err != nil {
			return nil, err
		}
		stats[reactionType] = count
	}
	return stats, rows.Err()
}
