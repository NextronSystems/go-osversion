//go:build linux
// +build linux

package osversion

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runFixtureDir runs `parse` against every .txt file in testdata/<subdir>,
// comparing the result to the contents of the matching .want sidecar file
// (trailing newline stripped, so editors can safely add one).
// `pathVar` is the package-level path variable the parser reads from; the
// test points it at the fixture file for the duration of each subtest.
func runFixtureDir(t *testing.T, subdir string, pathVar *string, parse func() string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join("testdata", subdir, "*.txt"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(matches) == 0 {
		t.Skipf("no fixtures in testdata/%s", subdir)
	}
	for _, fixture := range matches {
		name := strings.TrimSuffix(filepath.Base(fixture), ".txt")
		t.Run(name, func(t *testing.T) {
			wantPath := strings.TrimSuffix(fixture, ".txt") + ".want"
			wantBytes, err := os.ReadFile(wantPath)
			if err != nil {
				t.Fatalf("read .want sidecar: %v", err)
			}
			want := strings.TrimSuffix(string(wantBytes), "\n")
			defer swap(pathVar, fixture)()
			if got := parse(); got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}

func swap(target *string, value string) func() {
	orig := *target
	*target = value
	return func() { *target = orig }
}

func TestGetFromOSRelease_Fixtures(t *testing.T) {
	runFixtureDir(t, "os-release", &osReleasePath, getFromOSRelease)
}

func TestGetFromDebianVersion_Fixtures(t *testing.T) {
	runFixtureDir(t, "debian_version", &debianVersionPath, getFromDebianVersion)
}

func TestGetFromRedhatRelease_Fixtures(t *testing.T) {
	runFixtureDir(t, "redhat-release", &redhatReleasePath, getFromRedhatRelease)
}

func TestGetFromSuSeRelease_Fixtures(t *testing.T) {
	runFixtureDir(t, "SuSe-release", &suseReleasePath, getFromSuSeRelease)
}
