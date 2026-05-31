package ingestion

import "testing"

func TestToUnixMs(t *testing.T) {
	tests := []struct {
		name       string
		browser    string
		raw        int64
		wantUnixMs int64
	}{
		{name: "chromium", browser: "chrome", raw: 13351223880000000, wantUnixMs: 1706750280000},
		{name: "firefox", browser: "firefox", raw: 1706750280000000, wantUnixMs: 1706750280000},
		{name: "safari", browser: "safari", raw: 1706750280000, wantUnixMs: 1706750280000},
		{name: "unknown defaults to chromium", browser: "unknown", raw: 13351223880000000, wantUnixMs: 1706750280000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toUnixMs(tt.browser, tt.raw); got != tt.wantUnixMs {
				t.Fatalf("toUnixMs(%q, %d) = %d, want %d", tt.browser, tt.raw, got, tt.wantUnixMs)
			}
		})
	}
}

func TestReadSourceUnsupportedBrowser(t *testing.T) {
	var s Syncer

	_, err := s.readSource("not-a-browser", "/tmp/history", 0)
	if err == nil {
		t.Fatalf("readSource() error = nil, want non-nil")
	}
	if got := err.Error(); got != "unsupported browser \"not-a-browser\" (family unknown)" {
		t.Fatalf("readSource() error = %q, want %q", got, "unsupported browser \"not-a-browser\" (family unknown)")
	}
}
