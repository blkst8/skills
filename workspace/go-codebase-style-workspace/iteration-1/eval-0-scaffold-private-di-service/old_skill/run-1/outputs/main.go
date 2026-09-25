// Command invoice-service is the entry point for the invoice service.
//
// It delegates to the cobra commands declared in the cmd package.
package main

import (
	_ "go.uber.org/automaxprocs"

	"github.com/blkst8/invoice-service/cmd"
)

func main() {
	cmd.Execute()
}
