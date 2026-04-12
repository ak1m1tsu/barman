package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	"github.com/ak1m1tsu/barman/internal/adapter/repository/sqlite"
	domain "github.com/ak1m1tsu/barman/internal/domain/guild"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS guild_settings (
			guild_id     TEXT PRIMARY KEY,
			auto_role_id TEXT NOT NULL DEFAULT '',
			prefix       TEXT NOT NULL DEFAULT ''
		)
	`)
	require.NoError(t, err)

	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestGuildRepository_SaveAndFindByID(t *testing.T) {
	repo := sqlite.NewGuildRepository(newTestDB(t))
	ctx := context.Background()

	t.Run("saves and retrieves guild", func(t *testing.T) {
		g := &domain.Guild{ID: "guild1", AutoRoleID: "role1"}
		require.NoError(t, repo.Save(ctx, g))

		found, err := repo.FindByID(ctx, "guild1")
		require.NoError(t, err)
		assert.Equal(t, g, found)
	})

	t.Run("returns nil for missing guild", func(t *testing.T) {
		found, err := repo.FindByID(ctx, "nonexistent")
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("upserts on duplicate guild_id", func(t *testing.T) {
		g := &domain.Guild{ID: "guild2", AutoRoleID: "role1"}
		require.NoError(t, repo.Save(ctx, g))

		g.AutoRoleID = "role2"
		require.NoError(t, repo.Save(ctx, g))

		found, err := repo.FindByID(ctx, "guild2")
		require.NoError(t, err)
		assert.Equal(t, "role2", found.AutoRoleID)
	})
}

func TestGuildRepository_Delete(t *testing.T) {
	repo := sqlite.NewGuildRepository(newTestDB(t))
	ctx := context.Background()

	t.Run("deletes existing guild", func(t *testing.T) {
		g := &domain.Guild{ID: "guild3", AutoRoleID: "role1"}
		require.NoError(t, repo.Save(ctx, g))

		require.NoError(t, repo.Delete(ctx, "guild3"))

		found, err := repo.FindByID(ctx, "guild3")
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("delete of nonexistent guild is a no-op", func(t *testing.T) {
		assert.NoError(t, repo.Delete(ctx, "nobody"))
	})
}
