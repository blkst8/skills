// Package app wires the application dependencies.
//
// This project uses the private dependency injection pattern: there is no
// global application object. Every layer receives its dependencies explicitly
// through constructors, and cmd/start.go threads the typed bundles
// (Repository, Service) together in order.
//
// Build information is injected here via -ldflags (see the Makefile).
package app

// Build information, injected at build time by the Makefile.
var (
	GitCommit string
	GitTag    string
	BuildDate string
)
