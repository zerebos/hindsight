package browser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseProfilesIni(t *testing.T) {
	base := t.TempDir()
	absProfile := filepath.Join(base, "absolute.profile")

	contents := "" +
		"[General]\n" +
		"StartWithLastProfile=1\n" +
		"\n" +
		"[Profile0]\n" +
		"Name=default-release\n" +
		"IsRelative=1\n" +
		"Path=Profiles/abc123.default-release\n" +
		"Default=1\n" +
		"\n" +
		"[Install4F96D1932A9F858E]\n" +
		"Default=Profiles/abc123.default-release\n" +
		"\n" +
		"[Profile1]\n" +
		"Name=work\n" +
		"Path=" + absProfile + "\n" +
		"IsRelative=0\n"

	iniPath := filepath.Join(base, "profiles.ini")
	if err := os.WriteFile(iniPath, []byte(contents), 0o644); err != nil {
		t.Fatalf("write profiles.ini: %v", err)
	}

	got, err := parseProfilesIni(base)
	if err != nil {
		t.Fatalf("parseProfilesIni() error = %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("parseProfilesIni() len = %d, want %d", len(got), 2)
	}

	if got[0].name != "default-release" {
		t.Fatalf("profile[0].name = %q, want %q", got[0].name, "default-release")
	}
	if got[0].path != filepath.FromSlash("Profiles/abc123.default-release") {
		t.Fatalf("profile[0].path = %q, want %q", got[0].path, filepath.FromSlash("Profiles/abc123.default-release"))
	}
	if !got[0].default_ {
		t.Fatalf("profile[0].default_ = false, want true")
	}

	if got[1].name != "work" {
		t.Fatalf("profile[1].name = %q, want %q", got[1].name, "work")
	}
	if got[1].path != absProfile {
		t.Fatalf("profile[1].path = %q, want %q", got[1].path, absProfile)
	}
	if got[1].default_ {
		t.Fatalf("profile[1].default_ = true, want false")
	}
}

func TestParseProfilesIniMissingFile(t *testing.T) {
	_, err := parseProfilesIni(t.TempDir())
	if err == nil {
		t.Fatalf("parseProfilesIni() error = nil, want non-nil for missing profiles.ini")
	}
}
