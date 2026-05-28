package browser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type chromiumVariant struct {
	Browser string
	Label   string
	Windows string
	Darwin  string
	Linux   string
}

// knownChromiumVariants is the lookup table of supported Chromium-based browsers.
// To add a new variant, append an entry here — no other changes needed.
var knownChromiumVariants = []chromiumVariant{
	{
		Browser: "chrome",
		Label:   "Chrome",
		Windows: `Google\Chrome\User Data`,
		Darwin:  "Google/Chrome",
		Linux:   "google-chrome",
	},
	{
		Browser: "edge",
		Label:   "Edge",
		Windows: `Microsoft\Edge\User Data`,
		Darwin:  "Microsoft Edge",
		Linux:   "microsoft-edge",
	},
	{
		Browser: "brave",
		Label:   "Brave",
		Windows: `BraveSoftware\Brave-Browser\User Data`,
		Darwin:  "BraveSoftware/Brave-Browser",
		Linux:   "BraveSoftware/Brave-Browser",
	},
	{
		Browser: "arc",
		Label:   "Arc",
		Windows: "", // Arc on Windows is very new, skip for now
		Darwin:  "Arc/User Data",
		Linux:   "", // not available on Linux
	},
	{
		Browser: "vivaldi",
		Label:   "Vivaldi",
		Windows: `Vivaldi\User Data`,
		Darwin:  "Vivaldi",
		Linux:   "vivaldi",
	},
	{
		Browser: "opera",
		Label:   "Opera",
		Windows: `Opera Software\Opera Stable`,
		Darwin:  "com.operasoftware.Opera",
		Linux:   "opera",
	},
	{
		Browser: "helium",
		Label:   "Helium",
		Windows: `imput\Helium\User Data`,
		Darwin:  "imput/Helium",
		Linux:   "imput/Helium",
	},
}

// discoverChromium scans all known Chromium variant locations and returns
// every profile that has a readable History file.
func discoverChromium() ([]DetectedSource, error) {
	var sources []DetectedSource

	for _, variant := range knownChromiumVariants {
		base := chromiumBasePath(variant.Windows, variant.Darwin, variant.Linux)
		if base == "" || !exists(base) {
			continue
		}

		profiles, err := chromiumProfiles(base)
		if err != nil {
			// Log but don't fail — other variants should still be scanned
			continue
		}

		for _, profile := range profiles {
			historyPath := filepath.Join(base, profile, "History")
			if !exists(historyPath) {
				continue
			}

			sources = append(sources, DetectedSource{
				Browser: variant.Browser,
				Profile: profile,
				Path:    historyPath,
				Label:   profileLabel(variant.Label, friendlyProfileName(profile)),
			})
		}
	}

	return sources, nil
}

// chromiumProfiles enumerates profile directories within a Chromium user data dir.
// Chromium uses "Default" for the first profile and "Profile N" for subsequent ones.
// It also supports named profiles which appear as directories with a "Preferences"
// file inside.
func chromiumProfiles(userDataDir string) ([]string, error) {
	entries, err := os.ReadDir(userDataDir)
	if err != nil {
		return nil, fmt.Errorf("read chromium user data dir %q: %w", userDataDir, err)
	}

	var profiles []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		name := e.Name()

		// Standard Chromium profile directory names
		if name == "Default" || strings.HasPrefix(name, "Profile ") {
			profiles = append(profiles, name)
			continue
		}

		// Some Chromium variants use non-standard names — check for Preferences
		// file as a signal that it's a valid profile directory
		prefsPath := filepath.Join(userDataDir, name, "Preferences")
		if exists(prefsPath) {
			profiles = append(profiles, name)
		}
	}

	return profiles, nil
}

// friendlyProfileName converts a raw Chromium profile directory name into
// something more readable for display purposes.
// "Default" -> "Default", "Profile 1" -> "Profile 1", custom names pass through.
func friendlyProfileName(profile string) string {
	return profile
}