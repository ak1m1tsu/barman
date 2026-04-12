package database

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

// Open opens a SQLite database at path.
// Schema migrations are managed externally via golang-migrate (see migrations/).
func Open(path string) (*sql.DB, error) {
	return sql.Open("sqlite", path)
}
