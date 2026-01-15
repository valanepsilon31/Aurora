package aurora

import (
	"aurora/internal/config"
	"aurora/internal/repository"
	"slices"

	"github.com/dustin/go-humanize"
)

// Aurora is the main service providing all operations
type Aurora struct {
	cfg *config.Config
}

// New creates a new Aurora instance
func New() *Aurora {
	return &Aurora{
		cfg: config.NewConfig(false),
	}
}

// NewWithReset creates a new Aurora instance, optionally resetting config
func NewWithReset(reset bool) *Aurora {
	return &Aurora{
		cfg: config.NewConfig(reset),
	}
}

// GetConfig returns the current configuration
func (a *Aurora) GetConfig() ConfigResult {
	status := a.cfg.Status()
	return ConfigResult{
		ConfigFile:   a.cfg.ConfigFile,
		PenumbraPath: a.cfg.Penumbra.Path,
		ModsPath:     a.cfg.Mods.Path,
		Filters:      a.cfg.Filters,
		Concurrency:  a.cfg.Concurrency,
		Status: ConfigStatus{
			Valid:          status.Valid,
			PenumbraStatus: status.Penumbra,
			ModsStatus:     status.Mods,
		},
	}
}

// UpdateConfig updates the configuration paths
func (a *Aurora) UpdateConfig(penumbraPath, modsPath string) error {
	a.cfg.Penumbra.Path = penumbraPath
	a.cfg.Mods.Path = modsPath
	a.cfg.Save()
	// Reload config
	a.cfg = config.NewConfig(false)
	return nil
}

// IsConfigValid returns whether the current config is valid
func (a *Aurora) IsConfigValid() bool {
	return a.cfg.Status().Valid
}

// GetCollections returns all collections and mods
func (a *Aurora) GetCollections() CollectionsResult {
	repo := repository.NewPenumbraRepository(a.cfg)

	collections := make([]Collection, len(repo.Collections))
	for i, col := range repo.Collections {
		mods := make([]Mod, len(col.Mods))
		for j, mod := range col.Mods {
			mods[j] = Mod{
				Name:      mod.Name,
				Path:      mod.Path,
				Size:      mod.Size,
				SizeHuman: humanize.Bytes(mod.Size),
			}
		}
		collections[i] = Collection{
			Name: col.Name,
			Mods: mods,
		}
	}

	mods := make([]Mod, len(repo.Mods))
	for i, mod := range repo.Mods {
		colNames := make([]string, len(mod.Collections))
		for j, col := range mod.Collections {
			colNames[j] = col.Name
		}
		mods[i] = Mod{
			Name:        mod.Name,
			Path:        mod.Path,
			Size:        mod.Size,
			SizeHuman:   humanize.Bytes(mod.Size),
			Collections: colNames,
		}
	}

	usedMods := len(repo.Mods) - repo.Stats.ModsWithCollectionCount
	return CollectionsResult{
		Collections: collections,
		Mods:        mods,
		Stats: Stats{
			TotalMods:          len(repo.Mods),
			UsedMods:           usedMods,
			UnusedMods:         repo.Stats.ModsWithCollectionCount,
			TotalDiskSize:      repo.Stats.TotalDiskSize,
			TotalDiskSizeHuman: humanize.Bytes(repo.Stats.TotalDiskSize),
			UsedDiskSize:       repo.Stats.TotalUsedModsDiskSize,
			UsedDiskSizeHuman:  humanize.Bytes(repo.Stats.TotalUsedModsDiskSize),
			CollectionCount:    len(repo.Collections),
		},
	}
}

// Config returns the internal config (for CLI compatibility)
func (a *Aurora) Config() *config.Config {
	return a.cfg
}

// AddFilter adds a new filter pattern
func (a *Aurora) AddFilter(filter string) {
	if slices.Contains(a.cfg.Filters, filter) {
		return
	}
	a.cfg.Filters = append(a.cfg.Filters, filter)
	a.cfg.Save()
}

// RemoveFilter removes a filter pattern
func (a *Aurora) RemoveFilter(filter string) {
	for i, f := range a.cfg.Filters {
		if f == filter {
			a.cfg.Filters = append(a.cfg.Filters[:i], a.cfg.Filters[i+1:]...)
			a.cfg.Save()
			return
		}
	}
}

// SetConcurrency sets the concurrency level for backups
func (a *Aurora) SetConcurrency(concurrency int) {
	if concurrency < 0 {
		concurrency = 0
	}
	a.cfg.Concurrency = concurrency
	a.cfg.Save()
}

// GetConcurrency returns the current concurrency setting
func (a *Aurora) GetConcurrency() int {
	return a.cfg.Concurrency
}
