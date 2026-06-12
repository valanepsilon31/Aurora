package aurora

import (
	"aurora/internal/logger"
	"aurora/internal/repository"
	"os"
	"path/filepath"
	"strings"

	"github.com/creativeyann17/go-delta/pkg/compress"
	"github.com/dustin/go-humanize"
)

// BackupOutputPath is the base filename for backup archives
const BackupOutputPath = "backup_part.zip"

// Compression presets exposed in settings
const (
	CompressionMax    = "max"    // ZIP level 9: smallest archives, ~2x slower
	CompressionNormal = "normal" // ZIP level 5: ~5% bigger archives, fast
)

// CompressionLevel maps a compression preset to a go-delta ZIP level.
// Unknown or empty presets fall back to normal (the default).
func CompressionLevel(preset string) int {
	if preset == CompressionMax {
		return 9
	}
	return 5
}

// NewBackupOptions creates compress options for backup with standard settings.
// outputDir is the destination directory ("" = current working directory).
func NewBackupOptions(folders []string, threads int, compression, outputDir string, quiet bool) *compress.Options {
	return &compress.Options{
		OutputPath:   filepath.Join(outputDir, BackupOutputPath),
		Files:        folders,
		MaxThreads:   threads,
		Level:        CompressionLevel(compression),
		UseZipFormat: true,
		Quiet:        quiet,
	}
}

// hasPrefixFold reports whether s starts with prefix, ignoring case.
// Filters are case-insensitive to match the search bars' behavior.
func hasPrefixFold(s, prefix string) bool {
	return len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix)
}

// isModFiltered checks if a mod matches the exclusion filters.
// A mod is excluded when its name or path matches a filter, or when EVERY
// collection referencing it matches a filter (a mod still used by at least
// one non-excluded collection is kept).
// Returns (isFiltered, matchedFilter)
func isModFiltered(mod *repository.PenumbraMod, filters []string) (bool, string) {
	for _, filter := range filters {
		if hasPrefixFold(mod.Name, filter) || hasPrefixFold(mod.Path, filter) {
			return true, filter
		}
	}

	if len(mod.Collections) == 0 {
		return false, ""
	}

	matchCollection := func(name string) string {
		for _, filter := range filters {
			if hasPrefixFold(name, filter) {
				return filter
			}
		}
		return ""
	}

	firstMatch := ""
	for _, col := range mod.Collections {
		matched := matchCollection(col.Name)
		if matched == "" {
			return false, "" // used by a non-excluded collection: keep
		}
		if firstMatch == "" {
			firstMatch = matched
		}
	}
	return true, firstMatch
}

// isModIncluded checks if a mod matches any of the given inclusion filters.
// Inclusions pull mods into the backup even when no collection references
// them. Returns (isIncluded, matchedFilter)
func isModIncluded(mod *repository.PenumbraMod, inclusions []string) (bool, string) {
	for _, inclusion := range inclusions {
		if hasPrefixFold(mod.Name, inclusion) ||
			hasPrefixFold(mod.Path, inclusion) {
			return true, inclusion
		}
	}
	return false, ""
}

// inBackupSet decides whether a mod belongs to the backup and why.
// Rule: inclusions always win - a mod matching an inclusion is backed up
// even when no collection references it AND even when an exclusion matches.
// Otherwise: in a collection and not excluded.
// Only the decisive reason is reported: excludedBy when the exclusion drops
// the mod, includedBy when the inclusion is what puts it in the backup.
func inBackupSet(mod *repository.PenumbraMod, filters, inclusions []string) (selected bool, excludedBy, includedBy string) {
	_, excluded := isModFiltered(mod, filters)
	_, included := isModIncluded(mod, inclusions)

	if included != "" {
		// Report the inclusion only when it changes the outcome (no
		// collection, or overriding an exclusion); plain collection mods
		// were backed up anyway and the mark would be noise.
		if len(mod.Collections) == 0 || excluded != "" {
			includedBy = included
		}
		return true, "", includedBy
	}

	if excluded != "" {
		return false, excluded, ""
	}

	return len(mod.Collections) > 0, "", ""
}

