package db

import (
	"context"
	"sort"
	"time"

	dbgen "github.com/zerebos/hindsight/internal/db/generated"
)

// HeatmapCell represents a single populated cell in the activity heatmap.
// Cells with zero visits are not returned -- the caller fills those in for rendering.
type HeatmapCell struct {
	Day         int   // 0=Sunday through 6=Saturday, matches time.Weekday
	Hour        int   // 0-23
	TotalVisits int64
}

// FetchHeatmap runs GetRawVisitsForHeatmap and buckets the results by
// day-of-week and hour-of-day in the given timezone location.
//
// Pass time.Local for the user's current system timezone, or a specific
// *time.Location from time.LoadLocation for an explicit zone. Passing nil
// falls back to time.Local.
//
// Timezone handling is done in Go rather than SQL so that:
//   - DST transitions are handled correctly (a fixed UTC offset is wrong
//     half the year for zones that observe DST)
//   - The bucketing logic is fully testable without a database
func FetchHeatmap(ctx context.Context, q *dbgen.Queries, params dbgen.GetRawVisitsForHeatmapParams, loc *time.Location) ([]HeatmapCell, error) {
	if loc == nil {
		loc = time.Local
	}

	rows, err := q.GetRawVisitsForHeatmap(ctx, params)
	if err != nil {
		return nil, err
	}

	return BucketHeatmap(rows, loc), nil
}

// BucketHeatmap buckets raw visit rows by day-of-week and hour-of-day in
// the given timezone. Only cells with at least one visit are returned.
//
// Accepts the generated row type directly to avoid an unnecessary conversion.
func BucketHeatmap(rows []dbgen.GetRawVisitsForHeatmapRow, loc *time.Location) []HeatmapCell {
	if loc == nil {
		loc = time.Local
	}

	type key struct{ day, hour int }
	counts := make(map[key]int64, 168) // 24 hours * 7 days = 168 possible cells

	for _, row := range rows {
		t := time.UnixMilli(row.VisitedAt).In(loc)
		k := key{
			day:  int(t.Weekday()), // time.Weekday: 0=Sunday, matches strftime %w
			hour: t.Hour(),
		}
		counts[k] += row.VisitCount
	}

	cells := make([]HeatmapCell, 0, len(counts))
	for k, total := range counts {
		cells = append(cells, HeatmapCell{
			Day:         k.day,
			Hour:        k.hour,
			TotalVisits: total,
		})
	}

	sort.Slice(cells, func(i, j int) bool {
		if cells[i].Day != cells[j].Day {
			return cells[i].Day < cells[j].Day
		}
		return cells[i].Hour < cells[j].Hour
	})

	return cells
}