package main

import (
	"fmt"
	"log"
	"os"

	"github.com/zerebos/hindsight/internal/browser"
	"github.com/zerebos/hindsight/internal/config"
	"github.com/zerebos/hindsight/internal/db"
)

func main() {
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

	// Discover installed browsers
	fmt.Println("scanning for browsers...")
	sources, errs := browser.Discover()

	for _, e := range errs {
		fmt.Printf("  warning: %v\n", e)
	}

	if len(sources) == 0 {
		fmt.Println("  no browsers found")
	} else {
		fmt.Printf("  found %d profile(s):\n", len(sources))
		for _, s := range sources {
			fmt.Printf("    [%s] %s\n", s.Browser, s.Label)
			fmt.Printf("           profile: %s\n", s.Profile)
			fmt.Printf("           path:    %s\n", s.Path)
		}
	}
}