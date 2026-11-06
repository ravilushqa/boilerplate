package version_test

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ravilushqa/boilerplate/version"
)

func TestCurrent(t *testing.T) {
	t.Parallel()

	info := version.Current()
	require.NotEmpty(t, info.Version)
	require.Equal(t, runtime.Version(), info.GoVersion)
	require.Contains(t, info.String(), info.Version)
}

func TestInfoString(t *testing.T) {
	t.Parallel()

	info := version.Info{Version: "v1.2.3", Commit: "abc", Time: "2026-01-02T03:04:05Z", GoVersion: "go1.27.3", Dirty: true}
	require.Equal(t, "v1.2.3 (commit abc, built 2026-01-02T03:04:05Z, go1.27.3) dirty", info.String())
}
