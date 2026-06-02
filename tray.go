package main

import (
	"context"
	_ "embed"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/zerebos/hindsight/internal/app"
)

//go:embed build/appicon.png
var icon []byte

// setupTray configures the system tray icon and menu.
//
// NOTE: The Wails v3 tray API has been actively developed and method names
// may differ from what's written here. Verify against your installed version:
// https://v3.wails.io/guide/system-tray
//
// Common things that may need adjustment:
//   - wailsApp.NewSystemTray() vs application.NewSystemTray(wailsApp)
//   - SetIcon() argument type ([]byte vs path vs asset reference)
//   - Menu construction API
//   - OnClick / OnRightClick vs SetOnClick
func setupTray(wailsApp *application.App, window *application.WebviewWindow, hindsight *app.App) {
	tray := wailsApp.SystemTray.New()

	tray.SetIcon(icon)
	tray.SetLabel("Hindsight")

	// Build the tray menu
	menu := wailsApp.NewMenu()

	menu.Add("Open Hindsight").OnClick(func(ctx *application.Context) {
		window.Show()
		window.Focus()
	})

	menu.AddSeparator()

	menu.Add("Sync Now").OnClick(func(ctx *application.Context) {
		wailsApp.Event.Emit("sync:started", int64(0)) // 0 = all sources
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
	})

	menu.AddSeparator()

	menu.Add("Quit").OnClick(func(ctx *application.Context) {
		wailsApp.Quit()
	})

	tray.SetMenu(menu)

	// Show the main window when the tray icon is left-clicked.
	// NOTE: Verify the exact callback name in your version.
	// It may be OnClick, SetOnClick, or OnLeftClick.
	tray.OnClick(func() {
		window.Show()
		window.Focus()
	})
}