package main

import (
	"fmt"
	"os"

	"github.com/tuna-os/bluefin-cli/cmd"
	"github.com/tuna-os/bluefin-cli/internal/config"
)

func main() {
	if err := config.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize configuration: %v\n", err)
	}

	// fang prints styled errors itself; just set the exit code.
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
