package aurora

// ConfigResult represents the current configuration state
type ConfigResult struct {
	PenumbraPath string       `json:"penumbraPath"`
	ModsPath     string       `json:"modsPath"`
	OutputPath   string       `json:"outputPath"`
	Filters      []string     `json:"filters"`
	Inclusions   []string     `json:"inclusions"`
	Concurrency  int          `json:"concurrency"`
	Compression  string       `json:"compression"`
	Status       ConfigStatus `json:"status"`
}

// ConfigStatus represents validation status of paths
type ConfigStatus struct {
	Valid          bool   `json:"valid"`
	PenumbraStatus string `json:"penumbraStatus"`
	ModsStatus     string `json:"modsStatus"`
	OutputStatus   string `json:"outputStatus"`
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
	FilteredBy string `json:"filteredBy,omitempty"` // Matching exclusion filter
	IsFiltered bool   `json:"isFiltered"`
	IncludedBy string `json:"includedBy,omitempty"` // Matching inclusion filter (mod has no collections)
	IsIncluded bool   `json:"isIncluded"`
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

// FilterMatches reports per-pattern mod match counts for the config filters.
// Inclusions carries decisive matches (mods the inclusion adds or rescues);
// InclusionsAny counts every match, including mods already backed up via
// collections - so the UI can tell "redundant" apart from "dead".
type FilterMatches struct {
	Filters       map[string]int `json:"filters"`
	Inclusions    map[string]int `json:"inclusions"`
	InclusionsAny map[string]int `json:"inclusionsAny"`
}
