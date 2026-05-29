package ingestion

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	dbgen "github.com/zerebos/hindsight/internal/db/generated"
)

// RawVisit is a browser-agnostic representation of a single visit row
// as read directly from a source browser database. Timestamps are still
// in the source browser's native format at this stage — conversion to
// unix milliseconds happens in the normalization step.
// Exception: Safari timestamps are converted to unix ms in the reader
// since RawVisit.VisitedAt is int64 and Safari uses float64 natively.
type RawVisit struct {
	URL   string
	Title string
	// VisitedAt holds the raw timestamp in the source browser's native format
	// EXCEPT for Safari where it holds unix milliseconds already.
	//   Chromium: microseconds since Jan 1, 1601  → ChromiumTimeToUnixMs()
	//   Firefox:  microseconds since Jan 1, 1970  → FirefoxTimeToUnixMs()
	//   Safari:   already unix ms (converted in reader)
	VisitedAt  int64
	DurationMs int64 // 0 if not available
}

// SyncResult holds the outcome of a single source sync run.
type SyncResult struct {
	SourceID  int64
	NewVisits int   // rows successfully written
	Skipped   int   // rows that failed normalization and were skipped
	Error     error // non-nil = source-level failure, sync did not complete
}

// Syncer handles ingestion for all configured sources.
type Syncer struct {
	db       *sql.DB
	queries  *dbgen.Queries
	cacheDir string
}

// NewSyncer creates a Syncer using the provided database connection and
// browser cache directory.
func NewSyncer(database *sql.DB, cacheDir string) *Syncer {
	return &Syncer{
		db:       database,
		queries:  dbgen.New(database),
		cacheDir: cacheDir,
	}
}

// SyncSource runs a full ingestion cycle for a single source:
//  1. Determine the incremental checkpoint from last_visit_seen
//  2. Copy the source DB to the cache directory
//  3. Read raw visits newer than the checkpoint
//  4. Normalize and write each visit to the internal DB
//  5. Update last_synced_at (wall clock) and last_visit_seen (checkpoint)
func (s *Syncer) SyncSource(ctx context.Context, source dbgen.Source) SyncResult {
	result := SyncResult{SourceID: source.ID}

	// Step 1: determine the incremental checkpoint from last_visit_seen.
	// last_visit_seen holds the newest visit timestamp we've ingested —
	// this is used as the WHERE clause lower bound, not last_synced_at
	// which is only for UI display purposes.
	lastVisitSeenMs := int64(0)
	if source.LastVisitSeen.Valid {
		lastVisitSeenMs = source.LastVisitSeen.Int64
	}

	// Step 2: copy the source DB
	cacheKey := fmt.Sprintf("%s_%s", source.Browser, source.Profile)
	copiedPath, err := copyBrowserDB(source.Path, s.cacheDir, cacheKey)
	if err != nil {
		result.Error = fmt.Errorf("copy browser db: %w", err)
		s.recordError(ctx, source.ID, result.Error)
		return result
	}
	defer func() {
		if err := cleanupBrowserDB(copiedPath); err != nil {
			log.Printf("warning: failed to clean up browser db copy: %v", err)
		}
	}()

	// Step 3: read raw visits from the copied DB
	rawVisits, err := s.readSource(source.Browser, copiedPath, lastVisitSeenMs)
	if err != nil {
		result.Error = fmt.Errorf("read source: %w", err)
		s.recordError(ctx, source.ID, result.Error)
		return result
	}

	now := time.Now().UnixMilli()

	if len(rawVisits) == 0 {
		// Nothing new — still update last_synced_at so the UI shows
		// an accurate "last checked" time, but leave last_visit_seen unchanged
		if err := s.queries.UpdateSourceSyncSuccess(ctx, dbgen.UpdateSourceSyncSuccessParams{
			LastSyncedAt:  sql.NullInt64{Int64: now, Valid: true},
			LastVisitSeen: source.LastVisitSeen, // unchanged
			ID:            source.ID,
		}); err != nil {
			log.Printf("warning: failed to update last_synced_at for source %d: %v", source.ID, err)
		}
		return result
	}

	// Steps 4 + 5: normalize and write in a single transaction
	newVisits, skipped, latestVisitedAt, err := s.writeVisits(ctx, rawVisits, source)
	result.NewVisits = newVisits
	result.Skipped = skipped

	if err != nil {
		result.Error = fmt.Errorf("write visits: %w", err)
		s.recordError(ctx, source.ID, result.Error)
		return result
	}

	// Update both fields:
	//   last_synced_at  = now (wall clock, for UI display)
	//   last_visit_seen = latestVisitedAt (newest visit, for next checkpoint)
	if err := s.queries.UpdateSourceSyncSuccess(ctx, dbgen.UpdateSourceSyncSuccessParams{
		LastSyncedAt:  sql.NullInt64{Int64: now, Valid: true},
		LastVisitSeen: sql.NullInt64{Int64: latestVisitedAt, Valid: true},
		ID:            source.ID,
	}); err != nil {
		log.Printf("warning: failed to update sync success for source %d: %v", source.ID, err)
	}

	return result
}

