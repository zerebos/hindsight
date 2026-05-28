package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const appDirName = "hindsight"

// Settings mirrors the structure of config.toml.
type Settings struct {
	General GeneralSettings `toml:"general"`
	Sync    SyncSettings    `toml:"sync"`
}

type GeneralSettings struct {
	LaunchAtLogin   bool   `toml:"launch_at_login"`
	MinimizeToTray  bool   `toml:"minimize_to_tray"`
	Theme           string `toml:"theme"` // "system" | "light" | "dark"
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

// AppDataDir returns the platform-appropriate app data directory and
// ensures it exists.
func AppDataDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config dir: %w", err)
	}

	dir := filepath.Join(base, appDirName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create app dir: %w", err)
	}

	return dir, nil
}

// DBPath returns the path to the main SQLite database.
func DBPath(appDir string) string {
	return filepath.Join(appDir, "data.db")
}

// ConfigPath returns the path to config.toml.
func ConfigPath(appDir string) string {
	return filepath.Join(appDir, "config.toml")
}

// BrowserCacheDir returns the path used for temporary browser DB copies.
func BrowserCacheDir(appDir string) string {
	return filepath.Join(appDir, "browser_cache")
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