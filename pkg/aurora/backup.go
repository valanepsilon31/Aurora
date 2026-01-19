package aurora

import (
	"aurora/internal/repository"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/creativeyann17/go-delta/pkg/compress"
	"github.com/dustin/go-humanize"
)

// BackupOutputPath is the base filename for backup archives
const BackupOutputPath = "backup_part.zip"

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
		for _, filter := range a.cfg.Filters {
			if strings.HasPrefix(mod.Name, filter) ||
				strings.HasPrefix(mod.Path, filter) ||
				(len(mod.Collections) == 1 && strings.HasPrefix(mod.Collections[0].Name, filter)) {
				item.IsFiltered = true
				item.FilteredBy = filter
				break
			}
		}

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

	return BackupValidation{
		Items:               items,
		TotalSize:           totalSize,
		TotalSizeHuman:      humanize.Bytes(totalSize),
		EstimatedSize:       estimated,
		EstimatedSizeHuman:  humanize.Bytes(estimated),
		AvailableSpace:      availableSpace,
		AvailableSpaceHuman: humanize.Bytes(availableSpace),
		HasEnoughSpace:      hasEnoughSpace,
	}
}

// RunBackup executes the backup with progress callback
func (a *Aurora) RunBackup(threads int, progressCb func(BackupProgress)) (*BackupResult, error) {
	repo := repository.NewPenumbraRepository(a.cfg)

	folders := []string{}
	for _, mod := range repo.Mods {
		// Check if mod is in a collection
		if len(mod.Collections) == 0 {
			continue
		}

		// Check filters
		filtered := false
		for _, filter := range a.cfg.Filters {
			if strings.HasPrefix(mod.Name, filter) ||
				strings.HasPrefix(mod.Path, filter) ||
				(len(mod.Collections) == 1 && strings.HasPrefix(mod.Collections[0].Name, filter)) {
				filtered = true
				break
			}
		}

		if !filtered {
			file := filepath.Join(a.cfg.Mods.Path, mod.Path)
			folders = append(folders, file)
		}
	}

	if len(folders) == 0 {
		return nil, fmt.Errorf("no mods to backup")
	}

	if threads <= 0 {
		threads = 1
	}

	opts := &compress.Options{
		OutputPath:   BackupOutputPath,
		Files:        folders,
		MaxThreads:   threads,
		Level:        9,
		UseZipFormat: true,
		Quiet:        true,
	}

	// Use library progress callback
	libProgressCb, progress := compress.ProgressBarCallback()

	result, err := compress.Compress(opts, libProgressCb)

	if progress != nil {
		progress.Wait()
	}

	if err != nil {
		if progressCb != nil {
			progressCb(BackupProgress{Done: true, Error: err.Error()})
		}
		return nil, err
	}

	if progressCb != nil {
		progressCb(BackupProgress{Done: true, Percent: 100})
	}

	return &BackupResult{
		OutputPath:     opts.OutputPath,
		OriginalSize:   result.OriginalSize,
		CompressedSize: result.CompressedSize,
		Ratio:          fmt.Sprintf("%.1f%%", float64(result.CompressedSize)/float64(result.OriginalSize)*100),
	}, nil
}

// GetBackupFolders returns the list of folders that would be backed up (for CLI)
func (a *Aurora) GetBackupFolders() []string {
	repo := repository.NewPenumbraRepository(a.cfg)

	folders := []string{}
	for _, mod := range repo.Mods {
		if len(mod.Collections) == 0 {
			continue
		}

		filtered := false
		for _, filter := range a.cfg.Filters {
			if strings.HasPrefix(mod.Name, filter) ||
				strings.HasPrefix(mod.Path, filter) ||
				(len(mod.Collections) == 1 && strings.HasPrefix(mod.Collections[0].Name, filter)) {
				filtered = true
				break
			}
		}

		if !filtered {
			file := filepath.Join(a.cfg.Mods.Path, mod.Path)
			folders = append(folders, file)
		}
	}

	return folders
}
