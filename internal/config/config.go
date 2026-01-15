package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

const configFile = "config.json"

type Config struct {
	ConfigFile  string
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

	contentBytes, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	// strip BOM if present
	contentBytes = bytes.TrimPrefix(contentBytes, []byte("\xEF\xBB\xBF"))

	err = json.Unmarshal(contentBytes, &config)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	config.ConfigFile = configFile
	return &config
}

func createIfMissing(reset bool) {
	currentUser, err := user.Current()
	if err != nil {
		currentUser = &user.User{Username: "<user>"}
	}
	if _, err := os.Stat(configFile); errors.Is(err, os.ErrNotExist) || reset {
		config := Config{
			Penumbra: PenumbraConfig{
				Path: fmt.Sprintf("C:\\Users\\%s\\AppData\\roaming\\XIVLauncher\\pluginsConfig\\Penumbra", currentUser.Username),
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
	} else {
		sortOrderJsonPath := filepath.Join(c.Penumbra.Path, "sort_order.json")
		fileInfo, err = os.Stat(sortOrderJsonPath)
		if err != nil {
			status.Penumbra = "sort_order.json not found in Penumbra path"
			status.Valid = false
		} else {
			collectionsFolder := filepath.Join(c.Penumbra.Path, "collections")
			fileInfo, err = os.Stat(collectionsFolder)
			if err != nil || !fileInfo.IsDir() {
				status.Penumbra = "collections folder not found in Penumbra path"
				status.Valid = false
			}
		}
	}
	fileInfo, err = os.Stat(c.Mods.Path)
	if err != nil || !fileInfo.IsDir() {
		status.Mods = "Invalid Mods path"
		status.Valid = false
	}
	return status
}

func (c *Config) Save() {
	file, err := os.Create(configFile)
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
