package db

import "time"

// HeatmapCell represents a single cell in the activity heatmap.
type HeatmapCell struct {
	Day         int   // 0=Sunday through 6=Saturday
	Hour        int   // 0-23
	TotalVisits int64
}

// HeatmapRow is a raw row returned by GetRawVisitsForHeatmap.
type HeatmapRow struct {
	VisitedAt  int64
	VisitCount int64
}

// BucketHeatmap takes raw visit rows and a timezone location, and returns
// visit counts bucketed by day-of-week and hour-of-day in that timezone.
//
// Doing the bucketing in Go rather than SQL lets us use Go's full timezone
// database (via time.LoadLocation) and avoids sqlc parser limitations with
// named parameters in arithmetic expressions.
//
// Only cells with at least one visit are returned — the caller fills in
// zero-count cells for rendering.
func BucketHeatmap(rows []HeatmapRow, loc *time.Location) []HeatmapCell {
	if loc == nil {
		loc = time.Local
	}

	// bucket key: day*100 + hour (unique for all 168 combinations)
	type key struct{ day, hour int }
	counts := make(map[key]int64, 168)

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

	// Sort by day then hour for consistent output
	sortHeatmapCells(cells)
	return cells
}

// sortHeatmapCells sorts cells in ascending day then hour order.
func sortHeatmapCells(cells []HeatmapCell) {
	for i := 1; i < len(cells); i++ {
		for j := i; j > 0; j-- {
			a, b := cells[j-1], cells[j]
			if a.Day > b.Day || (a.Day == b.Day && a.Hour > b.Hour) {
				cells[j-1], cells[j] = cells[j], cells[j-1]
			}
		}
	}
}