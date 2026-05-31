package browser

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestChromiumBasePathLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux-specific path behavior")
	}

	home := t.TempDir()
	t.Setenv("HOME", home)

	got := chromiumBasePath("", "", "google-chrome")
	want := filepath.Join(home, ".config", "google-chrome")
	if got != want {
		t.Fatalf("chromiumBasePath() = %q, want %q", got, want)
	}
}

func TestChromiumBasePathEmptySegment(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux-specific path behavior")
	}

	got := chromiumBasePath("", "", "")
	if got != "" {
		t.Fatalf("chromiumBasePath() with empty linux segment = %q, want empty", got)
	}
}

func TestFirefoxBasePathLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux-specific path behavior")
	}

	home := t.TempDir()
	t.Setenv("HOME", home)

	got := firefoxBasePath("", "", "mozilla/firefox")
	want := filepath.Join(home, ".config", "mozilla/firefox")
	if got != want {
		t.Fatalf("firefoxBasePath() = %q, want %q", got, want)
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()

	if !exists(dir) {
		t.Fatalf("exists(%q) = false, want true for existing directory", dir)
	}

	if exists(filepath.Join(dir, "does-not-exist")) {
		t.Fatalf("exists() = true, want false for non-existent path")
	}
}
