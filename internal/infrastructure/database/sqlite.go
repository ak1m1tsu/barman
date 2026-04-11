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
		db.Close()
		return nil, err
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS guild_settings (
			guild_id     TEXT PRIMARY KEY,
			auto_role_id TEXT NOT NULL DEFAULT ''
		)
	`)
	return err
}
