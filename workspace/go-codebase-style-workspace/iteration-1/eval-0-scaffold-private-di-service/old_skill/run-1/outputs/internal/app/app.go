// Package app wires the application dependencies together.
//
// This project uses the private dependency injection pattern: there is no
// global application object. Dependencies are constructed explicitly in
// cmd/start.go via WithDatabase, WithRepository and WithServices, and each
// result is passed into the next constructor.
package app

import "fmt"

// Build information, injected at build time via -ldflags (see the Makefile).
var (
	GitCommit string
	GitTag    string
	BuildDate string
)

const (
	Name    = "invoice-service"
	Version = 1
)

// Banner returns the build information formatted for printing.
func Banner() string {
	return fmt.Sprintf(
		"App: %s v%d\nTag: %s\nCommit: %s\nBuild: %s\n",
		Name, Version, GitTag, GitCommit, BuildDate,
	)
}
