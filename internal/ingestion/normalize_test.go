package ingestion

import "testing"

func TestTimestampConversions(t *testing.T) {
	const expectedUnixMs int64 = 1706750280000

	tests := []struct {
		name string
		got  int64
	}{
		{
			name: "chromium",
			got:  ChromiumTimeToUnixMs(13351223880000000),
		},
		{
			name: "firefox",
			got:  FirefoxTimeToUnixMs(1706750280000000),
		},
		{
			name: "safari",
			got:  SafariTimeToUnixMs(728443080.0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != expectedUnixMs {
				t.Fatalf("got %d, want %d", tt.got, expectedUnixMs)
			}
		})
	}

	chromium := ChromiumTimeToUnixMs(13351223880000000)
	firefox := FirefoxTimeToUnixMs(1706750280000000)
	safari := SafariTimeToUnixMs(728443080.0)

	if chromium != firefox || firefox != safari {
		t.Fatalf("timestamps should match: chromium=%d firefox=%d safari=%d", chromium, firefox, safari)
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name           string
		rawURL         string
		wantNormalized string
		wantOriginal   string
	}{
		{
			name:           "tracking params stripped",
			rawURL:         "https://example.com/path?utm_source=newsletter&utm_medium=email",
			wantNormalized: "https://example.com/path",
			wantOriginal:   "https://example.com/path?utm_source=newsletter&utm_medium=email",
		},
		{
			name:           "no tracking params",
			rawURL:         "https://example.com/path?page=2",
			wantNormalized: "https://example.com/path?page=2",
			wantOriginal:   "",
		},
		{
			name:           "tracking params and legitimate query param",
			rawURL:         "https://example.com/path?page=2&utm_source=newsletter&utm_medium=email",
			wantNormalized: "https://example.com/path?page=2",
			wantOriginal:   "https://example.com/path?page=2&utm_source=newsletter&utm_medium=email",
		},
		{
			name:           "fragment and tracking stripped together",
			rawURL:         "https://example.com/path?page=2&utm_source=newsletter#section",
			wantNormalized: "https://example.com/path?page=2",
			wantOriginal:   "https://example.com/path?page=2&utm_source=newsletter#section",
		},
		{
			name:           "duplicate tracking params removed",
			rawURL:         "https://example.com/path?utm_source=newsletter&utm_source=partner&utm_medium=email",
			wantNormalized: "https://example.com/path",
			wantOriginal:   "https://example.com/path?utm_source=newsletter&utm_source=partner&utm_medium=email",
		},
		{
			name:           "fragment only change",
			rawURL:         "https://example.com/path#section",
			wantNormalized: "https://example.com/path",
			wantOriginal:   "https://example.com/path#section",
		},
		{
			name:           "unparseable url",
			rawURL:         "http://[::1",
			wantNormalized: "http://[::1",
			wantOriginal:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized, original := NormalizeURL(tt.rawURL)
			if normalized != tt.wantNormalized {
				t.Fatalf("normalized = %q, want %q", normalized, tt.wantNormalized)
			}
			if original != tt.wantOriginal {
				t.Fatalf("original = %q, want %q", original, tt.wantOriginal)
			}
		})
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
		want   string
	}{
		{
			name:   "www stripped",
			rawURL: "https://www.github.com/foo",
			want:   "github.com",
		},
		{
			name:   "uppercase www stripped",
			rawURL: "https://WWW.github.com/foo",
			want:   "github.com",
		},
		{
			name:   "plain domain",
			rawURL: "https://github.com/foo",
			want:   "github.com",
		},
		{
			name:   "subdomains preserved",
			rawURL: "https://sub.domain.github.com",
			want:   "sub.domain.github.com",
		},
		{
			name:   "empty url",
			rawURL: "",
			want:   "",
		},
		{
			name:   "unparseable url",
			rawURL: "http://[::1",
			want:   "",
		},
		{
			name:   "localhost with port",
			rawURL: "http://localhost:8080",
			want:   "localhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractDomain(tt.rawURL); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}