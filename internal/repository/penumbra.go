package repository

import (
	"aurora/internal/config"
	"aurora/internal/util"
	"io/fs"
	"log"
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
	TotalDiskSize           uint64
	TotalUsedModsDiskSize   uint64
	ModsWithCollectionCount int
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

const sortOrderFile = "sort_order.json"
const collectionsFolder = "collections"

// SortOrderData represents the parsed sort_order.json file
type SortOrderData struct {
	Data map[string]string `json:"Data"`
}

// LoadSortOrder reads and parses the sort_order.json file from the given path
func LoadSortOrder(sortOrderPath string) (*SortOrderData, error) {
	var data SortOrderData
	if err := util.ReadJSONFile(sortOrderPath, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func NewPenumbraRepository(config *config.Config) *PenumbraRepository {
	mods := loadSortOrder(config)
	repo := PenumbraRepository{
		path:        config.Penumbra.Path,
		Mods:        mods,
		Collections: loadCollections(mods, config),
		Stats:       PenumbraStats{},
	}
	for i, mod := range mods {
		repo.Stats.TotalDiskSize += mod.Size
		for _, col := range repo.Collections {
			for _, colMod := range col.Mods {
				if mod.Name == colMod.Name {
					mods[i].Collections = append(mods[i].Collections, &col)
				}
			}
		}
		if len(mods[i].Collections) > 0 {
			repo.Stats.TotalUsedModsDiskSize += mod.Size
		} else {
			repo.Stats.ModsWithCollectionCount++
		}
	}
	return &repo
}

func loadSortOrder(config *config.Config) []PenumbraMod {
	path := filepath.Join(config.Penumbra.Path, sortOrderFile)

	sortOrder, err := LoadSortOrder(path)
	if err != nil {
		log.Fatalf("Failed to load sort_order.json: %v", err)
	}

	mods := make([]PenumbraMod, 0, len(sortOrder.Data))
	seen := make(map[string]bool)
	for modPath, name := range sortOrder.Data {
		// Use filepath.Base to handle both / and \ separators
		modName := filepath.Base(name)
		// Skip duplicates by name
		if seen[modName] {
			continue
		}
		size, err := getModSize(filepath.Join(config.Mods.Path, modName))
		if err != nil || size == 0 {
			continue
		}
		seen[modName] = true
		mods = append(mods, PenumbraMod{Name: modName, Path: modPath, Size: size})
	}

	slices.SortFunc(mods, func(a, b PenumbraMod) int {
		if a.Name < b.Name {
			return -1
		} else if a.Name > b.Name {
			return 1
		}
		return 0
	})

	return mods
}

func loadCollections(mods []PenumbraMod, config *config.Config) []PenumbraCollection {
	path := filepath.Join(config.Penumbra.Path, collectionsFolder)
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Fatalf("Failed to read penumbra collections folder: %v", err)
	}

	collections := []PenumbraCollection{}
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			filePath := filepath.Join(path, entry.Name())
			var rawCollection collection
			if err := util.ReadJSONFile(filePath, &rawCollection); err != nil {
				log.Printf("Failed to read penumbra collection file: %s: %v", entry.Name(), err)
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

	return collections
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
			// Skip entries we can't access (e.g., permission denied)
			return nil
		}
		if !d.IsDir() {
			// Get file info for size
			info, err := d.Info()
			if err != nil {
				return nil // Skip files with errors
			}
			size += uint64(info.Size())
		}
		return nil
	})

	return size, err
}
