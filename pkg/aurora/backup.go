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

// NewBackupOptions creates compress options for backup with standard settings
func NewBackupOptions(folders []string, threads int, quiet bool) *compress.Options {
	return &compress.Options{
		OutputPath:   BackupOutputPath,
		Files:        folders,
		MaxThreads:   threads,
		Level:        9,
		UseZipFormat: true,
		Quiet:        quiet,
	}
}

// isModFiltered checks if a mod matches any of the given filters
// Returns (isFiltered, matchedFilter)
func isModFiltered(mod *repository.PenumbraMod, filters []string) (bool, string) {
	for _, filter := range filters {
		if strings.HasPrefix(mod.Name, filter) ||
			strings.HasPrefix(mod.Path, filter) ||
			(len(mod.Collections) == 1 && strings.HasPrefix(mod.Collections[0].Name, filter)) {
			return true, filter
		}
	}
	return false, ""
}

// ValidateBackup returns a preview of what will be backed up
func (a *Aurora) ValidateBackup() BackupValidation {
	repo := repository.NewPenumbraRepository(a.cfg)

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

		// Check filters
		item.IsFiltered, item.FilteredBy = isModFiltered(&mod, a.cfg.Filters)

		// Only count mods that are in collections and not filtered
		if len(mod.Collections) > 0 {
			items = append(items, item)
			if !item.IsFiltered {
				totalSize += mod.Size
			}
		}
	}

	estimated := uint64(float32(totalSize) * 0.25)

	// Check available disk space in current working directory (where backup.zip is written)
	// Require 5% extra margin to avoid running out of space
	var availableSpace uint64
	var hasEnoughSpace bool
	cwd, err := os.Getwd()
	if err == nil {
		availableSpace, err = getDiskAvailable(cwd)
		if err == nil {
			requiredSpace := uint64(float64(estimated) * 1.05)
			hasEnoughSpace = availableSpace >= requiredSpace
		}
	}

	validation := BackupValidation{
		Items:               items,
		TotalSize:           totalSize,
		TotalSizeHuman:      humanize.Bytes(totalSize),
		EstimatedSize:       estimated,
		EstimatedSizeHuman:  humanize.Bytes(estimated),
		AvailableSpace:      availableSpace,
		AvailableSpaceHuman: humanize.Bytes(availableSpace),
		HasEnoughSpace:      hasEnoughSpace,
	}

	logger.Info("Backup validation: %d items, total=%s, estimated=%s, available=%s, hasSpace=%v",
		len(items), validation.TotalSizeHuman, validation.EstimatedSizeHuman,
		validation.AvailableSpaceHuman, validation.HasEnoughSpace)

	return validation
}

// GetBackupFolders returns the list of mod folders that should be backed up
func (a *Aurora) GetBackupFolders() []string {
	repo := repository.NewPenumbraRepository(a.cfg)

	folders := []string{}
	for _, mod := range repo.Mods {
		if len(mod.Collections) == 0 {
			continue
		}

		filtered, filterName := isModFiltered(&mod, a.cfg.Filters)
		if filtered {
			logger.Info("Mod filtered: %s (by %s)", mod.Name, filterName)
		} else {
			file := filepath.Join(a.cfg.Mods.Path, mod.Name)
			folders = append(folders, file)
		}
	}

	logger.Info("GetBackupFolders: %d folders to backup", len(folders))
	return folders
}
