package main

import (
	"context"
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/zerebos/hindsight/internal/app"
	"github.com/zerebos/hindsight/internal/ingestion"
)

//go:embed all:frontend/dist
var assets embed.FS

func init() {
	// Register custom events so the binding generator produces typed
	// JS/TS wrappers for them. The frontend can listen with:
	//   Events.On("sync:complete", (result) => { ... })
	application.RegisterEvent[ingestion.SyncResult]("sync:complete")
	application.RegisterEvent[int64]("sync:started")
	application.RegisterEvent[string]("sync:error")
}

func main() {
	// Initialize the core application logic.
	// This opens the DB, runs migrations, and prepares the syncer.
	hindsight, err := app.New()
	if err != nil {
		log.Fatalf("failed to initialize hindsight: %v", err)
	}
	defer hindsight.Close()

	// Build Wails services wrapping our App.
	sourceSvc := NewSourceService(hindsight)
	dashSvc := NewDashboardService(hindsight)
	searchSvc := NewSearchService(hindsight)
	settingsSvc := NewSettingsService(hindsight)

	// Create the Wails application.
	wailsApp := application.New(application.Options{
		Name:        "Hindsight",
		Description: "Personal browser history intelligence",
		Services: []application.Service{
			application.NewService(sourceSvc),
			application.NewService(dashSvc),
			application.NewService(searchSvc),
			application.NewService(settingsSvc),
			application.NewService(&GreetService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			// On macOS the app should stay alive when the window is closed
			// since it runs as a tray app. The window closing just hides it.
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})

	// Create the main window.
	// Start maximized for first run; tray-first behavior (hidden on subsequent
	// launches) will be wired once we have settings persistence in the UI.
	mainWindow := wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Hindsight",
		Width:  1200,
		Height: 800,
		MinWidth:  800,
		MinHeight: 600,
		URL: "/",
		// Hide the window instead of closing it when the close button
		// is pressed — keeps the app alive in the tray.
		//
		// NOTE: Verify the exact option name for "hide on close" in your
		// version of Wails v3. In some builds this is:
		//   HideOnClose: true
		// or it may be handled via an OnClosing callback. Check:
		// https://v3.wails.io/reference/options#webviewwindowoptions
	})
	_ = mainWindow // used below once tray is wired

	// Set up the system tray.
	//
	// NOTE: The tray API has seen active development in Wails v3.
	// Verify the exact method names against your installed version.
	// The structure below reflects the API as of mid-2025 but may need
	// adjustment. Check: https://v3.wails.io/guide/system-tray
	setupTray(wailsApp, mainWindow, hindsight)

	// Run sync on launch if configured.
	cfg := hindsight.GetSettings()
	if cfg.Sync.SyncOnLaunch {
		go func() {
			results := hindsight.SyncAll(context.Background())
			for _, r := range results {
				if r.Error != nil {
					wailsApp.Event.Emit("sync:error", r.Error.Error())
				} else {
					wailsApp.Event.Emit("sync:complete", r)
				}
			}
		}()
	}

	// Run the application. Blocks until the app exits.
	if err := wailsApp.Run(); err != nil {
		log.Fatal(err)
	}
}