// readSource dispatches to the correct browser reader based on browser type.
// Returns raw visits with timestamps in native browser format, except Safari
// which converts to unix ms in its reader.
func (s *Syncer) readSource(browser, copiedPath string, lastVisitSeenMs int64) ([]RawVisit, error) {
	// Chromium-based browsers all share the same History schema
	chromiumBrowsers := map[string]struct{}{
		"chrome":  {},
		"edge":    {},
		"brave":   {},
		"arc":     {},
		"vivaldi": {},
		"opera":   {},
		"helium":  {},
	}

	// Firefox-based browsers all share the same places.sqlite schema
	firefoxBrowsers := map[string]struct{}{
		"firefox":   {},
		"zen":       {},
		"librewolf": {},
		"floorp":    {},
	}

	switch {
	case isIn(browser, chromiumBrowsers):
		since := UnixMsToChromiumTime(lastVisitSeenMs)
		return readChromiumHistory(copiedPath, since)

	case isIn(browser, firefoxBrowsers):
		since := UnixMsToFirefoxTime(lastVisitSeenMs)
		return readFirefoxHistory(copiedPath, since)

	case browser == "safari":
		since := UnixMsToSafariTime(lastVisitSeenMs)
		return readSafariHistory(copiedPath, since)

	default:
		return nil, fmt.Errorf("unsupported browser %q", browser)
	}
}

// writeVisits normalizes raw visits and writes them to the internal DB
// in a single transaction. Returns number of new rows written, number
// of skipped rows, the unix timestamp of the latest visit written,
// and any error.
func (s *Syncer) writeVisits(ctx context.Context, rawVisits []RawVisit, source dbgen.Source) (newVisits, skipped int, latestVisitedAt int64, err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	qtx := dbgen.New(tx)
	now := time.Now().UnixMilli()

	for _, raw := range rawVisits {
		// Convert timestamp to unix ms based on browser type
		visitedAtMs := toUnixMs(source.Browser, raw.VisitedAt)

		// Skip visits with invalid timestamps
		if visitedAtMs <= 0 {
			skipped++
			log.Printf("debug: skipping visit with invalid timestamp %d for url %q", raw.VisitedAt, raw.URL)
			continue
		}

		// Normalize URL, strip tracking params, extract domain
		normalizedURL, rawURL := NormalizeURL(raw.URL)
		host := ExtractDomain(normalizedURL)

		// Skip internal browser URLs with no extractable domain
		if host == "" {
			skipped++
			continue
		}

		// Upsert domain — get or create
		domain, err := qtx.UpsertDomain(ctx, dbgen.UpsertDomainParams{
			Host:      host,
			CreatedAt: now,
		})
		if err != nil {
			skipped++
			log.Printf("debug: failed to upsert domain %q: %v", host, err)
			continue
		}

		// Build nullable fields
		var rawURLNull sql.NullString
		if rawURL != "" {
			rawURLNull = sql.NullString{String: rawURL, Valid: true}
		}

		var titleNull sql.NullString
		if raw.Title != "" {
			titleNull = sql.NullString{String: raw.Title, Valid: true}
		}

		var durationNull sql.NullInt64
		if raw.DurationMs > 0 {
			durationNull = sql.NullInt64{Int64: raw.DurationMs, Valid: true}
		}

		// Insert visit — ON CONFLICT DO NOTHING handles dedup silently
		if err := qtx.InsertVisit(ctx, dbgen.InsertVisitParams{
			Url:        normalizedURL,
			RawUrl:     rawURLNull,
			Title:      titleNull,
			DomainID:   domain.ID,
			SourceID:   source.ID,
			VisitedAt:  visitedAtMs,
			DurationMs: durationNull,
			VisitCount: 1,
			CreatedAt:  now,
		}); err != nil {
			skipped++
			log.Printf("debug: failed to insert visit for url %q: %v", normalizedURL, err)
			continue
		}

		newVisits++
		if visitedAtMs > latestVisitedAt {
			latestVisitedAt = visitedAtMs
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, 0, 0, fmt.Errorf("commit transaction: %w", err)
	}

	return newVisits, skipped, latestVisitedAt, nil
}

// toUnixMs converts a raw browser timestamp to unix milliseconds.
// Safari timestamps are already unix ms (converted in the reader).
func toUnixMs(browser string, raw int64) int64 {
	switch browser {
	case "safari":
		return raw // already converted in readSafariHistory
	case "firefox", "zen", "librewolf", "floorp":
		return FirefoxTimeToUnixMs(raw)
	default:
		return ChromiumTimeToUnixMs(raw)
	}
}

// recordError persists a sync error against a source record.
func (s *Syncer) recordError(ctx context.Context, sourceID int64, syncErr error) {
	if err := s.queries.UpdateSourceSyncError(ctx, dbgen.UpdateSourceSyncErrorParams{
		LastError:   sql.NullString{String: syncErr.Error(), Valid: true},
		LastErrorAt: sql.NullInt64{Int64: time.Now().UnixMilli(), Valid: true},
		ID:          sourceID,
	}); err != nil {
		log.Printf("warning: failed to record sync error for source %d: %v", sourceID, err)
	}
}

// isIn checks whether a string key exists in a set.
func isIn(key string, set map[string]struct{}) bool {
	_, ok := set[strings.ToLower(key)]
	return ok
}

// SyncAll runs SyncSource for every source in the database concurrently.
// Results are collected and returned after all sources complete.
func (s *Syncer) SyncAll(ctx context.Context) []SyncResult {
	sources, err := s.queries.GetAllSources(ctx)
	if err != nil {
		return []SyncResult{{Error: fmt.Errorf("get sources: %w", err)}}
	}

	if len(sources) == 0 {
		return nil
	}

	resultCh := make(chan SyncResult, len(sources))

	for _, src := range sources {
		src := src // Capture loop variable
		go func() {
			resultCh <- s.SyncSource(ctx, src)
		}()
	}

	results := make([]SyncResult, 0, len(sources))
	for range sources {
		results = append(results, <-resultCh)
	}

	return results
}