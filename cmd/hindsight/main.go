package main

import (
	"fmt"
	"log"
	"os"

	"github.com/zerebos/hindsight/internal/config"
	"github.com/zerebos/hindsight/internal/db"
)

func main() {
	// Resolve app data directory
	appDir, err := config.AppDataDir()
	if err != nil {
		log.Fatalf("failed to resolve app data dir: %v", err)
	}
	fmt.Printf("app data dir: %s\n", appDir)

	// Load or create config
	cfg, err := config.Load(config.ConfigPath(appDir))
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	fmt.Printf("sync interval: %d minutes\n", cfg.Sync.IntervalMinutes)

	// Open database
	dbPath := config.DBPath(appDir)
	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()
	fmt.Printf("database ready: %s\n", dbPath)

	// Ensure browser cache dir exists
	cacheDir := config.BrowserCacheDir(appDir)
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		log.Fatalf("failed to create browser cache dir: %v", err)
	}
	fmt.Printf("browser cache dir: %s\n", cacheDir)

	fmt.Println("hindsight ready")
}