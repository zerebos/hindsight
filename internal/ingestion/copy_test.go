package ingestion

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeKey(t *testing.T) {
	input := `chrome/work:profile*?"<>| name`
	got := sanitizeKey(input)
	want := "chrome_work_profile_______name"

	if got != want {
		t.Fatalf("sanitizeKey(%q) = %q, want %q", input, got, want)
	}
}

func TestCopyFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.db")
	dst := filepath.Join(dir, "dest.db")

	if err := os.WriteFile(src, []byte("first"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := os.WriteFile(dst, []byte("old"), 0o644); err != nil {
		t.Fatalf("write existing dest: %v", err)
	}

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile() error = %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if string(got) != "first" {
		t.Fatalf("dest content = %q, want %q", string(got), "first")
	}
}

func TestCopyBrowserDBAndCleanup(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	cacheDir := filepath.Join(dir, "cache")

	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}

	srcPath := filepath.Join(srcDir, "History")
	if err := os.WriteFile(srcPath, []byte("main-db"), 0o644); err != nil {
		t.Fatalf("write main db: %v", err)
	}
	if err := os.WriteFile(srcPath+"-wal", []byte("wal-data"), 0o644); err != nil {
		t.Fatalf("write wal: %v", err)
	}
	if err := os.WriteFile(srcPath+"-shm", []byte("shm-data"), 0o644); err != nil {
		t.Fatalf("write shm: %v", err)
	}

	copiedPath, err := copyBrowserDB(srcPath, cacheDir, `chrome/default profile`)
	if err != nil {
		t.Fatalf("copyBrowserDB() error = %v", err)
	}

	if !strings.Contains(copiedPath, "chrome_default_profile") {
		t.Fatalf("copiedPath = %q, want sanitized key segment", copiedPath)
	}

	assertFileContent := func(path string, want string) {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %q: %v", path, err)
		}
		if string(data) != want {
			t.Fatalf("content for %q = %q, want %q", path, string(data), want)
		}
	}

	assertFileContent(copiedPath, "main-db")
	assertFileContent(copiedPath+"-wal", "wal-data")
	assertFileContent(copiedPath+"-shm", "shm-data")

	if err := cleanupBrowserDB(copiedPath); err != nil {
		t.Fatalf("cleanupBrowserDB() error = %v", err)
	}

	if _, err := os.Stat(filepath.Dir(copiedPath)); !os.IsNotExist(err) {
		t.Fatalf("copied directory should be removed, got err=%v", err)
	}
}
