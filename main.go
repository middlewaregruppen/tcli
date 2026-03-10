package main

import (
	"fmt"
	"os"

	"github.com/middlewaregruppen/tcli/cmd"
)

func main() {
	if err := cmd.NewDefaultCommand().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
