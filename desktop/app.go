package main

import (
	"aurora/internal/logger"
	"aurora/pkg/aurora"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/creativeyann17/go-delta/pkg/compress"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct holds the application state
type App struct {
	ctx     context.Context
	aurora  *aurora.Aurora
	version string
}

// NewApp creates a new App instance
func NewApp(version string) *App {
	return &App{version: version}
}

// GetVersion returns the application version
func (a *App) GetVersion() string {
	return a.version
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	logger.Init(logger.GetLogPath())
	logger.Info("Aurora desktop app starting, version=%s", a.version)
	a.aurora = aurora.New()
}

// GetConfig returns the current configuration
func (a *App) GetConfig() aurora.ConfigResult {
	return a.aurora.GetConfig()
}

// ReloadConfig reloads configuration from disk
func (a *App) ReloadConfig() aurora.ConfigResult {
	a.aurora.ReloadConfig()
	return a.aurora.GetConfig()
}

// BrowseDirectory opens a directory picker dialog
func (a *App) BrowseDirectory(title string, defaultPath string) (string, error) {
	opts := runtime.OpenDialogOptions{
		Title: title,
	}
	// If defaultPath exists, use it as the starting directory
	if defaultPath != "" {
		if info, err := os.Stat(defaultPath); err == nil && info.IsDir() {
			opts.DefaultDirectory = defaultPath
		}
	}
	return runtime.OpenDirectoryDialog(a.ctx, opts)
}

// UpdateConfig updates the configuration paths
func (a *App) UpdateConfig(penumbraPath, modsPath string) error {
	return a.aurora.UpdateConfig(penumbraPath, modsPath)
}

// IsConfigValid checks if configuration is valid
func (a *App) IsConfigValid() bool {
	return a.aurora.IsConfigValid()
}

// AddFilter adds a new filter pattern
func (a *App) AddFilter(filter string) {
	a.aurora.AddFilter(filter)
}

// RemoveFilter removes a filter pattern
func (a *App) RemoveFilter(filter string) {
	a.aurora.RemoveFilter(filter)
}

// SetConcurrency sets the concurrency level for backups
func (a *App) SetConcurrency(concurrency int) {
	a.aurora.SetConcurrency(concurrency)
}

// GetCollections returns all collections and mods
func (a *App) GetCollections() aurora.CollectionsResult {
	return a.aurora.GetCollections()
}

// ValidateBackup returns backup preview
func (a *App) ValidateBackup() aurora.BackupValidation {
	return a.aurora.ValidateBackup()
}

// BackupProgressEvent represents progress sent to frontend
type BackupProgressEvent struct {
	Percent float64 `json:"percent"`
	Current string  `json:"current"`
	Done    bool    `json:"done"`
	Error   string  `json:"error,omitempty"`
}

// RunBackup executes the backup operation with progress events
func (a *App) RunBackup(threads int) (*aurora.BackupResult, error) {
	logger.Info("RunBackup started with threads=%d", threads)
	folders := a.aurora.GetBackupFolders()
	if len(folders) == 0 {
		logger.Warn("RunBackup: no mods to backup")
		return nil, fmt.Errorf("no mods to backup")
	}

	if threads < 0 {
		threads = 0
	}

	// Emit initial progress
	runtime.EventsEmit(a.ctx, "backup:progress", BackupProgressEvent{
		Percent: 0,
		Current: "Preparing backup...",
		Done:    false,
	})

	opts := aurora.NewBackupOptions(folders, threads, true)

	var totalFiles int64
	var completedFiles int64
	currentFile := ""

	progressCb := func(event compress.ProgressEvent) {
		switch event.Type {
		case compress.EventStart:
			// EventStart contains total file count in event.Total
			totalFiles = event.Total
			runtime.EventsEmit(a.ctx, "backup:progress", BackupProgressEvent{
				Percent: 0,
				Current: fmt.Sprintf("Starting... (0/%d)", totalFiles),
				Done:    false,
			})

		case compress.EventFileStart:
			currentFile = filepath.Base(event.FilePath)
			if len(currentFile) > 35 {
				currentFile = currentFile[:32] + "..."
			}
			runtime.EventsEmit(a.ctx, "backup:progress", BackupProgressEvent{
				Percent: float64(completedFiles) / float64(totalFiles) * 100,
				Current: fmt.Sprintf("%s (%d/%d)", currentFile, completedFiles, totalFiles),
				Done:    false,
			})

		case compress.EventFileComplete:
			completedFiles++
			var percent float64
			if totalFiles > 0 {
				percent = float64(completedFiles) / float64(totalFiles) * 100
			}
			runtime.EventsEmit(a.ctx, "backup:progress", BackupProgressEvent{
				Percent: percent,
				Current: fmt.Sprintf("%s (%d/%d)", currentFile, completedFiles, totalFiles),
				Done:    false,
			})
		}
	}

	result, err := compress.Compress(opts, progressCb)

	if err != nil {
		logger.Error("Backup failed: %v", err)
		runtime.EventsEmit(a.ctx, "backup:progress", BackupProgressEvent{
			Percent: 0,
			Current: "",
			Done:    true,
			Error:   err.Error(),
		})
		return nil, err
	}

	// Emit completion
	runtime.EventsEmit(a.ctx, "backup:progress", BackupProgressEvent{
		Percent: 100,
		Current: "Complete!",
		Done:    true,
	})

	// Find actual output files created
	outputDisplay := findBackupOutputFiles()

	backupResult := &aurora.BackupResult{
		OutputPath:     outputDisplay,
		OriginalSize:   result.OriginalSize,
		CompressedSize: result.CompressedSize,
		Ratio:          fmt.Sprintf("%.1f%%", float64(result.CompressedSize)/float64(result.OriginalSize)*100),
	}
	logger.Info("Backup completed: output=%s, ratio=%s", backupResult.OutputPath, backupResult.Ratio)
	return backupResult, nil
}

// findBackupOutputFiles finds the backup files created and returns a display string
func findBackupOutputFiles() string {
	// Check for multi-part files first (backup_part_01.zip, etc.)
	pattern := "backup_part_*.zip"
	matches, err := filepath.Glob(pattern)
	if err == nil && len(matches) > 0 {
		if len(matches) == 1 {
			return filepath.Base(matches[0])
		}
		// Multiple files: show range
		return fmt.Sprintf("backup_part_01.zip ... backup_part_%02d.zip", len(matches))
	}

	// Fall back to single file
	if _, err := os.Stat(aurora.BackupOutputPath); err == nil {
		return aurora.BackupOutputPath
	}

	return aurora.BackupOutputPath
}
