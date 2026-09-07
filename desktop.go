package main

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/zerebos/hindsight/internal/app"
	"github.com/zerebos/hindsight/internal/config"
)

// desktopRuntime wires the persisted settings to actual desktop behavior.
//
// The core App (internal/app) only stores settings — it is shared with the
// CLI and must stay free of any GUI dependency. Everything that makes a
// setting "do something" on the desktop lives here:
//
//   - General.LaunchAtLogin  -> OS autostart registration
//   - General.MinimizeToTray -> hide instead of close on the window's X button
//   - Sync.IntervalMinutes   -> a background ticker that syncs periodically
//   - Sync.SyncOnWake        -> sync when the machine wakes from sleep
//   - Sync.SyncOnLaunch      -> sync once at startup
//
// Settings can change at runtime from the UI; reconcile() is registered as the
// App's settings-change handler so toggles take effect immediately.
type desktopRuntime struct {
	wailsApp  *application.App
	window    *application.WebviewWindow
	hindsight *app.App

	// quitting is set when the user explicitly quits, so the window-closing
	// hook lets the close through instead of hiding to tray.
	quitting atomic.Bool

	// syncing guards against overlapping full syncs triggered from different
	// sources (launch, tray, timer, wake).
	syncing atomic.Bool

	// tickerMu guards tickerStop, which signals the active periodic-sync
	// goroutine to stop. nil when no timer is running (manual-only).
	tickerMu   sync.Mutex
	tickerStop chan struct{}
}

func newDesktopRuntime(wailsApp *application.App, window *application.WebviewWindow, hindsight *app.App) *desktopRuntime {
	return &desktopRuntime{
		wailsApp:  wailsApp,
		window:    window,
		hindsight: hindsight,
	}
}

// start applies the current settings and installs the runtime hooks. Call once
// after the window and tray have been created.
func (d *desktopRuntime) start() {
	cfg := hindsightSettings(d.hindsight)

	// React to future settings changes from the UI.
	d.hindsight.SetSettingsChangeHandler(d.reconcile)

	// Apply current settings.
	d.applyAutostart(cfg.General.LaunchAtLogin)
	d.restartSyncTicker(cfg.Sync.IntervalMinutes)

	// Hide-to-tray: intercept the window close button. Registered as a hook so
	// it runs before (and can cancel) Wails' default close-and-destroy handler.
	d.window.RegisterHook(events.Common.WindowClosing, d.onWindowClosing)

	// Sync when the machine wakes from sleep, if enabled.
	d.wailsApp.Event.OnApplicationEvent(events.Common.SystemDidWake, func(_ *application.ApplicationEvent) {
		if hindsightSettings(d.hindsight).Sync.SyncOnWake {
			d.triggerSync()
		}
	})

	// Sync once on launch, if enabled.
	if cfg.Sync.SyncOnLaunch {
		d.triggerSync()
	}
}

// reconcile applies the effects of a settings change. MinimizeToTray and
// SyncOnWake are read live at event time, so only autostart and the sync
// interval need explicit re-application here.
func (d *desktopRuntime) reconcile(old, updated config.Settings) {
	if old.General.LaunchAtLogin != updated.General.LaunchAtLogin {
		d.applyAutostart(updated.General.LaunchAtLogin)
	}
	if old.Sync.IntervalMinutes != updated.Sync.IntervalMinutes {
		d.restartSyncTicker(updated.Sync.IntervalMinutes)
	}
}

// quit marks the runtime as quitting and tears the application down. Using this
// instead of wailsApp.Quit() directly ensures the close hook does not re-hide
// the window during shutdown.
func (d *desktopRuntime) quit() {
	d.quitting.Store(true)
	d.wailsApp.Quit()
}

// onWindowClosing implements the minimize-to-tray behavior. When enabled, the
// window is hidden and kept alive in the tray; otherwise closing the window
// exits the application.
func (d *desktopRuntime) onWindowClosing(e *application.WindowEvent) {
	if d.quitting.Load() {
		return // a real quit is in progress — let the window close
	}
	if hindsightSettings(d.hindsight).General.MinimizeToTray {
		e.Cancel()
		d.window.Hide()
		return
	}
	// Not minimizing to tray: closing the window should exit the app entirely
	// rather than leaving an orphaned tray icon with no window behind it.
	go d.quit()
}

// applyAutostart registers or unregisters the app to launch at login so the
// OS state matches the setting. Failures are logged but non-fatal — autostart
// is unsupported on some platforms (e.g. server/headless builds).
func (d *desktopRuntime) applyAutostart(enabled bool) {
	if d.wailsApp.Autostart == nil {
		return
	}

	current, err := d.wailsApp.Autostart.IsEnabled()
	if err != nil {
		log.Printf("autostart: check status: %v", err)
		// Fall through and attempt to apply the desired state anyway.
	} else if current == enabled {
		return // already in the desired state
	}

	if enabled {
		err = d.wailsApp.Autostart.Enable()
	} else {
		err = d.wailsApp.Autostart.Disable()
	}
	if err != nil {
		log.Printf("autostart: apply enabled=%v: %v", enabled, err)
	}
}

// restartSyncTicker stops any running periodic-sync goroutine and, if
// intervalMinutes > 0, starts a new one. An interval of 0 means manual-only.
func (d *desktopRuntime) restartSyncTicker(intervalMinutes int) {
	d.tickerMu.Lock()
	defer d.tickerMu.Unlock()

	if d.tickerStop != nil {
		close(d.tickerStop)
		d.tickerStop = nil
	}

	if intervalMinutes <= 0 {
		return
	}

	stop := make(chan struct{})
	d.tickerStop = stop
	interval := time.Duration(intervalMinutes) * time.Minute

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				d.triggerSync()
			}
		}
	}()
}

// triggerSync runs a full sync across all sources in the background, emitting
// the same sync:* events the UI listens for. Concurrent calls are ignored while
// a sync is already in progress so the launch/tray/timer/wake triggers can't
// stack up.
func (d *desktopRuntime) triggerSync() {
	if !d.syncing.CompareAndSwap(false, true) {
		return // a sync is already running
	}

	d.wailsApp.Event.Emit("sync:started", int64(0)) // 0 = all sources
	go func() {
		defer d.syncing.Store(false)
		results := d.hindsight.SyncAll(context.Background())
		for _, r := range results {
			if r.Error != nil {
				d.wailsApp.Event.Emit("sync:error", r.Error.Error())
			} else {
				d.wailsApp.Event.Emit("sync:complete", r)
			}
		}
	}()
}

// hindsightSettings is a tiny helper to keep the call sites terse.
func hindsightSettings(a *app.App) config.Settings {
	return a.GetSettings()
}
