package browser

import (
	"fmt"
	"strings"
)

// Family identifies which browser engine a browser is based on.
// Used by both discovery (for labeling) and ingestion (for dispatch).
type Family string

const (
	FamilyChromium Family = "chromium"
	FamilyFirefox  Family = "firefox"
	FamilySafari   Family = "safari"
	FamilyUnknown  Family = "unknown"
)

// browserFamilies is the single source of truth mapping browser identifiers
// to their engine family. Both discovery and ingestion derive from this.
// To add a new browser: add an entry here AND add a variant entry in the
// appropriate chromium.go or firefox.go variant table for path info.
var browserFamilies = map[string]Family{
	// Chromium-based
	"chrome":  FamilyChromium,
	"edge":    FamilyChromium,
	"brave":   FamilyChromium,
	"arc":     FamilyChromium,
	"vivaldi": FamilyChromium,
	"opera":   FamilyChromium,
	"helium":  FamilyChromium,

	// Firefox-based
	"firefox":   FamilyFirefox,
	"zen":       FamilyFirefox,
	"librewolf": FamilyFirefox,
	"floorp":    FamilyFirefox,

	// Safari
	"safari": FamilySafari,
}

// BrowserFamily returns the engine family for a given browser identifier.
// Returns FamilyUnknown for unrecognized browsers.
func BrowserFamily(browser string) Family {
	if f, ok := browserFamilies[strings.ToLower(browser)]; ok {
		return f
	}
	return FamilyUnknown
}

// DetectedSource represents a browser profile found on disk.
// It is not yet saved to the database -- it is a candidate presented
// to the user during onboarding or source management.
type DetectedSource struct {
	Browser string // e.g. "chrome" | "firefox" | "safari"
	Profile string // profile directory name or "default"
	Path    string // full path to the History or places.sqlite file
	Label   string // human-readable suggestion e.g. "Chrome (Default)"
	Family  Family // engine family, derived from BrowserFamily(Browser)
}

// Discover scans all known browser locations on the current platform
// and returns every profile that has a readable history database.
// Errors from individual browsers are collected and returned alongside
// any results -- a failure to read one browser does not prevent others
// from being discovered.
func Discover() ([]DetectedSource, []error) {
	var sources []DetectedSource
	var errs []error

	collect := func(found []DetectedSource, err error) {
		if err != nil {
			errs = append(errs, err)
		}
		sources = append(sources, found...)
	}

	collect(discoverChromium())
	collect(discoverFirefox())
	collect(discoverSafari())

	return sources, errs
}

// profileLabel builds a consistent human-readable label for a detected source.
// e.g. "Chrome (Default)" or "Firefox (work)"
func profileLabel(browserLabel, profile string) string {
	return fmt.Sprintf("%s (%s)", browserLabel, profile)
}