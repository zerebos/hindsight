package ingestion

import (
	"net/url"
	"strings"
)

// Timestamp epoch offsets in microseconds.
// All source timestamps are converted to unix milliseconds internally.
const (
	// chromiumEpochOffsetMs is the difference in milliseconds between
	// the Chromium epoch (Jan 1, 1601) and the Unix epoch (Jan 1, 1970).
	// Chromium stores timestamps as microseconds since Jan 1, 1601.
	chromiumEpochOffsetMs = 11644473600000

	// safariEpochOffsetSec is the difference in seconds between the
	// Unix epoch (Jan 1, 1970) and the Cocoa/Core Data epoch (Jan 1, 2001).
	// Safari stores timestamps as seconds since Jan 1, 2001.
	safariEpochOffsetSec = 978307200
)

// ChromiumTimeToUnixMs converts a Chromium timestamp (microseconds since
// Jan 1, 1601) to unix milliseconds.
func ChromiumTimeToUnixMs(chromiumMicros int64) int64 {
	return (chromiumMicros / 1000) - chromiumEpochOffsetMs
}

// FirefoxTimeToUnixMs converts a Firefox timestamp (microseconds since
// Jan 1, 1970) to unix milliseconds.
func FirefoxTimeToUnixMs(firefoxMicros int64) int64 {
	return firefoxMicros / 1000
}

// SafariTimeToUnixMs converts a Safari timestamp (seconds since
// Jan 1, 2001, as a float64) to unix milliseconds.
func SafariTimeToUnixMs(safariSecs float64) int64 {
	return int64((safariSecs+safariEpochOffsetSec)*1000)
}

// knownTrackingParams is the set of query parameters that are stripped
// during URL normalization. These parameters carry tracking information
// but do not affect page identity.
//
// This list intentionally stays small and well-known. The original URL
// is preserved in raw_url for future analytics.
var knownTrackingParams = map[string]struct{}{
	// UTM campaign parameters (Google Analytics)
	"utm_source":   {},
	"utm_medium":   {},
	"utm_campaign": {},
	"utm_term":     {},
	"utm_content":  {},
	"utm_id":       {},

	// Facebook
	"fbclid": {},

	// Google
	"gclid":  {},
	"gclsrc": {},
	"dclid":  {},

	// Microsoft / Bing
	"msclkid": {},

	// Twitter / X
	"twclid": {},

	// HubSpot
	"hsa_acc": {},
	"hsa_cam": {},
	"hsa_grp": {},
	"hsa_ad":  {},
	"hsa_src": {},
	"hsa_tgt": {},
	"hsa_kw":  {},
	"hsa_mt":  {},
	"hsa_net": {},
	"hsa_ver": {},

	// Marketo
	"mkt_tok": {},

	// Mailchimp
	"mc_cid": {},
	"mc_eid": {},

	// Generic click IDs
	"ref":      {},
	"referrer": {},
}

// IsTrackingParam reports whether name is a known tracking parameter that
// normalization strips.
func IsTrackingParam(name string) bool {
	_, ok := knownTrackingParams[name]
	return ok
}

// TrackingParamsIn returns the known tracking parameters present in rawURL's
// query string, in no particular order. Used by analytics to detect which
// visits actually carried trackers (raw_url alone is insufficient, since it
// is also set by non-tracking normalization such as fragment stripping).
func TrackingParamsIn(rawURL string) []string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil
	}
	var found []string
	for param := range u.Query() {
		if IsTrackingParam(param) {
			found = append(found, param)
		}
	}
	return found
}

// NormalizeURL strips known tracking parameters from a URL and normalizes
// its format. Returns the normalized URL and the original URL if
// normalization changed it (empty string if unchanged).
//
// The second return value is suitable for storing in visits.raw_url —
// it is only non-empty when the URL was actually modified.
func NormalizeURL(rawURL string) (normalized string, original string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		// Unparseable URL — return as-is, don't store raw_url
		return rawURL, ""
	}

	// Strip fragment — fragments are client-side only and don't affect
	// page identity for history purposes
	u.Fragment = ""

	q := u.Query()
	changed := false

	for param := range knownTrackingParams {
		if q.Has(param) {
			q.Del(param)
			changed = true
		}
	}

	if changed {
		u.RawQuery = q.Encode()
		return u.String(), rawURL
	}

	// Even if no tracking params were removed, re-encode to normalize
	// formatting (e.g. consistent percent-encoding). Only treat as
	// "changed" if the string actually differs.
	normalized = u.String()
	if normalized != rawURL {
		return normalized, rawURL
	}

	return rawURL, ""
}

// ExtractDomain parses a URL and returns the hostname with www. stripped.
// Returns an empty string if the URL cannot be parsed or has no host.
func ExtractDomain(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	host := u.Hostname()
	if host == "" {
		return ""
	}

	// Strip www. prefix case-insensitively — www.github.com and github.com are
	// the same domain for analytics purposes.
	if strings.HasPrefix(strings.ToLower(host), "www.") {
		host = host[4:]
	}

	return strings.ToLower(host)
}

// UnixMsToChromiumTime converts unix milliseconds to a Chromium timestamp
// (microseconds since Jan 1, 1601). Used to convert last_synced_at back to
// the native format for the incremental sync WHERE clause.
func UnixMsToChromiumTime(unixMs int64) int64 {
	return (unixMs + chromiumEpochOffsetMs) * 1000
}

// UnixMsToFirefoxTime converts unix milliseconds to a Firefox timestamp
// (microseconds since Jan 1, 1970). Used to convert last_synced_at back to
// the native format for the incremental sync WHERE clause.
func UnixMsToFirefoxTime(unixMs int64) int64 {
	return unixMs * 1000
}

// UnixMsToSafariTime converts unix milliseconds to a Safari timestamp
// (seconds since Jan 1, 2001, as float64). Used to convert last_synced_at
// back to the native format for the incremental sync WHERE clause.
func UnixMsToSafariTime(unixMs int64) float64 {
	return float64(unixMs/1000) - safariEpochOffsetSec
}