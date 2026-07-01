package db

import (
	"database/sql"
	"testing"

	dbgen "github.com/zerebos/hindsight/internal/db/generated"
)

func rawURL(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

func TestBucketTrackingCountsTrackedVisits(t *testing.T) {
	rows := []dbgen.GetTrackedVisitsRow{
		// Two known trackers on one visit, weighted by visit_count.
		{RawUrl: rawURL("https://a.com/?utm_source=nl&fbclid=xyz"), Host: "a.com", VisitCount: 2},
		// Single tracker.
		{RawUrl: rawURL("https://b.com/?utm_source=ad"), Host: "b.com", VisitCount: 1},
		// raw_url set but NO tracking param (e.g. fragment stripped) -> not tracked.
		{RawUrl: rawURL("https://c.com/page"), Host: "c.com", VisitCount: 5},
		// NULL raw_url -> ignored.
		{RawUrl: sql.NullString{}, Host: "d.com", VisitCount: 9},
	}

	got := BucketTracking(rows, 8)

	// Tracked visits: a.com(2) + b.com(1) = 3. c.com and d.com excluded.
	if got.TrackedVisits != 3 {
		t.Errorf("TrackedVisits = %d, want 3", got.TrackedVisits)
	}

	// utm_source appears on a.com(2) + b.com(1) = 3; fbclid on a.com(2) = 2.
	if len(got.TopParams) != 2 {
		t.Fatalf("len(TopParams) = %d, want 2 (%v)", len(got.TopParams), got.TopParams)
	}
	if got.TopParams[0].Name != "utm_source" || got.TopParams[0].Count != 3 {
		t.Errorf("TopParams[0] = %+v, want {utm_source 3}", got.TopParams[0])
	}
	if got.TopParams[1].Name != "fbclid" || got.TopParams[1].Count != 2 {
		t.Errorf("TopParams[1] = %+v, want {fbclid 2}", got.TopParams[1])
	}

	// Top tracked domain is a.com (2), then b.com (1).
	if len(got.TopDomains) != 2 || got.TopDomains[0].Host != "a.com" || got.TopDomains[0].Count != 2 {
		t.Errorf("TopDomains = %+v, want a.com first with 2", got.TopDomains)
	}
}

func TestBucketTrackingTopNBounds(t *testing.T) {
	rows := []dbgen.GetTrackedVisitsRow{
		{RawUrl: rawURL("https://a.com/?utm_source=1&utm_medium=2&utm_campaign=3&gclid=4"), Host: "a.com", VisitCount: 1},
	}
	got := BucketTracking(rows, 2)
	if len(got.TopParams) != 2 {
		t.Errorf("len(TopParams) with topN=2 = %d, want 2", len(got.TopParams))
	}
}

func TestBucketTrackingEmpty(t *testing.T) {
	got := BucketTracking(nil, 8)
	if got.TrackedVisits != 0 || len(got.TopParams) != 0 || len(got.TopDomains) != 0 {
		t.Errorf("empty input produced non-zero stats: %+v", got)
	}
}
