package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

const appDirName = "hindsight"

// Settings mirrors the structure of config.toml.
type Settings struct {
	General GeneralSettings `toml:"general"`
	Sync    SyncSettings    `toml:"sync"`
}

type GeneralSettings struct {
	LaunchAtLogin  bool   `toml:"launch_at_login"`
	MinimizeToTray bool   `toml:"minimize_to_tray"`
	Theme          string `toml:"theme"` // "system" | "light" | "dark"
}

type SyncSettings struct {
	IntervalMinutes int  `toml:"interval_minutes"` // 0 = manual only
	SyncOnLaunch    bool `toml:"sync_on_launch"`
	SyncOnWake      bool `toml:"sync_on_wake"`
}

// Defaults returns a Settings with sensible out-of-the-box values.
func Defaults() Settings {
	return Settings{
		General: GeneralSettings{
			LaunchAtLogin:  false,
			MinimizeToTray: true,
			Theme:          "system",
		},
		Sync: SyncSettings{
			IntervalMinutes: 30,
			SyncOnLaunch:    true,
			SyncOnWake:      true,
		},
	}
}

// ConfigDir returns the platform-appropriate directory for config.toml
// and ensures it exists.
//
//   - Windows: %APPDATA%\hindsight
//   - macOS:   ~/Library/Application Support/hindsight
//   - Linux:   $XDG_CONFIG_HOME/hindsight  (~/.config/hindsight)
func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config dir: %w", err)
	}

	dir := filepath.Join(base, appDirName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create config dir: %w", err)
	}

	return dir, nil
}

// DataDir returns the platform-appropriate directory for the SQLite
// database and browser cache, and ensures it exists.
//
//   - Windows: %APPDATA%\hindsight        (same as ConfigDir on Windows)
//   - macOS:   ~/Library/Application Support/hindsight  (same as ConfigDir on macOS)
//   - Linux:   $XDG_DATA_HOME/hindsight   (~/.local/share/hindsight)
//
// On Linux the config and data directories are intentionally separate per
// the XDG Base Directory Specification: config files go in XDG_CONFIG_HOME,
// application data (databases, caches) go in XDG_DATA_HOME.
func DataDir() (string, error) {
	base, err := xdgDataBase()
	if err != nil {
		return "", fmt.Errorf("resolve data dir: %w", err)
	}

	dir := filepath.Join(base, appDirName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create data dir: %w", err)
	}

	return dir, nil
}

// xdgDataBase returns the base directory for application data per platform.
// On Linux this respects $XDG_DATA_HOME with ~/.local/share as the fallback.
// On Windows and macOS it matches UserConfigDir since those platforms don't
// make the config/data distinction.
func xdgDataBase() (string, error) {
	if runtime.GOOS == "linux" {
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			return xdg, nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home dir: %w", err)
		}
		return filepath.Join(home, ".local", "share"), nil
	}

	// Windows and macOS: UserConfigDir is the right base for both
	return os.UserConfigDir()
}

// DBPath returns the path to the main SQLite database.
func DBPath(dataDir string) string {
	return filepath.Join(dataDir, "data.db")
}

// ConfigPath returns the path to config.toml.
func ConfigPath(configDir string) string {
	return filepath.Join(configDir, "config.toml")
}

// BrowserCacheDir returns the path used for temporary browser DB copies.
// Lives in the data directory since it's ephemeral application data,
// not configuration.
func BrowserCacheDir(dataDir string) string {
	return filepath.Join(dataDir, "browser_cache")
}

// Load reads config.toml from the given path. If the file does not exist,
// it returns defaults and writes them to disk.
func Load(path string) (Settings, error) {
	s := Defaults()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := Save(path, s); err != nil {
			return s, fmt.Errorf("write default config: %w", err)
		}
		return s, nil
	}

	if _, err := toml.DecodeFile(path, &s); err != nil {
		return s, fmt.Errorf("decode config: %w", err)
	}

	return s, nil
}

// Save writes settings to the given path as TOML.
func Save(path string, s Settings) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create config file: %w", err)
	}
	defer f.Close()

	if err := toml.NewEncoder(f).Encode(s); err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	return nil
}