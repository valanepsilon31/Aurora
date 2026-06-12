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
func New() (*Aurora, error) {
	return NewWithReset(false)
}

// NewWithReset creates a new Aurora instance, optionally resetting config
func NewWithReset(reset bool) (*Aurora, error) {
	cfg, err := config.NewConfig(reset)
	if err != nil {
		return nil, err
	}
	return &Aurora{cfg: cfg}, nil
}

// ReloadConfig reloads the configuration from disk
func (a *Aurora) ReloadConfig() error {
	cfg, err := config.NewConfig(false)
	if err != nil {
		return err
	}
	a.cfg = cfg
	return nil
}

// GetConfig returns the current configuration
func (a *Aurora) GetConfig() ConfigResult {
	status := a.cfg.Status()
	return ConfigResult{
		PenumbraPath: a.cfg.Penumbra.Path,
		ModsPath:     a.cfg.Mods.Path,
		OutputPath:   a.cfg.Output,
		Filters:      a.cfg.Filters,
		Inclusions:   a.cfg.Inclusions,
		Concurrency:  a.cfg.Concurrency,
		Compression:  a.GetCompression(),
		Status: ConfigStatus{
			Valid:          status.Valid,
			PenumbraStatus: status.Penumbra,
			ModsStatus:     status.Mods,
			OutputStatus:   status.Output,
		},
	}
}

// UpdateConfig updates the configuration paths
func (a *Aurora) UpdateConfig(penumbraPath, modsPath, outputPath string) error {
	a.cfg.Penumbra.Path = penumbraPath
	a.cfg.Mods.Path = modsPath
	a.cfg.Output = outputPath
	if err := a.cfg.Save(); err != nil {
		return err
	}
	return a.ReloadConfig()
}

// IsConfigValid returns whether the current config is valid
func (a *Aurora) IsConfigValid() bool {
	return a.cfg.Status().Valid
}

// GetCollections returns all collections and mods
func (a *Aurora) GetCollections() (CollectionsResult, error) {
	repo, err := repository.NewPenumbraRepository(a.cfg)
	if err != nil {
		return CollectionsResult{}, err
	}

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

	usedMods := len(repo.Mods) - repo.Stats.UnreferencedModsCount
	return CollectionsResult{
		Collections: collections,
		Mods:        mods,
		Stats: Stats{
			TotalMods:          len(repo.Mods),
			UsedMods:           usedMods,
			UnusedMods:         repo.Stats.UnreferencedModsCount,
			TotalDiskSize:      repo.Stats.TotalDiskSize,
			TotalDiskSizeHuman: humanize.Bytes(repo.Stats.TotalDiskSize),
			UsedDiskSize:       repo.Stats.TotalUsedModsDiskSize,
			UsedDiskSizeHuman:  humanize.Bytes(repo.Stats.TotalUsedModsDiskSize),
			CollectionCount:    len(repo.Collections),
		},
	}, nil
}

// Config returns the internal config (for CLI compatibility)
func (a *Aurora) Config() *config.Config {
	return a.cfg
}

// AddFilter adds a new filter pattern
func (a *Aurora) AddFilter(filter string) error {
	if slices.Contains(a.cfg.Filters, filter) {
		return nil
	}
	a.cfg.Filters = append(a.cfg.Filters, filter)
	return a.cfg.Save()
}

// RemoveFilter removes a filter pattern
func (a *Aurora) RemoveFilter(filter string) error {
	for i, f := range a.cfg.Filters {
		if f == filter {
			a.cfg.Filters = append(a.cfg.Filters[:i], a.cfg.Filters[i+1:]...)
			return a.cfg.Save()
		}
	}
	return nil
}

// AddInclusion adds a new inclusion pattern
func (a *Aurora) AddInclusion(inclusion string) error {
	if slices.Contains(a.cfg.Inclusions, inclusion) {
		return nil
	}
	a.cfg.Inclusions = append(a.cfg.Inclusions, inclusion)
	return a.cfg.Save()
}

// RemoveInclusion removes an inclusion pattern
func (a *Aurora) RemoveInclusion(inclusion string) error {
	for i, f := range a.cfg.Inclusions {
		if f == inclusion {
			a.cfg.Inclusions = append(a.cfg.Inclusions[:i], a.cfg.Inclusions[i+1:]...)
			return a.cfg.Save()
		}
	}
	return nil
}

// SetConcurrency sets the concurrency level for backups
func (a *Aurora) SetConcurrency(concurrency int) error {
	if concurrency < 0 {
		concurrency = 0
	}
	a.cfg.Concurrency = concurrency
	return a.cfg.Save()
}

// GetConcurrency returns the current concurrency setting
func (a *Aurora) GetConcurrency() int {
	return a.cfg.Concurrency
}

// SetCompression sets the backup compression preset ("normal" or "max")
func (a *Aurora) SetCompression(compression string) error {
	if compression != CompressionMax {
		compression = CompressionNormal
	}
	a.cfg.Compression = compression
	return a.cfg.Save()
}

// GetCompression returns the current compression preset, normalized
func (a *Aurora) GetCompression() string {
	if a.cfg.Compression == CompressionMax {
		return CompressionMax
	}
	return CompressionNormal
}
