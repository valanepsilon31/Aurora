package main

import (
	"aurora/pkg/aurora"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"
)

var penumbraCmd = &cobra.Command{
	Use:   "penumbra",
	Short: "Handle various penumbra operations",
	Run:   runPenumbraCmd,
}

func runPenumbraCmd(cmd *cobra.Command, args []string) {
	app := aurora.New()
	if !app.IsConfigValid() {
		cfg := app.GetConfig()
		fmt.Fprintf(os.Stderr, "Configuration is not valid:\n")
		fmt.Fprintf(os.Stderr, "  Penumbra: %s\n", cfg.Status.PenumbraStatus)
		fmt.Fprintf(os.Stderr, "  Mods: %s\n", cfg.Status.ModsStatus)
		fmt.Fprintf(os.Stderr, "\nRun 'aurora config --reset' to fix\n")
		return
	}

	result := app.GetCollections()

	data := [][]string{
		{"Collection", "Mods"},
	}

	for _, col := range result.Collections {
		var mods strings.Builder
		for _, mod := range col.Mods {
			mods.WriteString(mod.Name)
			mods.WriteString("\n")
		}
		row := []string{col.Name, mods.String()}
		data = append(data, row)
	}

	// Mods without collection
	var modsWithoutCollection strings.Builder
	for _, mod := range result.Mods {
		if len(mod.Collections) == 0 {
			modsWithoutCollection.WriteString(mod.Name)
			modsWithoutCollection.WriteString("\n")
		}
	}
	data = append(data, []string{"(without collection)", modsWithoutCollection.String()})

	footer := fmt.Sprintf("Collections: %d\nMods used: %d/%d\nDisk usage: %s/%s ",
		result.Stats.CollectionCount,
		result.Stats.UsedMods,
		result.Stats.TotalMods,
		result.Stats.UsedDiskSizeHuman,
		result.Stats.TotalDiskSizeHuman)

	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Settings: tw.Settings{
				Separators: tw.Separators{
					BetweenRows: tw.On,
				},
			},
		})),
	)
	table.Header(data[0])
	table.Bulk(data[1:])
	table.Footer([]string{footer, ""})
	table.Render()
}
