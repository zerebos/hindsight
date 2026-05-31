package db

import (
	"context"
	"testing"
	"time"

	dbgen "github.com/zerebos/hindsight/internal/db/generated"
)

func TestBucketHeatmapAggregatesAndSorts(t *testing.T) {
	loc := time.UTC

	rows := []dbgen.GetRawVisitsForHeatmapRow{
		{VisitedAt: time.Date(2026, 1, 5, 8, 10, 0, 0, loc).UnixMilli(), VisitCount: 2}, // Monday 08
		{VisitedAt: time.Date(2026, 1, 5, 8, 50, 0, 0, loc).UnixMilli(), VisitCount: 3}, // Monday 08
		{VisitedAt: time.Date(2026, 1, 4, 23, 0, 0, 0, loc).UnixMilli(), VisitCount: 1}, // Sunday 23
		{VisitedAt: time.Date(2026, 1, 5, 7, 5, 0, 0, loc).UnixMilli(), VisitCount: 4},  // Monday 07
	}

	got := BucketHeatmap(rows, loc)

	want := []HeatmapCell{
		{Day: int(time.Sunday), Hour: 23, TotalVisits: 1},
		{Day: int(time.Monday), Hour: 7, TotalVisits: 4},
		{Day: int(time.Monday), Hour: 8, TotalVisits: 5},
	}

	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d (got=%v)", len(got), len(want), got)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %+v, want %+v (all=%+v)", i, got[i], want[i], got)
		}
	}
}

func TestBucketHeatmapNilLocationUsesTimeLocal(t *testing.T) {
	original := time.Local
	time.Local = time.UTC
	t.Cleanup(func() {
		time.Local = original
	})

	rows := []dbgen.GetRawVisitsForHeatmapRow{
		{VisitedAt: time.Date(2026, 1, 6, 9, 0, 0, 0, time.UTC).UnixMilli(), VisitCount: 1}, // Tuesday 09
	}

	got := BucketHeatmap(rows, nil)
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}
	if got[0].Day != int(time.Tuesday) || got[0].Hour != 9 || got[0].TotalVisits != 1 {
		t.Fatalf("unexpected cell: %+v", got[0])
	}
}

func TestFetchHeatmap(t *testing.T) {
	database, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	defer database.Close()

	q := dbgen.New(database)
	ctx := context.Background()
	now := int64(1)

	domain, err := q.UpsertDomain(ctx, dbgen.UpsertDomainParams{Host: "fetch-test.com", CreatedAt: now})
	if err != nil {
		t.Fatalf("UpsertDomain(): %v", err)
	}

	source, err := q.UpsertSource(ctx, dbgen.UpsertSourceParams{
		Browser: "chrome", Profile: "Default", Path: t.TempDir() + "/History", CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("UpsertSource(): %v", err)
	}

	// 1706750280000 ms = 2024-02-01 01:18:00 UTC = Thursday (weekday=4), hour=1
	_, err = q.InsertVisit(ctx, dbgen.InsertVisitParams{
		Url: "https://fetch-test.com/page", DomainID: domain.ID, SourceID: source.ID,
		VisitedAt: 1706750280000, VisitCount: 5, CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("InsertVisit(): %v", err)
	}

	cells, err := FetchHeatmap(ctx, q, dbgen.GetRawVisitsForHeatmapParams{}, time.UTC)
	if err != nil {
		t.Fatalf("FetchHeatmap() error = %v", err)
	}
	if len(cells) != 1 {
		t.Fatalf("len(cells) = %d, want 1", len(cells))
	}
	if cells[0].Day != int(time.Thursday) {
		t.Fatalf("Day = %d, want %d (Thursday)", cells[0].Day, int(time.Thursday))
	}
	if cells[0].Hour != 1 {
		t.Fatalf("Hour = %d, want 1", cells[0].Hour)
	}
	if cells[0].TotalVisits != 5 {
		t.Fatalf("TotalVisits = %d, want 5", cells[0].TotalVisits)
	}
}
