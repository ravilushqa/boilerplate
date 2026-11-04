// Package version reports build information of the running binary.
package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
)

// version can be set at build time:
//
//	go build -ldflags "-X github.com/ravilushqa/boilerplate/internal/version.version=v1.2.3"
//
// Without it, the version Go stamps from VCS tags is used (when available).
var version string

// Info describes the build.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Time      string `json:"time"`
	Dirty     bool   `json:"dirty"`
	GoVersion string `json:"go_version"`
}

// String returns a one-line human-readable description.
func (i Info) String() string {
	s := fmt.Sprintf("%s (commit %s, built %s, %s)", i.Version, i.Commit, i.Time, i.GoVersion)
	if i.Dirty {
		s += " dirty"
	}

	return s
}

// Get returns build information. It is computed once.
var Get = sync.OnceValue(func() Info {
	info := Info{
		Version:   "dev",
		Commit:    "unknown",
		Time:      "unknown",
		GoVersion: runtime.Version(),
	}

	bi, ok := debug.ReadBuildInfo()
	if ok {
		if v := bi.Main.Version; v != "" && v != "(devel)" {
			info.Version = v
		}

		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				info.Commit = s.Value
			case "vcs.time":
				info.Time = s.Value
			case "vcs.modified":
				info.Dirty = s.Value == "true"
			}
		}
	}

	if version != "" {
		info.Version = version
	}

	return info
})
