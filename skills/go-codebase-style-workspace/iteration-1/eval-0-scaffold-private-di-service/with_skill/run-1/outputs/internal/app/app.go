// Package app wires the application dependencies using the private
// dependency injection pattern.
//
// There is no global state: every With* constructor returns a typed bundle
// and cmd/start.go chains them (db → repositories → services → usecases →
// server/workers). HTTP handlers and job handlers receive only the usecase
// bundle, which keeps every layer mockable in tests.
package app

import "fmt"

// Build information, injected at build time via -ldflags (see Makefile).
var (
	GitCommit       string
	GitRef          string
	GitTag          string
	BuildDate       string
	CompilerVersion string
)

const (
	Name    = "invoice-service"
	Version = 1
)

// Banner returns a printable build summary.
func Banner() string {
	return fmt.Sprintf(
		"App: %s v%d\nTag: %s\nCommit: %s\nBuild: %s\n",
		Name, Version, GitTag, GitCommit, BuildDate,
	)
}
