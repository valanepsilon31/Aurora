package repository

import (
	"aurora/internal/config"
	"aurora/internal/logger"
	"aurora/internal/util"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

type PenumbraRepository struct {
	path        string
	Mods        []PenumbraMod
	Collections []PenumbraCollection
	Stats       PenumbraStats
}

type PenumbraStats struct {
	TotalDiskSize         uint64
	TotalUsedModsDiskSize uint64
	UnreferencedModsCount int // Mods no collection references
}

type PenumbraMod struct {
	Path        string
	Name        string
	Collections []*PenumbraCollection
	Size        uint64
}

type PenumbraCollection struct {
	Name string
	Mods []*PenumbraMod
}

type collection struct {
	Name     string                        `json:"Name"`
	Settings map[string]collectionSettings `json:"Settings"`
}

type collectionSettings struct {
	Enabled bool `json:"Enabled"`
}

const collectionsFolder = "collections"

func NewPenumbraRepository(config *config.Config) (*PenumbraRepository, error) {
	return newRepository(config, true)
}

// NewPenumbraRepositoryNoSizes loads mods and collections without computing
// per-mod disk sizes. Walking every mod folder for sizes is by far the
// slowest part of a load; matching and counting don't need them.
// Note: empty mod folders are kept (the size==0 skip needs sizes).
func NewPenumbraRepositoryNoSizes(config *config.Config) (*PenumbraRepository, error) {
	return newRepository(config, false)
}

func newRepository(config *config.Config, withSizes bool) (*PenumbraRepository, error) {
	mods, err := loadMods(config, withSizes)
	if err != nil {
		return nil, err
	}
	collections, err := loadCollections(mods, config)
	if err != nil {
		return nil, err
	}
	repo := PenumbraRepository{
		path:        config.Penumbra.Path,
		Mods:        mods,
		Collections: collections,
		Stats:       PenumbraStats{},
	}
	for i, mod := range mods {
		repo.Stats.TotalDiskSize += mod.Size
		for ci := range repo.Collections {
			col := &repo.Collections[ci]
			for _, colMod := range col.Mods {
				if mod.Name == colMod.Name {
					mods[i].Collections = append(mods[i].Collections, col)
				}
			}
		}
		if len(mods[i].Collections) > 0 {
			repo.Stats.TotalUsedModsDiskSize += mod.Size
		} else {
			repo.Stats.UnreferencedModsCount++
		}
	}
	return &repo, nil
}

func loadMods(config *config.Config, withSizes bool) ([]PenumbraMod, error) {
	entries, err := os.ReadDir(config.Mods.Path)
	if err != nil {
		logger.Error("Failed to read mods directory: %v", err)
		return nil, fmt.Errorf("read mods directory %s: %w", config.Mods.Path, err)
	}

	mods := make([]PenumbraMod, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		modName := entry.Name()
		var size uint64
		if withSizes {
			modFullPath := filepath.Join(config.Mods.Path, modName)
			var err error
			size, err = getModSize(modFullPath)
			if err != nil {
				logger.Warn("Failed to get mod size for %s: %v", modName, err)
				continue
			}
			if size == 0 {
				logger.Info("Skipping mod with size 0: %s", modName)
				continue
			}
		}
		mods = append(mods, PenumbraMod{Name: modName, Path: modName, Size: size})
	}

	slices.SortFunc(mods, func(a, b PenumbraMod) int {
		if a.Name < b.Name {
			return -1
		} else if a.Name > b.Name {
			return 1
		}
		return 0
	})

	return mods, nil
}

func loadCollections(mods []PenumbraMod, config *config.Config) ([]PenumbraCollection, error) {
	path := filepath.Join(config.Penumbra.Path, collectionsFolder)
	entries, err := os.ReadDir(path)
	if err != nil {
		logger.Error("Failed to read penumbra collections folder: %v", err)
		return nil, fmt.Errorf("read penumbra collections folder %s: %w", path, err)
	}

	collections := []PenumbraCollection{}
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			filePath := filepath.Join(path, entry.Name())
			var rawCollection collection
			if err := util.ReadJSONFile(filePath, &rawCollection); err != nil {
				logger.Warn("Failed to read penumbra collection file %s: %v", entry.Name(), err)
				continue
			}

			penumbraCollection := PenumbraCollection{}
			penumbraCollection.Name = rawCollection.Name
			for name, settings := range rawCollection.Settings {
				if settings.Enabled {
					penumbraMod := findModByName(mods, name)
					if penumbraMod != nil {
						penumbraCollection.Mods = append(penumbraCollection.Mods, penumbraMod)
					}
				}
			}
			collections = append(collections, penumbraCollection)
		}
	}

	return collections, nil
}

func findModByName(mods []PenumbraMod, name string) *PenumbraMod {
	for i := range mods {
		if mods[i].Name == name {
			return &mods[i]
		}
	}
	return nil
}

func getModSize(root string) (uint64, error) {
	var size uint64

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			logger.Warn("Cannot access path %s: %v", path, err)
			return nil
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				logger.Warn("Cannot get file info for %s: %v", path, err)
				return nil
			}
			size += uint64(info.Size())
		}
		return nil
	})

	return size, err
}
