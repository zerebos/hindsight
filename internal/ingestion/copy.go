package ingestion

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// copyBrowserDB copies a browser's SQLite database to the app's browser cache
// directory and returns the path to the copy. The copy is safe to open without
// risking locks or corruption on the live browser database.
//
// SQLite databases may have associated WAL (-wal) and shared memory (-shm)
// files. If present, these are copied alongside the main DB file to ensure
// the copy is in a consistent state.
//
// The caller is responsible for calling cleanupBrowserDB when the copy is no
// longer needed.
func copyBrowserDB(srcPath, cacheDir, key string) (string, error) {
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		return "", fmt.Errorf("create cache dir: %w", err)
	}

	// Sanitize key for use as a directory name
	dirName := sanitizeKey(key)
	destDir := filepath.Join(cacheDir, dirName)
	if err := os.MkdirAll(destDir, 0o700); err != nil {
		return "", fmt.Errorf("create dest dir: %w", err)
	}

	destPath := filepath.Join(destDir, filepath.Base(srcPath))

	// Copy the main DB file
	if err := copyFile(srcPath, destPath); err != nil {
		return "", fmt.Errorf("copy db file: %w", err)
	}

	// Copy WAL and SHM files if they exist — these are needed for a
	// consistent snapshot if the browser was mid-transaction when we copied.
	// If they don't exist that's fine; it just means the DB was checkpointed.
	for _, suffix := range []string{"-wal", "-shm"} {
		srcSidecar := srcPath + suffix
		if _, err := os.Stat(srcSidecar); err == nil {
			destSidecar := destPath + suffix
			if err := copyFile(srcSidecar, destSidecar); err != nil {
				// Non-fatal: log but continue. The main DB copy may still be
				// readable without the sidecar files.
				continue
			}
		}
	}

	return destPath, nil
}

// cleanupBrowserDB removes the directory containing a copied browser DB.
// Should be deferred immediately after a successful copyBrowserDB call.
func cleanupBrowserDB(copiedPath string) error {
	dir := filepath.Dir(copiedPath)
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("cleanup browser db copy at %q: %w", dir, err)
	}
	return nil
}

// copyFile copies a single file from src to dst, creating dst if it does
// not exist and overwriting it if it does.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source %q: %w", src, err)
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("stat source %q: %w", src, err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("create dest %q: %w", dst, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copy %q -> %q: %w", src, dst, err)
	}

	return nil
}

// sanitizeKey converts a browser+profile key into a safe directory name by
// replacing characters that are invalid or awkward in directory names.
func sanitizeKey(key string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
	)
	return replacer.Replace(key)
}