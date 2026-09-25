// Package app holds build information and the dependency wiring bundles
// (database, repositories, external service clients, usecases) that
// cmd/start.go chains together at startup.
package app

import "fmt"

// Build information, injected via ldflags at build time.
var (
	GitCommit       string
	GitRef          string
	GitTag          string
	BuildDate       string
	CompilerVersion string
)

const (
	Name    = "client-service"
	Version = 1
)

// Banner returns the build information as a printable banner.
func Banner() string {
	return fmt.Sprintf(
		"App: %s v%d\nTag: %s\nCommit: %s\nBuild: %s\n",
		Name, Version, GitTag, GitCommit, BuildDate,
	)
}
