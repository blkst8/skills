// Package main is the entry point of the client service.
package main

import (
	_ "go.uber.org/automaxprocs"

	"github.com/blkst8/client-service/cmd"
)

func main() {
	cmd.Execute()
}
