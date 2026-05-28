package db

import (
	"database/sql"
	_ "embed"
	"fmt"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

// Open opens (or creates) the Hindsight SQLite database at the given path,
// applies the schema, and configures pragmas for performance and safety.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := configure(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("configure db: %w", err)
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate db: %w", err)
	}

	return db, nil
}

// configure sets SQLite pragmas. Called once on every open.
func configure(db *sql.DB) error {
	pragmas := []string{
		// WAL mode: reads don't block writes, writes don't block reads.
		// Critical for background sync running alongside UI queries.
		`PRAGMA journal_mode = WAL`,

		// Wait up to 5s if the DB is locked rather than failing immediately.
		`PRAGMA busy_timeout = 5000`,

		// Stricter FK enforcement — SQLite has it off by default.
		`PRAGMA foreign_keys = ON`,

		// Recommended alongside WAL for crash safety.
		`PRAGMA synchronous = NORMAL`,

		// 64MB page cache — helps with large aggregation queries on the
		// visits table which can have millions of rows.
		`PRAGMA cache_size = -64000`,
	}

	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return fmt.Errorf("pragma %q: %w", p, err)
		}
	}

	return nil
}

// migrate runs the embedded schema SQL to create tables and indexes
// if they don't already exist. Safe to call on every open thanks to
// IF NOT EXISTS guards in schema.sql.
func migrate(db *sql.DB) error {
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("run schema: %w", err)
	}
	return nil
}