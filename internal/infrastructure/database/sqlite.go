package database

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

// Open opens a SQLite database at path and applies schema migrations.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS guild_settings (
			guild_id     TEXT PRIMARY KEY,
			auto_role_id TEXT NOT NULL DEFAULT ''
		)
	`); err != nil {
		return err
	}

	// Add prefix column if it doesn't exist yet (SQLite has no ADD COLUMN IF NOT EXISTS).
	if _, err := db.Exec(`ALTER TABLE guild_settings ADD COLUMN prefix TEXT NOT NULL DEFAULT ''`); err != nil {
		// Ignore "duplicate column name" — migration already applied.
		if !isDuplicateColumnErr(err) {
			return err
		}
	}

	return nil
}

func isDuplicateColumnErr(err error) bool {
	return err != nil && len(err.Error()) >= 22 && err.Error()[:22] == "duplicate column name:"
}
