package db

import (
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
