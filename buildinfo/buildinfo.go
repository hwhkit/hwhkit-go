// Package buildinfo exposes git/build metadata for /version and /info endpoints.
//
// Override at link time via -ldflags:
//
//	go build -ldflags "-X github.com/hwhkit/hwhkit-go/buildinfo.GitSHA=$(git rev-parse HEAD) \
//	                  -X github.com/hwhkit/hwhkit-go/buildinfo.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
//	                  -X github.com/hwhkit/hwhkit-go/buildinfo.Version=v0.1.0" .
package buildinfo

import "runtime"

var (
	Version   = "dev"
	GitSHA    = "unknown"
	BuildTime = "unknown"
)

func GoVersion() string { return runtime.Version() }

type Info struct {
	Version   string `json:"version"`
	GitSHA    string `json:"git_sha"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
}

func Get() Info {
	return Info{
		Version:   Version,
		GitSHA:    GitSHA,
		BuildTime: BuildTime,
		GoVersion: GoVersion(),
	}
}