// ValidateBackup returns a preview of what will be backed up
func (a *Aurora) ValidateBackup() (BackupValidation, error) {
	repo, err := repository.NewPenumbraRepository(a.cfg)
	if err != nil {
		return BackupValidation{}, err
	}

	items := []BackupItem{}
	var totalSize uint64

	for _, mod := range repo.Mods {
		// Get collection names
		colNames := make([]string, len(mod.Collections))
		for i, col := range mod.Collections {
			colNames[i] = col.Name
		}

		item := BackupItem{
			Mod: Mod{
				Name:        mod.Name,
				Path:        mod.Path,
				Size:        mod.Size,
				SizeHuman:   humanize.Bytes(mod.Size),
				Collections: colNames,
			},
		}

		selected, excludedBy, includedBy := inBackupSet(&mod, a.cfg.Filters, a.cfg.Inclusions)
		item.IsFiltered = excludedBy != ""
		item.FilteredBy = excludedBy
		item.IsIncluded = includedBy != ""
		item.IncludedBy = includedBy

		// List backup candidates: collection mods plus inclusion-matched ones
		if len(mod.Collections) > 0 || item.IsIncluded {
			items = append(items, item)
			if selected {
				totalSize += mod.Size
			}
		}
	}

	// Rough zstd/deflate estimate: ~25% of original (integer math; the old
	// float32 conversion lost precision on large sizes)
	estimated := totalSize / 4

	// Check available disk space in the backup output directory.
	// Require 5% extra margin to avoid running out of space.
	// If detection fails, don't block the backup - report space as unknown.
	availableSpace := uint64(0)
	hasEnoughSpace := true
	spaceKnown := false
	outputDir := a.cfg.Output
	if outputDir == "" {
		outputDir, _ = os.Getwd()
	}
	if outputDir != "" {
		if avail, err := getDiskAvailable(outputDir); err == nil {
			availableSpace = avail
			spaceKnown = true
			requiredSpace := uint64(float64(estimated) * 1.05)
			hasEnoughSpace = availableSpace >= requiredSpace
		} else {
			logger.Warn("Disk space detection failed for %s: %v", outputDir, err)
		}
	}
	availableSpaceHuman := "unknown"
	if spaceKnown {
		availableSpaceHuman = humanize.Bytes(availableSpace)
	}

	validation := BackupValidation{
		Items:               items,
		TotalSize:           totalSize,
		TotalSizeHuman:      humanize.Bytes(totalSize),
		EstimatedSize:       estimated,
		EstimatedSizeHuman:  humanize.Bytes(estimated),
		AvailableSpace:      availableSpace,
		AvailableSpaceHuman: availableSpaceHuman,
		HasEnoughSpace:      hasEnoughSpace,
	}

	logger.Info("Backup validation: %d items, total=%s, estimated=%s, available=%s, hasSpace=%v",
		len(items), validation.TotalSizeHuman, validation.EstimatedSizeHuman,
		validation.AvailableSpaceHuman, validation.HasEnoughSpace)

	return validation, nil
}

// GetBackupFolders returns the list of mod folders that should be backed up
func (a *Aurora) GetBackupFolders() ([]string, error) {
	repo, err := repository.NewPenumbraRepository(a.cfg)
	if err != nil {
		return nil, err
	}

	folders := []string{}
	for _, mod := range repo.Mods {
		selected, excludedBy, includedBy := inBackupSet(&mod, a.cfg.Filters, a.cfg.Inclusions)
		if !selected {
			if excludedBy != "" && len(mod.Collections) > 0 {
				logger.Info("Mod excluded: %s (by %s)", mod.Name, excludedBy)
			}
			continue
		}

		if includedBy != "" {
			logger.Info("Mod included by inclusion filter: %s (by %s)", mod.Name, includedBy)
		}
		folders = append(folders, filepath.Join(a.cfg.Mods.Path, mod.Name))
	}

	logger.Info("GetBackupFolders: %d folders to backup", len(folders))
	return folders, nil
}

// GetFilterMatches reports how many mods each filter pattern matches,
// each pattern evaluated in isolation. Inclusions are counted only where
// they are decisive (mod has no collection, or an exclusion would drop it).
// A count of 0 signals a dead filter.
func (a *Aurora) GetFilterMatches() (FilterMatches, error) {
	repo, err := repository.NewPenumbraRepository(a.cfg)
	if err != nil {
		return FilterMatches{}, err
	}

	result := FilterMatches{
		Filters:    make(map[string]int, len(a.cfg.Filters)),
		Inclusions: make(map[string]int, len(a.cfg.Inclusions)),
	}
	for _, f := range a.cfg.Filters {
		result.Filters[f] = 0
	}
	for _, f := range a.cfg.Inclusions {
		result.Inclusions[f] = 0
	}

	for _, mod := range repo.Mods {
		for _, f := range a.cfg.Filters {
			if matched, _ := isModFiltered(&mod, []string{f}); matched {
				result.Filters[f]++
			}
		}
		for _, f := range a.cfg.Inclusions {
			matched, _ := isModIncluded(&mod, []string{f})
			if !matched {
				continue
			}
			if len(mod.Collections) == 0 {
				result.Inclusions[f]++
				continue
			}
			if excluded, _ := isModFiltered(&mod, a.cfg.Filters); excluded {
				result.Inclusions[f]++ // rescue case
			}
		}
	}

	return result, nil
}
