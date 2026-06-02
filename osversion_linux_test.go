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

func writeFixture(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}

func swap(target *string, value string) func() {
	orig := *target
	*target = value
	return func() { *target = orig }
}

func TestGetFromOSRelease(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "trailing newline",
			content: `PRETTY_NAME="Debian GNU/Linux 9 (stretch)"` + "\n",
			want:    "Debian GNU/Linux 9 (stretch)",
		},
		{
			name:    "no trailing newline",
			content: `PRETTY_NAME="CentOS Linux 7 (Core)"`,
			want:    "CentOS Linux 7 (Core)",
		},
		{
			name:    "PRETTY_NAME not first line",
			content: "NAME=\"Ubuntu\"\nPRETTY_NAME=\"Ubuntu 22.04 LTS\"\nID=ubuntu\n",
			want:    "Ubuntu 22.04 LTS",
		},
		{
			name:    "PRETTY_NAME on last line without newline",
			content: "NAME=\"Ubuntu\"\nPRETTY_NAME=\"Ubuntu 22.04 LTS\"",
			want:    "Ubuntu 22.04 LTS",
		},
		{
			name:    "no PRETTY_NAME",
			content: "NAME=\"Foo\"\nID=foo\n",
			want:    "",
		},
		{
			name:    "empty file",
			content: "",
			want:    "",
		},
		{
			name:    "empty PRETTY_NAME value",
			content: `PRETTY_NAME=""` + "\n",
			want:    "",
		},
		{
			name:    "single-char PRETTY_NAME value",
			content: `PRETTY_NAME="x"` + "\n",
			want:    "x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer swap(&osReleasePath, writeFixture(t, "os-release", tt.content))()
			if got := getFromOSRelease(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetFromOSRelease_MissingFile(t *testing.T) {
	defer swap(&osReleasePath, filepath.Join(t.TempDir(), "does-not-exist"))()
	if got := getFromOSRelease(); got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestGetFromDebianVersion(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{name: "trailing newline", content: "9.13\n", want: "Debian 9.13"},
		{name: "no trailing newline", content: "9.13", want: "Debian 9.13"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer swap(&debianVersionPath, writeFixture(t, "debian_version", tt.content))()
			if got := getFromDebianVersion(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetFromRedhatRelease(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{name: "trailing newline", content: "CentOS release 6.10 (Final)\n", want: "CentOS release 6.10 (Final)"},
		{name: "no trailing newline", content: "CentOS release 6.10 (Final)", want: "CentOS release 6.10 (Final)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer swap(&redhatReleasePath, writeFixture(t, "redhat-release", tt.content))()
			if got := getFromRedhatRelease(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetFromSuSeRelease(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{name: "trailing newline", content: "openSUSE 13.2 (Harlequin) (x86_64)\n", want: "openSUSE 13.2 (Harlequin) (x86_64)"},
		{name: "no trailing newline", content: "openSUSE 13.2 (Harlequin) (x86_64)", want: "openSUSE 13.2 (Harlequin) (x86_64)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer swap(&suseReleasePath, writeFixture(t, "SuSe-release", tt.content))()
			if got := getFromSuSeRelease(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
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
