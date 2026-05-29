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

	fmt.Printf("  found %d profile(s) — registering sources...\n", len(detected))
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

	// Run sync across all sources
	fmt.Println("\nsyncing history...")
	syncer := ingestion.NewSyncer(database, cacheDir)
	results := syncer.SyncAll(ctx)

	totalNew := 0
	totalSkipped := 0
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

	fmt.Printf("\ntotal new visits: %d  skipped: %d\n", totalNew, totalSkipped)

	// Verify row counts
	count, err := queries.CountVisits(ctx)
	if err != nil {
		log.Printf("warning: count query failed: %v", err)
	} else {
		fmt.Printf("total visits in db: %d\n", count)
	}
}