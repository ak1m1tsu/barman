package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ak1m1tsu/barman/internal/domain/guild"
)

// GuildRepository implements guild.Repository using SQLite.
type GuildRepository struct {
	db *sql.DB
}

// NewGuildRepository returns a GuildRepository backed by the given SQLite connection.
func NewGuildRepository(db *sql.DB) *GuildRepository {
	return &GuildRepository{db: db}
}

func (r *GuildRepository) FindByID(ctx context.Context, guildID string) (*guild.Guild, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT guild_id, auto_role_id, prefix FROM guild_settings WHERE guild_id = ?`, guildID)

	g := &guild.Guild{}
	if err := row.Scan(&g.ID, &g.AutoRoleID, &g.Prefix); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return g, nil
}

func (r *GuildRepository) Save(ctx context.Context, g *guild.Guild) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO guild_settings (guild_id, auto_role_id, prefix) VALUES (?, ?, ?)
		 ON CONFLICT(guild_id) DO UPDATE SET
		   auto_role_id = excluded.auto_role_id,
		   prefix       = excluded.prefix`,
		g.ID, g.AutoRoleID, g.Prefix)
	return err
}

func (r *GuildRepository) Delete(ctx context.Context, guildID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM guild_settings WHERE guild_id = ?`, guildID)
	return err
}
