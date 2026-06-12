package config

import (
	"aurora/internal/logger"
	"aurora/internal/util"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ConfigFile is the path to the config file (next to the executable)
var ConfigFile = getConfigPath()

func getConfigPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "config.json"
	}
	return filepath.Join(filepath.Dir(exe), "config.json")
}

type Config struct {
	Penumbra    PenumbraConfig
	Mods        ModsConfig
	Filters     []string `json:"filters"`    // Exclusions: matching mods are dropped from backups
	Inclusions  []string `json:"inclusions"` // Matching mods are always backed up (wins over exclusions and missing collections)
	Concurrency int      `json:"concurrency"`
	Compression string   `json:"compression"` // "normal" (default) or "max"
	Output      string   `json:"output"`      // Backup output directory ("" = current working directory)
}

type PenumbraConfig struct {
	Path string `json:"path"`
}

type ModsConfig struct {
	Path string `json:"path"`
}

type ConfigStatus struct {
	Valid    bool
	Penumbra string
	Mods     string
	Output   string
}

func NewConfig(reset bool) (*Config, error) {
	if err := createIfMissing(reset); err != nil {
		return nil, err
	}
	var config Config

	if err := util.ReadJSONFile(ConfigFile, &config); err != nil {
		logger.Error("Failed to read config file: %v", err)
		return nil, fmt.Errorf("read config file %s: %w", ConfigFile, err)
	}

	logger.Info("Config loaded: penumbra=%s, mods=%s", config.Penumbra.Path, config.Mods.Path)
	return &config, nil
}

func createIfMissing(reset bool) error {
	// Default Penumbra path only when the home directory is known;
	// an empty path is reported by validation and fixable in the UI
	defaultPenumbra := ""
	if homeDir, err := os.UserHomeDir(); err == nil {
		defaultPenumbra = filepath.Join(homeDir, "AppData", "Roaming", "XIVLauncher", "pluginConfigs", "Penumbra")
	}
	if _, err := os.Stat(ConfigFile); errors.Is(err, os.ErrNotExist) || reset {
		config := Config{
			Penumbra: PenumbraConfig{
				Path: defaultPenumbra,
			},
			Mods: ModsConfig{
				Path: "",
			},
			Compression: "normal",
		}
		return config.Save()
	}
	return nil
}

func (c *Config) Status() ConfigStatus {
	status := ConfigStatus{
		Valid:    true,
		Penumbra: "OK",
		Mods:     "OK",
		Output:   "OK",
	}

	// Output dir is optional ("" = current working directory)
	if c.Output != "" {
		if fileInfo, err := os.Stat(c.Output); err != nil || !fileInfo.IsDir() {
			status.Output = "Invalid output path"
			status.Valid = false
			logger.Warn("Invalid output path: %s", c.Output)
		}
	}

	fileInfo, err := os.Stat(c.Penumbra.Path)
	if err != nil || !fileInfo.IsDir() {
		status.Penumbra = "Invalid Penumbra path"
		status.Valid = false
		logger.Warn("Invalid Penumbra path: %s", c.Penumbra.Path)
	} else {
		collectionsFolder := filepath.Join(c.Penumbra.Path, "collections")
		fileInfo, err = os.Stat(collectionsFolder)
		if err != nil || !fileInfo.IsDir() {
			status.Penumbra = "collections folder not found in Penumbra path"
			status.Valid = false
			logger.Warn("collections folder not found at: %s", collectionsFolder)
		}
	}

	fileInfo, err = os.Stat(c.Mods.Path)
	if err != nil || !fileInfo.IsDir() {
		status.Mods = "Invalid Mods path"
		status.Valid = false
		logger.Warn("Invalid Mods path: %s", c.Mods.Path)
	} else if status.Penumbra == "OK" {
		if !hasModFolders(c.Mods.Path) {
			status.Mods = "No mod folders found in Mods path"
			status.Valid = false
			logger.Warn("No mod folders found in Mods path: %s", c.Mods.Path)
		}
	}

	logger.Info("Config status: valid=%v, penumbra=%s, mods=%s", status.Valid, status.Penumbra, status.Mods)
	return status
}

// hasModFolders checks if the mods directory contains at least one subdirectory
func hasModFolders(modsPath string) bool {
	entries, err := os.ReadDir(modsPath)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			return true
		}
	}

	return false
}

func (c *Config) Save() error {
	file, err := os.Create(ConfigFile)
	if err != nil {
		logger.Error("Failed to create config file: %v", err)
		return fmt.Errorf("create config file %s: %w", ConfigFile, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(c); err != nil {
		logger.Error("Failed to write config file: %v", err)
		return fmt.Errorf("write config file %s: %w", ConfigFile, err)
	}
	return nil
}
