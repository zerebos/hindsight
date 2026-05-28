package browser

import (
	"os"
	"path/filepath"
	"runtime"
)

// localAppData returns %LOCALAPPDATA% on Windows, falling back to %APPDATA%.
// On other platforms it returns an empty string.
func localAppData() string {
	if p := os.Getenv("LOCALAPPDATA"); p != "" {
		return p
	}
	return os.Getenv("APPDATA")
}

// appData returns %APPDATA% on Windows, empty string on other platforms.
func appData() string {
	return os.Getenv("APPDATA")
}

// homeDir returns the current user's home directory.
func homeDir() string {
	h, _ := os.UserHomeDir()
	return h
}

// chromiumBasePath returns the platform-specific base directory for a
// Chromium variant given its per-OS relative path segments.
// Returns an empty string if the platform is unsupported.
func chromiumBasePath(windows, darwin, linux string) string {
	switch runtime.GOOS {
	case "windows":
		if windows == "" {
			return ""
		}
		return filepath.Join(localAppData(), windows)
	case "darwin":
		if darwin == "" {
			return ""
		}
		return filepath.Join(homeDir(), "Library", "Application Support", darwin)
	case "linux":
		if linux == "" {
			return ""
		}
		// TODO: should follow XDG_CONFIG_HOME if set, but in practice Chromium variants always use ~/.config/<name>
		return filepath.Join(homeDir(), ".config", linux)
	}
	return ""
}

// firefoxBasePath returns the platform-specific base directory for a
// Firefox variant. Firefox uses %APPDATA% on Windows (not %LOCALAPPDATA%),
// unlike Chromium.
func firefoxBasePath(windows, darwin, linux string) string {
	switch runtime.GOOS {
	case "windows":
		if windows == "" {
			return ""
		}
		return filepath.Join(appData(), windows)
	case "darwin":
		if darwin == "" {
			return ""
		}
		return filepath.Join(homeDir(), "Library", "Application Support", darwin)
	case "linux":
		if linux == "" {
			return ""
		}
		// TODO: should follow XDG_CONFIG_HOME if set, but in practice Firefox always uses ~/.config/firefox
		return filepath.Join(homeDir(), ".config", linux)
	}
	return ""
}

// exists returns true if the given path exists on disk.
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}