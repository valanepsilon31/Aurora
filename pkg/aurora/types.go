package aurora

// ConfigResult represents the current configuration state
type ConfigResult struct {
	ConfigFile   string       `json:"configFile"`
	PenumbraPath string       `json:"penumbraPath"`
	ModsPath     string       `json:"modsPath"`
	Filters      []string     `json:"filters"`
	Concurrency  int          `json:"concurrency"`
	Status       ConfigStatus `json:"status"`
}

// ConfigStatus represents validation status of paths
type ConfigStatus struct {
	Valid          bool   `json:"valid"`
	PenumbraStatus string `json:"penumbraStatus"`
	ModsStatus     string `json:"modsStatus"`
}

// Collection represents a Penumbra mod collection
type Collection struct {
	Name string `json:"name"`
	Mods []Mod  `json:"mods"`
}

// Mod represents a single mod
type Mod struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Size        uint64   `json:"size"`
	SizeHuman   string   `json:"sizeHuman"`
	Collections []string `json:"collections"`
}

// Stats represents repository statistics
type Stats struct {
	TotalMods          int    `json:"totalMods"`
	UsedMods           int    `json:"usedMods"`
	UnusedMods         int    `json:"unusedMods"`
	TotalDiskSize      uint64 `json:"totalDiskSize"`
	TotalDiskSizeHuman string `json:"totalDiskSizeHuman"`
	UsedDiskSize       uint64 `json:"usedDiskSize"`
	UsedDiskSizeHuman  string `json:"usedDiskSizeHuman"`
	CollectionCount    int    `json:"collectionCount"`
}

// CollectionsResult represents the full penumbra data
type CollectionsResult struct {
	Collections []Collection `json:"collections"`
	Mods        []Mod        `json:"mods"`
	Stats       Stats        `json:"stats"`
}

// BackupItem represents a mod to be backed up
type BackupItem struct {
	Mod        Mod    `json:"mod"`
	FilteredBy string `json:"filteredBy,omitempty"`
	IsFiltered bool   `json:"isFiltered"`
}

// BackupValidation represents the backup preview
type BackupValidation struct {
	Items               []BackupItem `json:"items"`
	TotalSize           uint64       `json:"totalSize"`
	TotalSizeHuman      string       `json:"totalSizeHuman"`
	EstimatedSize       uint64       `json:"estimatedSize"`
	EstimatedSizeHuman  string       `json:"estimatedSizeHuman"`
	AvailableSpace      uint64       `json:"availableSpace"`
	AvailableSpaceHuman string       `json:"availableSpaceHuman"`
	HasEnoughSpace      bool         `json:"hasEnoughSpace"`
}

// BackupProgress represents backup progress updates
type BackupProgress struct {
	Percent float64 `json:"percent"`
	Current string  `json:"current"`
	Done    bool    `json:"done"`
	Error   string  `json:"error,omitempty"`
}

// BackupResult represents the backup operation result
type BackupResult struct {
	OutputPath     string `json:"outputPath"`
	OriginalSize   uint64 `json:"originalSize"`
	CompressedSize uint64 `json:"compressedSize"`
	Ratio          string `json:"ratio"`
}
