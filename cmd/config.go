package main

import (
	"aurora/pkg/aurora"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure the tool settings (config.json)",
	Run:   runConfigCmd,
}

func init() {
	configCmd.Flags().BoolP("reset", "r", false, "reset the config file with default values")
}

func runConfigCmd(cmd *cobra.Command, args []string) {
	reset, err := cmd.Flags().GetBool("reset")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading flag: %v\n", err)
		return
	}

	app := aurora.NewWithReset(reset)
	cfg := app.GetConfig()

	if reset {
		penumbraPath := prompt(fmt.Sprintf("Enter the path to Penumbra folder (current: %s)", cfg.PenumbraPath))
		modsPath := prompt(fmt.Sprintf("Enter the path to mods folder (current: %s)", cfg.ModsPath))
		app.UpdateConfig(penumbraPath, modsPath)
		cfg = app.GetConfig()
	}

	data := [][]string{
		{"FIELD", "VALUE", "STATUS"},
		{"Penumbra path", abbreviatePath(cfg.PenumbraPath, 100), cfg.Status.PenumbraStatus},
		{"Mods path", abbreviatePath(cfg.ModsPath, 100), cfg.Status.ModsStatus},
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header(data[0])
	table.Bulk(data[1:])
	table.Render()

	if !cfg.Status.Valid {
		fmt.Printf("Current configuration is not valid\nPlease either:\n- run: aurora config --reset\n- edit the config file: %s\n", cfg.ConfigFile)
	} else {
		fmt.Printf("Current configuration is valid\n")
	}
}
