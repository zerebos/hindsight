package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// migrate runs any outstanding migrations against the database in version order.
// It creates the schema_migrations tracking table on first run if it doesn't
// exist, then applies any migration files that haven't been recorded yet.
//
// Migration files must be named NNN_description.sql where NNN is a
// zero-padded integer (e.g. 001_initial_schema.sql). They are applied in
// ascending numeric order. Each migration runs in its own transaction —
// a failure rolls back that migration and halts further execution, leaving
// the database at the last successfully applied version.
func migrate(db *sql.DB) error {
	if err := ensureMigrationsTable(db); err != nil {
		return fmt.Errorf("ensure migrations table: %w", err)
	}

	applied, err := appliedMigrations(db)
	if err != nil {
		return fmt.Errorf("get applied migrations: %w", err)
	}

	pending, err := pendingMigrations(applied)
	if err != nil {
		return fmt.Errorf("get pending migrations: %w", err)
	}

	for _, name := range pending {
		if err := applyMigration(db, name); err != nil {
			return fmt.Errorf("apply migration %q: %w", name, err)
		}
	}

	return nil
}

// ensureMigrationsTable creates the schema_migrations table if it doesn't
// exist. This is the one table we create outside the migration system itself.
func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    TEXT    NOT NULL PRIMARY KEY,
			applied_at INTEGER NOT NULL
		)
	`)
	return err
}

// appliedMigrations returns the set of migration versions already recorded
// in schema_migrations.
func appliedMigrations(db *sql.DB) (map[string]struct{}, error) {
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]struct{})
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = struct{}{}
	}

	return applied, rows.Err()
}

// pendingMigrations returns migration filenames that have not yet been applied,
// sorted in ascending order so they apply in the correct sequence.
func pendingMigrations(applied map[string]struct{}) ([]string, error) {
	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	var pending []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		if _, done := applied[e.Name()]; !done {
			pending = append(pending, e.Name())
		}
	}

	// Sort lexicographically — NNN_ prefix guarantees correct order
	sort.Strings(pending)
	return pending, nil
}

// applyMigration reads a single migration file, executes it in a transaction,
// and records the version in schema_migrations on success.
func applyMigration(db *sql.DB, name string) error {
	data, err := migrationFiles.ReadFile("migrations/" + name)
	if err != nil {
		return fmt.Errorf("read migration file: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.Exec(string(data)); err != nil {
		return fmt.Errorf("execute migration sql: %w", err)
	}

	if _, err = tx.Exec(
		`INSERT INTO schema_migrations (version, applied_at) VALUES (?, unixepoch() * 1000)`,
		name,
	); err != nil {
		return fmt.Errorf("record migration version: %w", err)
	}

	return tx.Commit()
}