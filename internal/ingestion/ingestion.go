package ingestion

// RawVisit is a browser-agnostic representation of a single visit row
// as read directly from a source browser database. Timestamps are still
// in the source browser's native format at this stage — conversion to
// unix milliseconds happens in the normalization step.
type RawVisit struct {
	URL      string
	Title    string
	// VisitedAt holds the raw timestamp in the source browser's native
	// format. Use the appropriate conversion function from normalize.go:
	//   Chromium: microseconds since Jan 1, 1601  → ChromiumTimeToUnixMs()
	//   Firefox:  microseconds since Jan 1, 1970  → FirefoxTimeToUnixMs()
	//   Safari:   seconds since Jan 1, 2001       → SafariTimeToUnixMs()
	VisitedAt  int64
	DurationMs int64 // 0 if not available
}

// SyncResult holds the outcome of a single source sync run.
type SyncResult struct {
	SourceID  int64
	NewVisits int           // rows successfully written
	Skipped   int           // rows that failed normalization and were skipped
	Error     error         // non-nil = source-level failure, sync did not complete
}