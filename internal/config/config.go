package config

import (
	"aurora/internal/logger"
	"aurora/internal/util"
	"encoding/json"
	"errors"
	"log"
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
	Filters     []string `json:"filters"`
	Concurrency int      `json:"concurrency"`
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
}

func NewConfig(reset bool) *Config {
	createIfMissing(reset)
	var config Config

	if err := util.ReadJSONFile(ConfigFile, &config); err != nil {
		logger.Error("Failed to read config file: %v", err)
		log.Fatalf("Failed to read config file: %v", err)
	}

	logger.Info("Config loaded: penumbra=%s, mods=%s", config.Penumbra.Path, config.Mods.Path)
	return &config
}

func createIfMissing(reset bool) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "C:\\Users\\<user>"
	}
	if _, err := os.Stat(ConfigFile); errors.Is(err, os.ErrNotExist) || reset {
		config := Config{
			Penumbra: PenumbraConfig{
				Path: filepath.Join(homeDir, "AppData", "Roaming", "XIVLauncher", "pluginConfigs", "Penumbra"),
			},
			Mods: ModsConfig{
				Path: "",
			},
		}
		config.Save()
	}
}

func (c *Config) Status() ConfigStatus {
	status := ConfigStatus{
		Valid:    true,
		Penumbra: "OK",
		Mods:     "OK",
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

func (c *Config) Save() {
	file, err := os.Create(ConfigFile)
	if err != nil {
		log.Fatalf("Failed to create config file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(c); err != nil {
		log.Fatalf("Failed to write default config to file: %v", err)
	}
}
