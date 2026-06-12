package main

import (
	"aurora/internal/logger"
	"aurora/pkg/aurora"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"sync"

	"github.com/creativeyann17/go-delta/pkg/compress"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct holds the application state
type App struct {
	ctx     context.Context
	aurora  *aurora.Aurora
	initErr error // set when config loading failed at startup
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
	svc, err := aurora.New()
	if err != nil {
		// Don't crash the window: surface the error through GetConfig status
		logger.Error("Startup failed: %v", err)
		a.initErr = err
		return
	}
	a.aurora = svc
}

// svc returns the aurora service or the startup error
func (a *App) svc() (*aurora.Aurora, error) {
	if a.aurora == nil {
		if a.initErr != nil {
			return nil, a.initErr
		}
		return nil, fmt.Errorf("aurora service not initialized")
	}
	return a.aurora, nil
}

// GetConfig returns the current configuration. A startup config failure is
// reported through the status fields instead of crashing the app.
func (a *App) GetConfig() aurora.ConfigResult {
	if a.aurora == nil {
		msg := "configuration could not be loaded"
		if a.initErr != nil {
			msg = a.initErr.Error()
		}
		return aurora.ConfigResult{
			Status: aurora.ConfigStatus{
				Valid:          false,
				PenumbraStatus: msg,
				ModsStatus:     "-",
				OutputStatus:   "-",
			},
		}
	}
	return a.aurora.GetConfig()
}

// ReloadConfig reloads configuration from disk
func (a *App) ReloadConfig() (aurora.ConfigResult, error) {
	svc, err := a.svc()
	if err != nil {
		// Retry full startup init: the user may have fixed config.json
		if svc2, err2 := aurora.New(); err2 == nil {
			a.aurora = svc2
			a.initErr = nil
			return a.aurora.GetConfig(), nil
		}
		return a.GetConfig(), err
	}
	if err := svc.ReloadConfig(); err != nil {
		return a.GetConfig(), err
	}
	return svc.GetConfig(), nil
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
func (a *App) UpdateConfig(penumbraPath, modsPath, outputPath string) error {
	svc, err := a.svc()
	if err != nil {
		return err
	}
	return svc.UpdateConfig(penumbraPath, modsPath, outputPath)
}

// IsConfigValid checks if configuration is valid
func (a *App) IsConfigValid() bool {
	return a.aurora != nil && a.aurora.IsConfigValid()
}

// AddFilter adds a new filter pattern
func (a *App) AddFilter(filter string) error {
	svc, err := a.svc()
	if err != nil {
		return err
	}
	return svc.AddFilter(filter)
}

// RemoveFilter removes a filter pattern
func (a *App) RemoveFilter(filter string) error {
	svc, err := a.svc()
	if err != nil {
		return err
	}
	return svc.RemoveFilter(filter)
}

// AddInclusion adds a new inclusion pattern
func (a *App) AddInclusion(inclusion string) error {
	svc, err := a.svc()
	if err != nil {
		return err
	}
	return svc.AddInclusion(inclusion)
}

// RemoveInclusion removes an inclusion pattern
func (a *App) RemoveInclusion(inclusion string) error {
	svc, err := a.svc()
	if err != nil {
		return err
	}
	return svc.RemoveInclusion(inclusion)
}

// SetConcurrency sets the concurrency level for backups
func (a *App) SetConcurrency(concurrency int) error {
	svc, err := a.svc()
	if err != nil {
		return err
	}
	return svc.SetConcurrency(concurrency)
}

// SetCompression sets the backup compression preset ("max" or "normal")
func (a *App) SetCompression(compression string) error {
	svc, err := a.svc()
	if err != nil {
		return err
	}
	return svc.SetCompression(compression)
}

// GetCollections returns all collections and mods
func (a *App) GetCollections() (aurora.CollectionsResult, error) {
	svc, err := a.svc()
	if err != nil {
		return aurora.CollectionsResult{}, err
	}
	return svc.GetCollections()
}

// ValidateBackup returns backup preview
func (a *App) ValidateBackup() (aurora.BackupValidation, error) {
	svc, err := a.svc()
	if err != nil {
		return aurora.BackupValidation{}, err
	}
	return svc.ValidateBackup()
}

// GetFilterMatches returns per-pattern mod match counts for the filters
func (a *App) GetFilterMatches() (aurora.FilterMatches, error) {
	svc, err := a.svc()
	if err != nil {
		return aurora.FilterMatches{}, err
	}
	return svc.GetFilterMatches()
}

// OpenOutputFolder opens the backup output directory in the file manager.
// Launches the platform file manager directly: BrowserOpenURL with a file://
// directory URL is unreliable (may target a browser or nothing at all).
func (a *App) OpenOutputFolder() error {
	svc, err := a.svc()
	if err != nil {
		return err
	}
	dir := svc.GetConfig().OutputPath
	if dir == "" {
		if dir, err = os.Getwd(); err != nil {
			return fmt.Errorf("resolve output directory: %w", err)
		}
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve output directory: %w", err)
	}

	var cmd *exec.Cmd
	switch goruntime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", abs)
	case "darwin":
		cmd = exec.Command("open", abs)
	default:
		cmd = exec.Command("xdg-open", abs)
	}
	logger.Info("Opening output folder: %s", abs)
	// Start (not Run): explorer.exe famously exits non-zero even on success
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open folder %s: %w", abs, err)
	}
	return nil
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
	svc, err := a.svc()
	if err != nil {
		return nil, err
	}
	folders, err := svc.GetBackupFolders()
	if err != nil {
		return nil, err
	}
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

	outputDir := svc.GetConfig().OutputPath
	opts := aurora.NewBackupOptions(folders, threads, svc.GetCompression(), outputDir, true)

	// Progress is byte-weighted: file counting makes the bar crawl through
	// big mods then leap across thousands of small files. Events arrive from
	// several compression workers, hence the mutex.
	var mu sync.Mutex
	var totalFiles, completedFiles int64
	var totalBytes, completedBytes uint64
	inflight := make(map[string]int64) // bytes read so far per in-flight file
	currentFile := ""

	percentLocked := func() float64 {
		if totalBytes == 0 {
			if totalFiles == 0 {
				return 0
			}
			return float64(completedFiles) / float64(totalFiles) * 100
		}
		bytes := completedBytes
		for _, b := range inflight {
			bytes += uint64(b)
		}
		return float64(bytes) / float64(totalBytes) * 100
	}

	progressCb := func(event compress.ProgressEvent) {
		mu.Lock()
		switch event.Type {
		case compress.EventStart:
			totalFiles = event.Total
			totalBytes = event.TotalBytes
			currentFile = "Starting..."

		case compress.EventFileStart:
			name := filepath.Base(event.FilePath)
			if len(name) > 35 {
				name = name[:32] + "..."
			}
			currentFile = name

		case compress.EventFileProgress:
			inflight[event.FilePath] = event.Current

		case compress.EventFileComplete:
			completedFiles++
			completedBytes += uint64(event.Total)
			delete(inflight, event.FilePath)

		case compress.EventError:
			delete(inflight, event.FilePath)

		default:
			mu.Unlock()
			return
		}
		percent := percentLocked()
		label := fmt.Sprintf("%s (%d/%d)", currentFile, completedFiles, totalFiles)
		mu.Unlock()

		runtime.EventsEmit(a.ctx, "backup:progress", BackupProgressEvent{
			Percent: percent,
			Current: label,
			Done:    false,
		})
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
	outputDisplay := findBackupOutputFiles(outputDir)

	backupResult := &aurora.BackupResult{
		OutputPath:     outputDisplay,
		OriginalSize:   result.OriginalSize,
		CompressedSize: result.CompressedSize,
		Ratio:          fmt.Sprintf("%.1f%%", float64(result.CompressedSize)/float64(result.OriginalSize)*100),
	}
	logger.Info("Backup completed: output=%s, ratio=%s", backupResult.OutputPath, backupResult.Ratio)
	return backupResult, nil
}

// findBackupOutputFiles finds the backup files created in the output
// directory ("" = current working directory) and returns a display string
func findBackupOutputFiles(outputDir string) string {
	// Check for multi-part files first (backup_part_01.zip, etc.)
	pattern := filepath.Join(outputDir, "backup_part_*.zip")
	matches, err := filepath.Glob(pattern)
	if err == nil && len(matches) > 0 {
		if len(matches) == 1 {
			return filepath.Base(matches[0])
		}
		// Multiple files: show range
		return fmt.Sprintf("backup_part_01.zip ... backup_part_%02d.zip", len(matches))
	}

	// Fall back to single file
	single := filepath.Join(outputDir, aurora.BackupOutputPath)
	if _, err := os.Stat(single); err == nil {
		return filepath.Base(single)
	}

	return aurora.BackupOutputPath
}
