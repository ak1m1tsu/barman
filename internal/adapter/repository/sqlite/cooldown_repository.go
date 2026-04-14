package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// CooldownRepository implements cooldown.Repository using SQLite.
type CooldownRepository struct {
	db *sql.DB
}

func NewCooldownRepository(db *sql.DB) *CooldownRepository {
	return &CooldownRepository{db: db}
}

func (r *CooldownRepository) GetUsedAt(ctx context.Context, userID string) (time.Time, error) {
	var usedAt time.Time
	err := r.db.QueryRowContext(ctx,
		`SELECT used_at FROM bot_reaction_cooldowns WHERE user_id = ?`, userID,
	).Scan(&usedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, nil
	}
	return usedAt, err
}

func (r *CooldownRepository) SetUsedAt(ctx context.Context, userID string, t time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO bot_reaction_cooldowns (user_id, used_at) VALUES (?, ?)
		 ON CONFLICT(user_id) DO UPDATE SET used_at = excluded.used_at`,
		userID, t,
	)
	return err
}
