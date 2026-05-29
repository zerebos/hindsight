package ingestion

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// readChromiumHistory opens a copied Chromium History SQLite database and
// returns all visits newer than sinceChromiumMicros (a raw Chromium timestamp).
// Pass 0 to read all history.
//
// Chromium stores history across two tables:
//   - urls:   one row per distinct URL, with aggregate visit_count
//   - visits: one row per individual visit, with precise visit_time
//
// We always read from the visits table joined to urls so we get individual
// timestamps rather than collapsed aggregates. visit_time is in microseconds
// since Jan 1, 1601 (Windows FILETIME epoch).
//
// duration is stored in microseconds in Chromium. We convert to milliseconds.
// A duration of 0 means the visit duration was not recorded.
func readChromiumHistory(dbPath string, sinceChromiumMicros int64) ([]RawVisit, error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("open chromium db: %w", err)
	}
	defer db.Close()

	// Open read-only — we never write to source databases
	rows, err := db.Query(`
		SELECT
			u.url,
			u.title,
			v.visit_time,
			v.visit_duration
		FROM visits v
		JOIN urls u ON v.url = u.id
		WHERE v.visit_time > ?
		ORDER BY v.visit_time ASC
	`, sinceChromiumMicros)
	if err != nil {
		return nil, fmt.Errorf("query chromium visits: %w", err)
	}
	defer rows.Close()

	var visits []RawVisit
	for rows.Next() {
		var v RawVisit
		var durationMicros int64

		if err := rows.Scan(&v.URL, &v.Title, &v.VisitedAt, &durationMicros); err != nil {
			return nil, fmt.Errorf("scan chromium row: %w", err)
		}

		// Convert duration from microseconds to milliseconds.
		// Chromium records 0 when duration is unknown.
		v.DurationMs = durationMicros / 1000

		visits = append(visits, v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chromium rows: %w", err)
	}

	return visits, nil
}