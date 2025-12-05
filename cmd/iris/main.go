// Package main is the entry point for the iris CLI.
package main

import (
	"fmt"
	"os"

	"github.com/sivchari/iris/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
