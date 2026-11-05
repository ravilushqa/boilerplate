package version_test

import (
	"runtime"
	"strings"
	"testing"

	"github.com/ravilushqa/boilerplate/version"
)

func TestCurrent(t *testing.T) {
	t.Parallel()

	info := version.Current()

	if info.Version == "" {
		t.Error("Version is empty")
	}

	if info.GoVersion != runtime.Version() {
		t.Errorf("GoVersion = %q, want %q", info.GoVersion, runtime.Version())
	}

	if !strings.Contains(info.String(), info.Version) {
		t.Errorf("String() = %q does not contain the version %q", info.String(), info.Version)
	}
}

func TestInfoString(t *testing.T) {
	t.Parallel()

	info := version.Info{Version: "v1.2.3", Commit: "abc", Time: "2026-01-02T03:04:05Z", GoVersion: "go1.27.3", Dirty: true}

	const want = "v1.2.3 (commit abc, built 2026-01-02T03:04:05Z, go1.27.3) dirty"
	if got := info.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
