package main

import (
	"aurora/pkg/aurora"
	"fmt"
	"os"
	"strings"

	"github.com/creativeyann17/go-delta/pkg/compress"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Generate a backup of used mods",
	Run:   runBackupCmd,
}

func init() {
	backupCmd.Flags().BoolP("validate", "v", false, "display list of mods to backup only")
	backupCmd.Flags().IntP("thread", "t", 1, "compress folders concurrently")
}

func runBackupCmd(cmd *cobra.Command, args []string) {
	app := aurora.New()
	if !app.IsConfigValid() {
		cfg := app.GetConfig()
		fmt.Fprintf(os.Stderr, "Configuration is not valid:\n")
		fmt.Fprintf(os.Stderr, "  Penumbra: %s\n", cfg.Status.PenumbraStatus)
		fmt.Fprintf(os.Stderr, "  Mods: %s\n", cfg.Status.ModsStatus)
		fmt.Fprintf(os.Stderr, "\nRun 'aurora config --reset' to fix\n")
		return
	}

	validate, err := cmd.Flags().GetBool("validate")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading validate flag: %v\n", err)
		return
	}

	thread, err := cmd.Flags().GetInt("thread")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading thread flag: %v\n", err)
		return
	}

	validation := app.ValidateBackup()

	// Display table
	data := [][]string{
		{"Mod", "Collections", "Size"},
	}
	for _, item := range validation.Items {
		collections := ""
		if item.IsFiltered {
			collections = fmt.Sprintf("Filtered by: %s", item.FilteredBy)
		} else {
			collections = abbreviatePath(strings.Join(item.Mod.Collections, ", "), 100)
		}
		row := []string{item.Mod.Name, collections, item.Mod.SizeHuman}
		data = append(data, row)
	}

	table := tablewriter.NewTable(os.Stdout)
	table.Header(data[0])
	table.Bulk(data[1:])
	table.Render()

	fmt.Printf("Total initial size: %s\n", validation.TotalSizeHuman)
	fmt.Printf("Backup size: %s (estimated)\n", validation.EstimatedSizeHuman)
	fmt.Printf("Available disk space: %s\n", validation.AvailableSpaceHuman)

	if validate {
		return
	}

	if !validation.HasEnoughSpace {
		fmt.Fprintf(os.Stderr, "Error: Not enough disk space for backup\n")
		return
	}

	// Run backup
	if thread <= 0 {
		thread = 1
	}

	folders := app.GetBackupFolders()
	if len(folders) == 0 {
		fmt.Fprintf(os.Stderr, "No mods to backup\n")
		return
	}

	opts := aurora.NewBackupOptions(folders, thread, false)

	progressCb, progress := compress.ProgressBarCallback()
	result, err := compress.Compress(opts, progressCb)

	if progress != nil {
		progress.Wait()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to backup: %v\n", err)
		return
	}

	fmt.Print(compress.FormatSummary(result, opts))
}
