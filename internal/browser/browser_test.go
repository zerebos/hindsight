package browser

import "testing"

func TestBrowserFamily(t *testing.T) {
	tests := []struct {
		name    string
		browser string
		want    Family
	}{
		{name: "chromium browser", browser: "chrome", want: FamilyChromium},
		{name: "firefox browser", browser: "firefox", want: FamilyFirefox},
		{name: "safari browser", browser: "safari", want: FamilySafari},
		{name: "case-insensitive", browser: "BrAvE", want: FamilyChromium},
		{name: "unknown", browser: "not-a-browser", want: FamilyUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BrowserFamily(tt.browser); got != tt.want {
				t.Fatalf("BrowserFamily(%q) = %q, want %q", tt.browser, got, tt.want)
			}
		})
	}
}

func TestProfileLabel(t *testing.T) {
	got := profileLabel("Chrome", "Default")
	if got != "Chrome (Default)" {
		t.Fatalf("profileLabel() = %q, want %q", got, "Chrome (Default)")
	}
}
