package browser

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestChromiumProfiles(t *testing.T) {
	base := t.TempDir()

	mkdir := func(name string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Join(base, name), 0o755); err != nil {
			t.Fatalf("mkdir %q: %v", name, err)
		}
	}

	writeFile := func(rel string) {
		t.Helper()
		path := filepath.Join(base, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir parent for %q: %v", rel, err)
		}
		if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
			t.Fatalf("write %q: %v", rel, err)
		}
	}

	// Standard profile names.
	mkdir("Default")
	mkdir("Profile 1")

	// Custom profile should be detected only when Preferences exists.
	mkdir("Work")
	writeFile(filepath.Join("Work", "Preferences"))

	// Directory without Preferences should be ignored.
	mkdir("NotAProfile")

	// Non-directory entries are ignored.
	writeFile("README.txt")

	got, err := chromiumProfiles(base)
	if err != nil {
		t.Fatalf("chromiumProfiles() error = %v", err)
	}

	sort.Strings(got)
	want := []string{"Default", "Profile 1", "Work"}

	if len(got) != len(want) {
		t.Fatalf("chromiumProfiles() len = %d, want %d (got %v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("chromiumProfiles()[%d] = %q, want %q (all: %v)", i, got[i], want[i], got)
		}
	}
}

func TestFriendlyProfileName(t *testing.T) {
	tests := []string{"Default", "Profile 1", "Work"}
	for _, tc := range tests {
		if got := friendlyProfileName(tc); got != tc {
			t.Fatalf("friendlyProfileName(%q) = %q, want %q", tc, got, tc)
		}
	}
}
