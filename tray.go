package main

import (
	_ "embed"

	"github.com/wailsapp/wails/v3/pkg/application"
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
func setupTray(wailsApp *application.App, window *application.WebviewWindow, runtime *desktopRuntime) {
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
		runtime.triggerSync()
	})

	menu.AddSeparator()

	menu.Add("Quit").OnClick(func(ctx *application.Context) {
		// Route through the runtime so the minimize-to-tray close hook lets
		// the window actually close during shutdown.
		runtime.quit()
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