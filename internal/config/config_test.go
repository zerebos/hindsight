package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDefaults(t *testing.T) {
	got := Defaults()

	if got.General.LaunchAtLogin {
		t.Fatalf("LaunchAtLogin = true, want false")
	}
	if !got.General.MinimizeToTray {
		t.Fatalf("MinimizeToTray = false, want true")
	}
	if got.General.Theme != "default" {
		t.Fatalf("Theme = %q, want %q", got.General.Theme, "default")
	}
	if got.Sync.IntervalMinutes != 30 {
		t.Fatalf("IntervalMinutes = %d, want %d", got.Sync.IntervalMinutes, 30)
	}
	if !got.Sync.SyncOnLaunch || !got.Sync.SyncOnWake {
		t.Fatalf("Sync defaults should be true, got SyncOnLaunch=%v SyncOnWake=%v", got.Sync.SyncOnLaunch, got.Sync.SyncOnWake)
	}
}

func TestPathHelpers(t *testing.T) {
	base := "/tmp/hindsight-test"

	if got := DBPath(base); got != filepath.Join(base, "data.db") {
		t.Fatalf("DBPath() = %q, want %q", got, filepath.Join(base, "data.db"))
	}
	if got := ConfigPath(base); got != filepath.Join(base, "config.toml") {
		t.Fatalf("ConfigPath() = %q, want %q", got, filepath.Join(base, "config.toml"))
	}
	if got := BrowserCacheDir(base); got != filepath.Join(base, "browser_cache") {
		t.Fatalf("BrowserCacheDir() = %q, want %q", got, filepath.Join(base, "browser_cache"))
	}
}

func TestLoadCreatesDefaultsWhenMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got != Defaults() {
		t.Fatalf("Load() defaults mismatch: got %+v want %+v", got, Defaults())
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected config file to be written, stat error: %v", err)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	want := Settings{
		General: GeneralSettings{
			LaunchAtLogin:  true,
			MinimizeToTray: false,
			Theme:          "dark",
		},
		Sync: SyncSettings{
			IntervalMinutes: 5,
			SyncOnLaunch:    false,
			SyncOnWake:      true,
		},
	}

	if err := Save(path, want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got != want {
		t.Fatalf("round-trip mismatch: got %+v want %+v", got, want)
	}
}

func TestLoadNormalizesTheme(t *testing.T) {
	cases := map[string]string{
		"system":  "default", // legacy default value
		"":        "default",
		"bogus":   "default",
		"default": "default",
		"light":   "light",
		"dark":    "dark",
		"amoled":  "amoled",
	}

	for stored, want := range cases {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.toml")

		s := Defaults()
		s.General.Theme = stored
		if err := Save(path, s); err != nil {
			t.Fatalf("Save(theme=%q) error = %v", stored, err)
		}

		got, err := Load(path)
		if err != nil {
			t.Fatalf("Load(theme=%q) error = %v", stored, err)
		}
		if got.General.Theme != want {
			t.Fatalf("Load() theme = %q for stored %q, want %q", got.General.Theme, stored, want)
		}
	}
}

func TestXdgDataBaseLinuxBehavior(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux-specific behavior")
	}

	t.Run("uses XDG_DATA_HOME when set", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", "/tmp/xdg-data")
		base, err := xdgDataBase()
		if err != nil {
			t.Fatalf("xdgDataBase() error = %v", err)
		}
		if base != "/tmp/xdg-data" {
			t.Fatalf("xdgDataBase() = %q, want %q", base, "/tmp/xdg-data")
		}
	})

	t.Run("falls back to HOME/.local/share", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("XDG_DATA_HOME", "")
		t.Setenv("HOME", home)

		base, err := xdgDataBase()
		if err != nil {
			t.Fatalf("xdgDataBase() error = %v", err)
		}

		want := filepath.Join(home, ".local", "share")
		if base != want {
			t.Fatalf("xdgDataBase() = %q, want %q", base, want)
		}
	})
}

func TestConfigDirCreatesDirectory(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux-specific XDG behavior")
	}

	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)

	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error = %v", err)
	}

	want := filepath.Join(base, "hindsight")
	if dir != want {
		t.Fatalf("ConfigDir() = %q, want %q", dir, want)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("ConfigDir() should have created directory: %v", err)
	}
}

func TestDataDirCreatesDirectory(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux-specific XDG behavior")
	}

	base := t.TempDir()
	t.Setenv("XDG_DATA_HOME", base)

	dir, err := DataDir()
	if err != nil {
		t.Fatalf("DataDir() error = %v", err)
	}

	want := filepath.Join(base, "hindsight")
	if dir != want {
		t.Fatalf("DataDir() = %q, want %q", dir, want)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("DataDir() should have created directory: %v", err)
	}
}
