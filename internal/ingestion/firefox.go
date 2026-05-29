package ingestion

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// readFirefoxHistory opens a copied Firefox places.sqlite database and
// returns all visits newer than sinceFirefoxMicros (a raw Firefox timestamp).
// Pass 0 to read all history.
//
// Firefox stores history across two tables:
//   - moz_places:       one row per distinct URL
//   - moz_historyvisits: one row per individual visit, with precise visit_date
//
// visit_date is in microseconds since Jan 1, 1970 (Unix epoch).
// Firefox does not record visit duration — DurationMs will always be 0.
//
// visit_type values we filter on:
//   1 = TRANSITION_LINK        (clicked a link)
//   2 = TRANSITION_TYPED       (typed in address bar)
//   3 = TRANSITION_BOOKMARK    (visited from bookmark)
//   4 = TRANSITION_EMBED       (embedded, e.g. iframe — excluded)
//   6 = TRANSITION_RELOAD      (page reload — excluded)
//   7 = TRANSITION_REDIRECT_PERMANENT
//   8 = TRANSITION_REDIRECT_TEMPORARY
//
// We exclude embeds (4) and reloads (6) since they inflate visit counts
// without representing intentional navigation.
func readFirefoxHistory(dbPath string, sinceFirefoxMicros int64) ([]RawVisit, error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("open firefox db: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT
			p.url,
			p.title,
			v.visit_date
		FROM moz_historyvisits v
		JOIN moz_places p ON v.place_id = p.id
		WHERE v.visit_date > ?
		  AND v.visit_type NOT IN (4, 6)
		ORDER BY v.visit_date ASC
	`, sinceFirefoxMicros)
	if err != nil {
		return nil, fmt.Errorf("query firefox visits: %w", err)
	}
	defer rows.Close()

	var visits []RawVisit
	for rows.Next() {
		var v RawVisit
		var title sql.NullString

		if err := rows.Scan(&v.URL, &title, &v.VisitedAt); err != nil {
			return nil, fmt.Errorf("scan firefox row: %w", err)
		}

		if title.Valid {
			v.Title = title.String
		}

		// Firefox does not record visit duration
		v.DurationMs = 0

		visits = append(visits, v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate firefox rows: %w", err)
	}

	return visits, nil
}