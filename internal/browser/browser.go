package browser

import "fmt"

// DetectedSource represents a browser profile found on disk.
// It is not yet saved to the database — it is a candidate presented
// to the user during onboarding or source management.
type DetectedSource struct {
	Browser string // "chrome" | "edge" | "brave" | "arc" | "firefox" | "zen" | "librewolf" | "floorp" | "safari"
	Profile string // profile directory name or "default"
	Path    string // full path to the History or places.sqlite file
	Label   string // human-readable suggestion e.g. "Chrome (Default)"
}

// Discover scans all known browser locations on the current platform
// and returns every profile that has a readable history database.
// Errors from individual browsers are collected and returned alongside
// any results — a failure to read one browser does not prevent others
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