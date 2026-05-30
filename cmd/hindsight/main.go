package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/zerebos/hindsight/internal/browser"
	"github.com/zerebos/hindsight/internal/config"
	"github.com/zerebos/hindsight/internal/db"
	dbgen "github.com/zerebos/hindsight/internal/db/generated"
	"github.com/zerebos/hindsight/internal/ingestion"
)

func main() {
	ctx := context.Background()

	// Resolve platform-appropriate directories
	configDir, err := config.ConfigDir()
	if err != nil {
		log.Fatalf("failed to resolve config dir: %v", err)
	}
	dataDir, err := config.DataDir()
	if err != nil {
		log.Fatalf("failed to resolve data dir: %v", err)
	}

	fmt.Printf("config dir:        %s\n", configDir)
	fmt.Printf("data dir:          %s\n", dataDir)

	// Load or create config
	cfg, err := config.Load(config.ConfigPath(configDir))
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	fmt.Printf("sync interval:     %d minutes\n", cfg.Sync.IntervalMinutes)

	// Open database
	dbPath := config.DBPath(dataDir)
	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()
	fmt.Printf("database ready:    %s\n", dbPath)

	// Ensure browser cache dir exists
	cacheDir := config.BrowserCacheDir(dataDir)
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		log.Fatalf("failed to create browser cache dir: %v", err)
	}
	fmt.Printf("browser cache dir: %s\n\n", cacheDir)

	queries := dbgen.New(database)

	// Discover and register sources
	fmt.Println("scanning for browsers...")
	detected, errs := browser.Discover()
	for _, e := range errs {
		fmt.Printf("  warning: %v\n", e)
	}
	if len(detected) == 0 {
		fmt.Println("  no browsers found")
		return
	}
	fmt.Printf("  found %d profile(s) - registering sources...\n", len(detected))
	for _, d := range detected {
		src, err := queries.UpsertSource(ctx, dbgen.UpsertSourceParams{
			Browser:   d.Browser,
			Profile:   d.Profile,
			Path:      d.Path,
			Label:     sql.NullString{String: d.Label, Valid: true},
			CreatedAt: time.Now().UnixMilli(),
		})
		if err != nil {
			log.Printf("  warning: failed to register source %q: %v", d.Label, err)
			continue
		}
		fmt.Printf("  registered: [%d] %s\n", src.ID, src.Label.String)
	}

	// Sync
	fmt.Println("\nsyncing history...")
	syncer := ingestion.NewSyncer(database, cacheDir)
	results := syncer.SyncAll(ctx)

	totalNew, totalSkipped := 0, 0
	for _, r := range results {
		src, _ := queries.GetAllSources(ctx)
		label := fmt.Sprintf("source %d", r.SourceID)
		for _, s := range src {
			if s.ID == r.SourceID {
				label = s.Label.String
				break
			}
		}
		if r.Error != nil {
			fmt.Printf("  %-40s error: %v\n", label, r.Error)
		} else {
			fmt.Printf("  %-40s new: %-8d skipped: %d\n", label, r.NewVisits, r.Skipped)
		}
		totalNew += r.NewVisits
		totalSkipped += r.Skipped
	}
	fmt.Printf("\ntotal new: %d  skipped: %d\n", totalNew, totalSkipped)

	// ----------------------------------------------------------------
	// Dashboard query validation
	// ----------------------------------------------------------------
	fmt.Println("\n--- dashboard queries ---")

	// All-time filter (0 = no bound)
	allTime := dbgen.GetDashboardStatsParams{StartTime: 0, EndTime: 0}

	// 1. Summary stats
	stats, err := queries.GetDashboardStats(ctx, allTime)
	if err != nil {
		log.Printf("GetDashboardStats error: %v", err)
	} else {
		fmt.Printf("\n[summary stats]\n")
		fmt.Printf("  total visits:   %d\n", stats.TotalVisits)
		fmt.Printf("  unique domains: %d\n", stats.UniqueDomains)
		fmt.Printf("  active days:    %d\n", stats.ActiveDays)
	}

	// 2. Top 10 domains
	topDomains, err := queries.GetTopDomains(ctx, dbgen.GetTopDomainsParams{
		StartTime: 0,
		EndTime:   0,
		Limit:     10,
	})
	if err != nil {
		log.Printf("GetTopDomains error: %v", err)
	} else {
		fmt.Printf("\n[top 10 domains]\n")
		for i, d := range topDomains {
			fmt.Printf("  %2d. %-40s %d visits\n", i+1, d.Host, d.TotalVisits)
		}
	}

	// 3. Visit time series (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30).UnixMilli()
	series, err := queries.GetVisitTimeSeries(ctx, dbgen.GetVisitTimeSeriesParams{
		StartTime: thirtyDaysAgo,
		EndTime:   0,
	})
	if err != nil {
		log.Printf("GetVisitTimeSeries error: %v", err)
	} else {
		fmt.Printf("\n[visit time series - last 30 days] (%d data points)\n", len(series))
		for _, p := range series {
			date := time.UnixMilli(p.Day).UTC().Format("2006-01-02")
			fmt.Printf("  %s  %d\n", date, p.TotalVisits)
		}
	}

	// Timezone sanity check
	zone, offset := time.Now().Zone()
	fmt.Printf("\n[timezone] %s (UTC%+.0f hours) via time.Local\n",
		zone, float64(offset)/3600)

	// 4. Activity heatmap
	cells, err := db.FetchHeatmap(ctx, queries, dbgen.GetRawVisitsForHeatmapParams{
		StartTime: 0,
		EndTime:   0,
	}, time.Local)
	if err != nil {
		log.Printf("FetchHeatmap error: %v", err)
	} else {
		days := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
		fmt.Printf("\n[activity heatmap - all time] (%d populated cells / 168 total)\n", len(cells))
		// Print peak cell per day as a sanity check rather than all 168
		type peak struct {
			hour  int
			total int64
		}
		peaks := make(map[int]peak)
		for _, c := range cells {
			if existing, ok := peaks[c.Day]; !ok || c.TotalVisits > existing.total {
				peaks[c.Day] = peak{c.Hour, c.TotalVisits}
			}
		}
		for day := 0; day <= 6; day++ {
			if p, ok := peaks[day]; ok {
				fmt.Printf("  %s  peak hour: %02d:00  (%d visits)\n", days[day], p.hour, p.total)
			}
		}
	}

	// 5. Browser breakdown
	breakdown, err := queries.GetBrowserBreakdown(ctx, dbgen.GetBrowserBreakdownParams{
		StartTime: 0,
		EndTime:   0,
	})
	if err != nil {
		log.Printf("GetBrowserBreakdown error: %v", err)
	} else {
		fmt.Printf("\n[browser breakdown]\n")
		for _, b := range breakdown {
			label := b.Label.String
			if label == "" {
				label = b.Browser
			}
			fmt.Printf("  %-40s %d visits\n", label, b.TotalVisits)
		}
	}
}