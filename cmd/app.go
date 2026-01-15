package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:     "aurora",
	Short:   "Utils for penumbra and mods operations",
	Version: fmt.Sprintf("%s (commit %s, built %s)", version, commit, date),
}

func init() {
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(penumbraCmd)
	rootCmd.AddCommand(backupCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
