package browser

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type firefoxVariant struct {
	Browser string
	Label   string
	Windows string
	Darwin  string
	Linux   string
}

// knownFirefoxVariants is the lookup table of supported Firefox-based browsers.
// To add a new variant, append an entry here — no other changes needed.
// All variants use the same profiles.ini format and places.sqlite schema.
var knownFirefoxVariants = []firefoxVariant{
	{
		Browser: "firefox",
		Label:   "Firefox",
		Windows: `Mozilla\Firefox`,
		Darwin:  "Firefox",
		Linux:   "firefox",
	},
	{
		Browser: "zen",
		Label:   "Zen",
		Windows: `Zen Browser`,
		Darwin:  "Zen Browser",
		Linux:   "zen",
	},
	{
		Browser: "librewolf",
		Label:   "LibreWolf",
		Windows: `LibreWolf`,
		Darwin:  "LibreWolf",
		Linux:   "librewolf",
	},
	{
		Browser: "floorp",
		Label:   "Floorp",
		Windows: `Floorp`,
		Darwin:  "Floorp",
		Linux:   "floorp",
	},
}

// discoverFirefox scans all known Firefox variant locations and returns
// every profile that has a readable places.sqlite file.
func discoverFirefox() ([]DetectedSource, error) {
	var sources []DetectedSource

	for _, variant := range knownFirefoxVariants {
		base := firefoxBasePath(variant.Windows, variant.Darwin, variant.Linux)
		if base == "" || !exists(base) {
			continue
		}

		profiles, err := parseProfilesIni(base)
		if err != nil {
			// profiles.ini missing or unreadable — skip this variant
			continue
		}

		for _, profile := range profiles {
			// Profile paths in profiles.ini can be relative or absolute
			var profileDir string
			if filepath.IsAbs(profile.path) {
				profileDir = profile.path
			} else {
				profileDir = filepath.Join(base, profile.path)
			}

			placesPath := filepath.Join(profileDir, "places.sqlite")
			if !exists(placesPath) {
				continue
			}

			// Use the profile name from profiles.ini if available,
			// otherwise fall back to the directory name
			profileName := profile.name
			if profileName == "" {
				profileName = filepath.Base(profileDir)
			}

			sources = append(sources, DetectedSource{
				Browser: variant.Browser,
				Profile: profileName,
				Path:    placesPath,
				Label:   profileLabel(variant.Label, profileName),
			})
		}
	}

	return sources, nil
}

// firefoxProfile holds the parsed data for a single Firefox profile entry.
type firefoxProfile struct {
	name    string // human-readable name from profiles.ini, may be empty
	path    string // relative or absolute path to profile directory
	default_ bool  // whether this is the default profile
}

// parseProfilesIni reads Firefox's profiles.ini and returns all profiles.
// profiles.ini is a simple INI-like format with sections like [Profile0], [Profile1].
//
// Example profiles.ini:
//
//	[Profile0]
//	Name=default-release
//	IsRelative=1
//	Path=Profiles/abc123.default-release
//	Default=1
//
//	[Profile1]
//	Name=work
//	IsRelative=1
//	Path=Profiles/def456.work
func parseProfilesIni(baseDir string) ([]firefoxProfile, error) {
	iniPath := filepath.Join(baseDir, "profiles.ini")

	f, err := os.Open(iniPath)
	if err != nil {
		return nil, fmt.Errorf("open profiles.ini at %q: %w", iniPath, err)
	}
	defer f.Close()

	var profiles []firefoxProfile
	var current *firefoxProfile
	var isRelative bool

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// New section
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			// Save previous profile if it had a path
			if current != nil && current.path != "" {
				profiles = append(profiles, *current)
			}

			section := line[1 : len(line)-1]
			// Only process Profile sections, not [General] or [Install...]
			if strings.HasPrefix(section, "Profile") {
				current = &firefoxProfile{}
				isRelative = false
			} else {
				current = nil
			}
			continue
		}

		if current == nil {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		switch key {
		case "Name":
			current.name = value
		case "Path":
			current.path = value
		case "IsRelative":
			isRelative = value == "1"
		case "Default":
			current.default_ = value == "1"
		}

		// Normalize path separators — profiles.ini uses forward slashes on all
		// platforms but filepath.Join on Windows expects backslashes
		if key == "Path" && isRelative {
			current.path = filepath.FromSlash(current.path)
		}
	}

	// Don't forget the last profile
	if current != nil && current.path != "" {
		profiles = append(profiles, *current)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan profiles.ini: %w", err)
	}

	return profiles, nil
}