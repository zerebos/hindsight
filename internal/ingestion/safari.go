package ingestion

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// readSafariHistory opens a copied Safari History.db SQLite database and
// returns all visits newer than sinceSafariSecs (a raw Safari timestamp).
// Pass 0 to read all history.
//
// Safari stores history across two tables:
//   - history_items:  one row per distinct URL
//   - history_visits: one row per individual visit, with precise visit_time
//
// visit_time is in seconds since Jan 1, 2001 (Cocoa/Core Data epoch),
// stored as a REAL (float64). Sub-second precision is present but we
// truncate to milliseconds.
//
// The origin field indicates where the visit came from:
//   0 = local (this device)
//   1 = iCloud sync (another device)
//
// We include all origins — iCloud-synced visits from iPhone/iPad are
// legitimate history entries and the cross-device data is a bonus feature.
//
// Safari does not record visit duration — DurationMs will always be 0.
func readSafariHistory(dbPath string, sinceSafariSecs float64) ([]RawVisit, error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("open safari db: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT
			i.url,
			v.title,
			v.visit_time
		FROM history_visits v
		JOIN history_items i ON v.history_item = i.id
		WHERE v.visit_time > ?
		ORDER BY v.visit_time ASC
	`, sinceSafariSecs)
	if err != nil {
		return nil, fmt.Errorf("query safari visits: %w", err)
	}
	defer rows.Close()

	var visits []RawVisit
	for rows.Next() {
		var url   string
		var title sql.NullString
		var visitTimeSecs float64

		if err := rows.Scan(&url, &title, &visitTimeSecs); err != nil {
			return nil, fmt.Errorf("scan safari row: %w", err)
		}

		v := RawVisit{
			URL:       url,
			VisitedAt: SafariTimeToUnixMs(visitTimeSecs),
		}

		if title.Valid {
			v.Title = title.String
		}

		// Safari does not record visit duration
		v.DurationMs = 0

		visits = append(visits, v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate safari rows: %w", err)
	}

	return visits, nil
}