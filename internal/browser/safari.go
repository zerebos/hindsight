package browser

import (
	"path/filepath"
	"runtime"
)

// discoverSafari looks for Safari's History.db on macOS.
// Safari is only available on macOS — this returns nothing on other platforms.
//
// Note: accessing ~/Library/Safari/ requires explicit user permission via
// macOS TCC (Transparency, Consent, and Control). The app will need to
// prompt for this during onboarding. If the file exists but is unreadable
// due to permissions, the copy-then-read step in ingestion will surface
// the error at sync time rather than discovery time.
func discoverSafari() ([]DetectedSource, error) {
	if runtime.GOOS != "darwin" {
		return nil, nil
	}

	home := homeDir()
	if home == "" {
		return nil, nil
	}

	historyPath := filepath.Join(home, "Library", "Safari", "History.db")
	if !exists(historyPath) {
		return nil, nil
	}

	return []DetectedSource{
		{
			Browser: "safari",
			Profile: "default",
			Path:    historyPath,
			Label:   "Safari",
		},
	}, nil
}