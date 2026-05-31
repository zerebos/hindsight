package db

import (
	"database/sql"
	"sort"
	"strings"
	"testing"
)

func openTestSQLite(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open(): %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func TestEnsureAndAppliedMigrations(t *testing.T) {
	db := openTestSQLite(t)

	if err := ensureMigrationsTable(db); err != nil {
		t.Fatalf("ensureMigrationsTable() error = %v", err)
	}

	_, err := db.Exec(`INSERT INTO schema_migrations (version, applied_at) VALUES ('0001_initial_schema.sql', 1), ('0002_split_last_synced.sql', 2)`)
	if err != nil {
		t.Fatalf("insert schema_migrations: %v", err)
	}

	applied, err := appliedMigrations(db)
	if err != nil {
		t.Fatalf("appliedMigrations() error = %v", err)
	}

	if len(applied) != 2 {
		t.Fatalf("len(applied) = %d, want 2", len(applied))
	}
	if _, ok := applied["0001_initial_schema.sql"]; !ok {
		t.Fatalf("expected 0001_initial_schema.sql in applied map")
	}
	if _, ok := applied["0002_split_last_synced.sql"]; !ok {
		t.Fatalf("expected 0002_split_last_synced.sql in applied map")
	}
}

func TestPendingMigrationsSortedAndFiltered(t *testing.T) {
	all, err := pendingMigrations(map[string]struct{}{})
	if err != nil {
		t.Fatalf("pendingMigrations(all) error = %v", err)
	}
	if len(all) == 0 {
		t.Fatalf("pendingMigrations(all) returned no migrations")
	}

	sorted := append([]string(nil), all...)
	sort.Strings(sorted)
	for i := range all {
		if all[i] != sorted[i] {
			t.Fatalf("pending migrations not sorted: got %v", all)
		}
		if !strings.HasSuffix(all[i], ".sql") {
			t.Fatalf("migration %q missing .sql suffix", all[i])
		}
	}

	applied := map[string]struct{}{all[0]: {}}
	filtered, err := pendingMigrations(applied)
	if err != nil {
		t.Fatalf("pendingMigrations(filtered) error = %v", err)
	}

	if len(filtered) != len(all)-1 {
		t.Fatalf("len(filtered) = %d, want %d", len(filtered), len(all)-1)
	}
	for _, name := range filtered {
		if name == all[0] {
			t.Fatalf("filtered list should not contain applied migration %q", all[0])
		}
	}
}

func TestMigrateAppliesAllAndIsIdempotent(t *testing.T) {
	db := openTestSQLite(t)

	all, err := pendingMigrations(map[string]struct{}{})
	if err != nil {
		t.Fatalf("pendingMigrations() error = %v", err)
	}

	if err := migrate(db); err != nil {
		t.Fatalf("migrate() first run error = %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count schema_migrations after first run: %v", err)
	}
	if count != len(all) {
		t.Fatalf("applied migration count = %d, want %d", count, len(all))
	}

	if err := migrate(db); err != nil {
		t.Fatalf("migrate() second run error = %v", err)
	}

	var countAfterSecondRun int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&countAfterSecondRun); err != nil {
		t.Fatalf("count schema_migrations after second run: %v", err)
	}
	if countAfterSecondRun != count {
		t.Fatalf("second migrate run changed migration count: got %d want %d", countAfterSecondRun, count)
	}
}
