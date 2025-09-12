package main

import (
	"fmt"
	"os"

	"github.com/23prime/gh-download/internal/config"
	"github.com/23prime/gh-download/internal/download"
)

func main() {
	cfg := config.ParseArgs()

	if cfg.Help {
		config.PrintUsage()
		return
	}

	if err := download.DownloadFromRelease(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
