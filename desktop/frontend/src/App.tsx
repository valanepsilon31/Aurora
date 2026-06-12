import { useState, useEffect, useMemo, useRef } from 'react'

// Wails runtime bindings
declare global {
  interface Window {
    go: {
      main: {
        App: {
          GetConfig: () => Promise<ConfigResult>
          ReloadConfig: () => Promise<ConfigResult>
          UpdateConfig: (penumbraPath: string, modsPath: string, outputPath: string) => Promise<void>
          IsConfigValid: () => Promise<boolean>
          AddFilter: (filter: string) => Promise<void>
          RemoveFilter: (filter: string) => Promise<void>
          AddInclusion: (inclusion: string) => Promise<void>
          RemoveInclusion: (inclusion: string) => Promise<void>
          SetConcurrency: (concurrency: number) => Promise<void>
          SetCompression: (compression: string) => Promise<void>
          GetFilterMatches: () => Promise<FilterMatches>
          OpenOutputFolder: () => Promise<void>
          GetCollections: () => Promise<CollectionsResult>
          ValidateBackup: () => Promise<BackupValidation>
          RunBackup: (threads: number) => Promise<BackupResult>
          BrowseDirectory: (title: string, defaultPath: string) => Promise<string>
          GetVersion: () => Promise<string>
        }
      }
    }
    runtime: {
      EventsOn: (eventName: string, callback: (data: unknown) => void) => () => void
      EventsOff: (eventName: string) => void
    }
  }
}

interface ConfigResult {
  penumbraPath: string
  modsPath: string
  outputPath: string
  filters: string[]
  inclusions: string[]
  concurrency: number
  compression: string
  status: {
    valid: boolean
    penumbraStatus: string
    modsStatus: string
    outputStatus: string
  }
}

interface FilterMatches {
  filters: Record<string, number>
  inclusions: Record<string, number>
}

// Count chip next to a filter. undefined = count not computed yet (initial
// load, or a freshly added pattern): show a counting hint, never a fake 0.
function FilterCount({ count }: { count: number | undefined }) {
  if (count === undefined) {
    return <span className="filter-count">(…)</span>
  }
  return (
    <span className={`filter-count ${count === 0 ? 'filter-count-zero' : ''}`}>
      ({count})
    </span>
  )
}

interface Mod {
  name: string
  path: string
  size: number
  sizeHuman: string
  collections: string[]
}

interface Collection {
  name: string
  mods: Mod[]
}

interface Stats {
  totalMods: number
  usedMods: number
  unusedMods: number
  totalDiskSize: number
  totalDiskSizeHuman: string
  usedDiskSize: number
  usedDiskSizeHuman: string
  collectionCount: number
}

interface CollectionsResult {
  collections: Collection[]
  mods: Mod[]
  stats: Stats
}

interface BackupItem {
  mod: Mod
  filteredBy: string
  isFiltered: boolean
  includedBy: string
  isIncluded: boolean
}

interface BackupValidation {
  items: BackupItem[]
  totalSize: number
  totalSizeHuman: string
  estimatedSize: number
  estimatedSizeHuman: string
  availableSpace: number
  availableSpaceHuman: string
  hasEnoughSpace: boolean
}

interface BackupResult {
  outputPath: string
  originalSize: number
  compressedSize: number
  ratio: string
}

interface BackupProgress {
  percent: number
  current: string
  done: boolean
  error?: string
}

type Tab = 'config' | 'collections' | 'backup'

function App() {
  const [activeTab, setActiveTab] = useState<Tab>('config')
  const [config, setConfig] = useState<ConfigResult | null>(null)
  const [collections, setCollections] = useState<CollectionsResult | null>(null)
  const [backup, setBackup] = useState<BackupValidation | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Config form state
  const [penumbraPath, setPenumbraPath] = useState('')
  const [modsPath, setModsPath] = useState('')
  const [outputPath, setOutputPath] = useState('')
  const [isEditing, setIsEditing] = useState(false)

  // Backup state
  const [backupRunning, setBackupRunning] = useState(false)
  const [backupProgress, setBackupProgress] = useState<BackupProgress | null>(null)
  const [backupResult, setBackupResult] = useState<BackupResult | null>(null)
  const [backupError, setBackupError] = useState<string | null>(null)

  // Expanded collections
  const [expandedCollections, setExpandedCollections] = useState<Set<string>>(new Set())

  // Version
  const [version, setVersion] = useState('')

  useEffect(() => {
    loadConfig()
    window.go.main.App.GetVersion().then(setVersion)
  }, [])

  // Listen for backup progress events
  useEffect(() => {
    if (window.runtime?.EventsOn) {
      const unsubscribe = window.runtime.EventsOn('backup:progress', (data) => {
        const progress = data as BackupProgress
        setBackupProgress(progress)
        if (progress.error) {
          setBackupError(progress.error)
        }
        if (progress.done) {
          setBackupRunning(false)
        }
      })
      return () => unsubscribe()
    }
  }, [])

  useEffect(() => {
    if (activeTab === 'config' && config?.status.valid && !collections) {
      // The filter autocomplete (mod/collection suggestions) is built from
      // collections data, so the Config tab needs it too — but cached
      loadCollections()
    } else if (activeTab === 'collections' && config?.status.valid) {
      loadCollections()
    } else if (activeTab === 'backup' && config?.status.valid) {
      loadBackup()
    }
  }, [activeTab, config?.status.valid])

  const loadConfig = async (reload = false) => {
    try {
      setLoading(true)
      setError(null)
      const cfg = reload
        ? await window.go.main.App.ReloadConfig()
        : await window.go.main.App.GetConfig()
      setConfig(cfg)
      setPenumbraPath(cfg.penumbraPath)
      setModsPath(cfg.modsPath)
      setOutputPath(cfg.outputPath)
      // Clear collections cache when config changes to force refresh
      if (reload) {
        setCollections(null)
        setBackup(null)
      }
    } catch (err) {
      setError(`Failed to load config: ${err}`)
    } finally {
      setLoading(false)
    }
  }

  const saveConfig = async () => {
    try {
      setLoading(true)
      setError(null)
      await window.go.main.App.UpdateConfig(penumbraPath, modsPath, outputPath)
      await loadConfig(true) // Reload from disk after save
      setIsEditing(false)
    } catch (err) {
      setError(`Failed to save config: ${err}`)
    } finally {
      setLoading(false)
    }
  }

  const loadCollections = async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await window.go.main.App.GetCollections()
      setCollections(data)
    } catch (err) {
      setError(`Failed to load collections: ${err}`)
    } finally {
      setLoading(false)
    }
  }

  const loadBackup = async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await window.go.main.App.ValidateBackup()
      setBackup(data)
    } catch (err) {
      setError(`Failed to validate backup: ${err}`)
    } finally {
      setLoading(false)
    }
  }

  const [showBackupModal, setShowBackupModal] = useState(false)

  const runBackup = async () => {
    try {
      setBackupRunning(true)
      setShowBackupModal(true)
      setError(null)
      setBackupResult(null)
      setBackupError(null)
      setBackupProgress({ percent: 0, current: 'Starting...', done: false })
      const result = await window.go.main.App.RunBackup(config?.concurrency ?? 0)
      setBackupResult(result)
      setBackupProgress({ percent: 100, current: 'Complete!', done: true })
    } catch (err) {
      // Keep the modal open and show the error there
      setBackupError(String(err))
    } finally {
      setBackupRunning(false)
    }
  }

  const closeBackupModal = () => {
    setShowBackupModal(false)
    setBackupProgress(null)
    setBackupError(null)
  }

  // Escape closes the backup modal once it is no longer running
  useEffect(() => {
    if (!showBackupModal || backupRunning) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') closeBackupModal()
    }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [showBackupModal, backupRunning])

  const toggleCollection = (name: string) => {
    const newExpanded = new Set(expandedCollections)
    if (newExpanded.has(name)) {
      newExpanded.delete(name)
    } else {
      newExpanded.add(name)
    }
    setExpandedCollections(newExpanded)
  }

  return (
    <div className="app">
      {/* Backup Modal - Progress or Result */}
      {showBackupModal && (
        <div className="overlay">
          <div className="progress-modal">
            {backupResult ? (
              <>
                <div className="progress-icon success">✓</div>
                <h3 className="progress-title">Backup Complete</h3>
                <div className="result-summary">
                  <div className="result-row">
                    <span className="result-label">File</span>
                    <span className="result-value">{backupResult.outputPath}</span>
                  </div>
                  <div className="result-row">
                    <span className="result-label">Compression</span>
                    <span className="result-value">{backupResult.ratio}</span>
                  </div>
                </div>
                <div className="modal-actions">
                  <button className="btn btn-secondary" onClick={() => window.go.main.App.OpenOutputFolder()}>Open folder</button>
                  <button className="btn" onClick={closeBackupModal}>OK</button>
                </div>
              </>
            ) : backupError ? (
              <>
                <div className="progress-icon error">⚠</div>
                <h3 className="progress-title">Backup Failed</h3>
                <p className="progress-error">{backupError}</p>
                <button className="btn" onClick={closeBackupModal}>Close</button>
              </>
            ) : (
              <>
                <div className="progress-icon">📦</div>
                <h3 className="progress-title">Creating Backup</h3>
                <p className="progress-subtitle">Please wait while your mods are being compressed...</p>
                <div className="progress-bar-container">
                  <div className="progress-bar">
                    <div
                      className="progress-bar-fill"
                      style={{ width: `${backupProgress?.percent || 0}%` }}
                    />
                  </div>
                </div>
                <div className="progress-percent">{Math.round(backupProgress?.percent || 0)}%</div>
                <p className="progress-status">{backupProgress?.current || 'Processing...'}</p>
              </>
            )}
          </div>
        </div>
      )}

      <header className="header">
        <div className="header-icon">
          <svg viewBox="0 0 1024 1024" width="24" height="24">
            <defs>
              <linearGradient id="bgGrad" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" stopColor="#0b0f14"/>
                <stop offset="100%" stopColor="#111921"/>
              </linearGradient>
              <linearGradient id="letterGrad" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" stopColor="#0ea5e9"/>
                <stop offset="100%" stopColor="#38bdf8"/>
              </linearGradient>
            </defs>
            <rect x="64" y="64" width="896" height="896" rx="128" ry="128" fill="url(#bgGrad)"/>
            <rect x="64" y="64" width="896" height="896" rx="128" ry="128" fill="none" stroke="rgba(255,255,255,0.1)" strokeWidth="4"/>
            <text x="512" y="720" fontFamily="Inter, sans-serif" fontSize="600" fontWeight="700" textAnchor="middle" fill="url(#letterGrad)">A</text>
          </svg>
        </div>
        <h1>Aurora</h1>
        {version && <span className="header-version">{version}</span>}
        <span className="header-tagline">Smart mod backups powered by your Penumbra collections</span>
      </header>

      <nav className="tabs">
        <button
          className={`tab ${activeTab === 'config' ? 'active' : ''}`}
          onClick={() => setActiveTab('config')}
        >
          ⚙️ Config
        </button>
        <button
          className={`tab ${activeTab === 'collections' ? 'active' : ''}`}
          onClick={() => setActiveTab('collections')}
          disabled={!config?.status.valid}
        >
          📚 Collections
        </button>
        <button
          className={`tab ${activeTab === 'backup' ? 'active' : ''}`}
          onClick={() => setActiveTab('backup')}
          disabled={!config?.status.valid}
        >
          💾 Backup
        </button>
      </nav>

      <main className="content">
        {error && <div className="error-message">{error}</div>}

        {activeTab === 'config' && (
          <ConfigTab
            config={config}
            loading={loading}
            isEditing={isEditing}
            penumbraPath={penumbraPath}
            modsPath={modsPath}
            outputPath={outputPath}
            filterSuggestions={[...new Set([
              ...(collections?.collections.map(c => c.name) || []),
              ...(collections?.mods.map(m => m.name) || [])
            ])]}
            setPenumbraPath={setPenumbraPath}
            setModsPath={setModsPath}
            setOutputPath={setOutputPath}
            setIsEditing={setIsEditing}
            saveConfig={saveConfig}
            reloadConfig={() => loadConfig(true)}
            addFilter={async (filter) => {
              await window.go.main.App.AddFilter(filter)
              await loadConfig() // no reload: keeps the collections cache (autocomplete)
              setBackup(null)
            }}
            removeFilter={async (filter) => {
              await window.go.main.App.RemoveFilter(filter)
              await loadConfig() // no reload: keeps the collections cache (autocomplete)
              setBackup(null)
            }}
            addInclusion={async (inclusion) => {
              await window.go.main.App.AddInclusion(inclusion)
              await loadConfig() // no reload: keeps the collections cache (autocomplete)
              setBackup(null)
            }}
            removeInclusion={async (inclusion) => {
              await window.go.main.App.RemoveInclusion(inclusion)
              await loadConfig() // no reload: keeps the collections cache (autocomplete)
              setBackup(null)
            }}
            setConcurrency={async (concurrency) => {
              await window.go.main.App.SetConcurrency(concurrency)
              await loadConfig()
            }}
            setCompression={async (compression) => {
              await window.go.main.App.SetCompression(compression)
              await loadConfig()
            }}
          />
        )}

        {activeTab === 'collections' && (
          <CollectionsTab
            collections={collections}
            loading={loading}
            expandedCollections={expandedCollections}
            toggleCollection={toggleCollection}
          />
        )}

        {activeTab === 'backup' && (
          <BackupTab
            backup={backup}
            loading={loading}
            backupRunning={backupRunning}
            runBackup={runBackup}
          />
        )}
      </main>
    </div>
  )
}

interface ConfigTabProps {
  config: ConfigResult | null
  loading: boolean
  isEditing: boolean
  penumbraPath: string
  modsPath: string
  outputPath: string
  filterSuggestions: string[]
  setPenumbraPath: (path: string) => void
  setModsPath: (path: string) => void
  setOutputPath: (path: string) => void
  setIsEditing: (editing: boolean) => void
  saveConfig: () => void
  reloadConfig: () => void
  addFilter: (filter: string) => Promise<void>
  removeFilter: (filter: string) => Promise<void>
  addInclusion: (inclusion: string) => Promise<void>
  removeInclusion: (inclusion: string) => Promise<void>
  setConcurrency: (concurrency: number) => Promise<void>
  setCompression: (compression: string) => Promise<void>
}

function ConfigTab({
  config,
  loading,
  isEditing,
  penumbraPath,
  modsPath,
  outputPath,
  filterSuggestions,
  setPenumbraPath,
  setModsPath,
  setOutputPath,
  setIsEditing,
  saveConfig,
  reloadConfig,
  addFilter,
  removeFilter,
  addInclusion,
  removeInclusion,
  setConcurrency,
  setCompression,
}: ConfigTabProps) {
  const [newFilter, setNewFilter] = useState('')
  const [suggestOpen, setSuggestOpen] = useState(false)
  const [filterMatches, setFilterMatches] = useState<FilterMatches | null>(null)

  // Suggestions matching the typed text, capped: mounting the full
  // mod/collection list (1000+ options) freezes the webview's datalist popup
  const visibleSuggestions = useMemo(() => {
    const query = newFilter.trim().toLowerCase()
    if (!query) return []
    return filterSuggestions
      .filter((name) => name.toLowerCase().includes(query))
      .slice(0, 20)
  }, [newFilter, filterSuggestions])
  const [concurrencyValue, setConcurrencyValue] = useState(config?.concurrency ?? 0)

  // Load per-filter match counts (a 0 means a dead filter). Silent on
  // failure: counts are a hint, not required for the page to work.
  useEffect(() => {
    if (!config?.status.valid) {
      setFilterMatches(null)
      return
    }
    let cancelled = false
    window.go.main.App.GetFilterMatches()
      .then((m) => { if (!cancelled) setFilterMatches(m) })
      .catch(() => { if (!cancelled) setFilterMatches(null) })
    return () => { cancelled = true }
  }, [config?.status.valid, config?.filters, config?.inclusions])
  const [compressionValue, setCompressionValue] = useState(config?.compression ?? 'normal')

  // Sync concurrency value when config loads
  useEffect(() => {
    if (config) {
      setConcurrencyValue(config.concurrency)
    }
  }, [config?.concurrency])

  // Sync compression value when config loads
  useEffect(() => {
    if (config) {
      setCompressionValue(config.compression || 'normal')
    }
  }, [config?.compression])

  const handleConcurrencyChange = async (value: number) => {
    setConcurrencyValue(value)
    await setConcurrency(value)
  }

  const handleCompressionChange = async (value: string) => {
    setCompressionValue(value)
    await setCompression(value)
  }

  const handleAddFilter = async () => {
    if (newFilter.trim()) {
      await addFilter(newFilter.trim())
      setNewFilter('')
    }
  }

  const handleAddInclusion = async () => {
    if (newFilter.trim()) {
      await addInclusion(newFilter.trim())
      setNewFilter('')
    }
  }

  // Enter adds an exclusion (the most common case); inclusion via its button
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleAddFilter()
    } else if (e.key === 'Escape') {
      setSuggestOpen(false)
    }
  }

  if (loading && !config) {
    return <div className="loading">Loading configuration...</div>
  }

  return (
    <div className="config-tab">
      <div className="card">
        <h2>Configuration</h2>

        {isEditing ? (
          <>
            <div className="field-label" style={{ marginBottom: '0.5rem', marginTop: '0.5rem' }}>Penumbra Path</div>
            <div className="input-with-browse">
              <input
                type="text"
                value={penumbraPath}
                onChange={(e) => setPenumbraPath(e.target.value)}
                placeholder="Path to Penumbra folder"
              />
              <button
                className="browse-btn"
                onClick={async () => {
                  const path = await window.go.main.App.BrowseDirectory('Select Penumbra Folder', penumbraPath)
                  if (path) setPenumbraPath(path)
                }}
                title="Browse..."
              >
                📁
              </button>
            </div>
            <div className="field-label" style={{ marginBottom: '0.5rem' }}>Mods Path</div>
            <div className="input-with-browse">
              <input
                type="text"
                value={modsPath}
                onChange={(e) => setModsPath(e.target.value)}
                placeholder="Path to mods folder"
              />
              <button
                className="browse-btn"
                onClick={async () => {
                  const path = await window.go.main.App.BrowseDirectory('Select Mods Folder', modsPath)
                  if (path) setModsPath(path)
                }}
                title="Browse..."
              >
                📁
              </button>
            </div>
            <div className="field-label" style={{ marginBottom: '0.5rem' }}>Backup Output Path (empty = app folder)</div>
            <div className="input-with-browse">
              <input
                type="text"
                value={outputPath}
                onChange={(e) => setOutputPath(e.target.value)}
                placeholder="Where backup archives are written"
              />
              <button
                className="browse-btn"
                onClick={async () => {
                  const path = await window.go.main.App.BrowseDirectory('Select Backup Output Folder', outputPath)
                  if (path) setOutputPath(path)
                }}
                title="Browse..."
              >
                📁
              </button>
            </div>
            <div className="actions">
              <button className="btn" onClick={saveConfig} disabled={loading}>
                Save Changes
              </button>
              <button className="btn btn-secondary" onClick={() => setIsEditing(false)}>
                Cancel
              </button>
            </div>
          </>
        ) : (
          <>
            <div className="field field-stacked">
              <div className="field-row">
                <span className="field-label">Penumbra</span>
                <span className="field-value field-with-status">
                  <span className="field-path">{config?.penumbraPath || '(not set)'}</span>
                  <span className={`field-status ${config?.status.penumbraStatus === 'OK' ? 'status-ok' : 'status-error'}`}>
                    {config?.status.penumbraStatus}
                  </span>
                </span>
              </div>
              <div className="field-row">
                <span className="field-label">Mods</span>
                <span className="field-value field-with-status">
                  <span className="field-path">{config?.modsPath || '(not set)'}</span>
                  <span className={`field-status ${config?.status.modsStatus === 'OK' ? 'status-ok' : 'status-error'}`}>
                    {config?.status.modsStatus}
                  </span>
                </span>
              </div>
              <div className="field-row">
                <span className="field-label">Output</span>
                <span className="field-value field-with-status">
                  <span className="field-path">{config?.outputPath || '(app folder)'}</span>
                  <span className={`field-status ${config?.status.outputStatus === 'OK' ? 'status-ok' : 'status-error'}`}>
                    {config?.status.outputStatus}
                  </span>
                </span>
              </div>
            </div>
            <div className="field">
              <span className="field-label">
                Concurrency
                <span className="help-badge tooltip-right" data-tooltip="Number of parallel threads for backup compression.&#10;&#10;0 = auto (uses all CPU cores).&#10;Warning: may slow down your machine significantly during backup.">?</span>
              </span>
              <span className="field-value field-inline">
                <input
                  type="number"
                  min="0"
                  max="32"
                  value={concurrencyValue}
                  onChange={(e) => handleConcurrencyChange(parseInt(e.target.value) || 0)}
                  className="concurrency-input"
                />
                <span className="field-label">
                  Compression
                  <span className="help-badge tooltip-right" data-tooltip="Normal: fast backups (default).&#10;Max: smallest backups, about twice as slow for ~5% smaller files.">?</span>
                </span>
                <select
                  value={compressionValue}
                  onChange={(e) => handleCompressionChange(e.target.value)}
                  className="compression-select"
                >
                  <option value="normal">Normal</option>
                  <option value="max">Max</option>
                </select>
              </span>
            </div>
            <div className="actions">
              <button className="btn" onClick={() => setIsEditing(true)}>
                Edit Configuration
              </button>
              <button className="btn btn-secondary" onClick={reloadConfig}>
                Refresh
              </button>
            </div>
          </>
        )}
      </div>

      {config?.status.valid && (
        <div className="card card-filters">
          <h2>
            Backup Filters
            <span className="help-badge tooltip-right" data-tooltip="Filters match by prefix, case-insensitive:&#10;'shadow' matches mod 'ShadowKnight' and collection 'Shadow-Pack'.&#10;Checked against mod names, paths and collection names.">?</span>
          </h2>

          <div className="filter-add">
            {/* Custom suggestions popup instead of a native datalist: the
                native popup is unstylable (tiny and light-themed on Windows)
                and suggestions are already filtered in JS anyway */}
            <div className="filter-input-wrap">
              <input
                type="text"
                value={newFilter}
                onChange={(e) => { setNewFilter(e.target.value); setSuggestOpen(true) }}
                onKeyDown={handleKeyDown}
                onFocus={() => setSuggestOpen(true)}
                onBlur={() => setSuggestOpen(false)}
                placeholder="Enter filter pattern or select a mod/collection..."
              />
              {suggestOpen && visibleSuggestions.length > 0 && (
                <div className="suggestions-popup">
                  {visibleSuggestions.map((name) => (
                    <button
                      key={name}
                      type="button"
                      className="suggestion-item"
                      // mousedown (not click): fires before the input blur
                      // hides the popup
                      onMouseDown={(e) => {
                        e.preventDefault()
                        setNewFilter(name)
                        setSuggestOpen(false)
                      }}
                    >
                      {name}
                    </button>
                  ))}
                </div>
              )}
            </div>
            <button className="btn" onClick={handleAddFilter} disabled={!newFilter.trim()}>
              + Exclusion
            </button>
            <button className="btn" onClick={handleAddInclusion} disabled={!newFilter.trim()}>
              + Inclusion
            </button>
          </div>

          <div className="filter-columns">
            <div className="filter-column">
              <h3>
                Exclusions
                <span className="help-badge tooltip-right" data-tooltip="Matching mods are excluded from backups,&#10;even when collections use them">?</span>
              </h3>
              <div className="filter-list">
                {config?.filters && config.filters.length > 0 ? (
                  config.filters.map((filter) => (
                    <div key={filter} className="filter-item">
                      <span className="filter-text">{filter}</span>
                      <FilterCount count={filterMatches?.filters[filter]} />
                      <button className="filter-delete" onClick={() => removeFilter(filter)} title="Remove exclusion">
                        ×
                      </button>
                    </div>
                  ))
                ) : (
                  <div className="filter-empty">No exclusions configured</div>
                )}
              </div>
            </div>
            <div className="filter-column">
              <h3>
                Inclusions
                <span className="help-badge tooltip-right" data-tooltip="Matching mods are always backed up, overriding&#10;exclusions and missing collections">?</span>
              </h3>
              <div className="filter-list">
                {config?.inclusions && config.inclusions.length > 0 ? (
                  config.inclusions.map((inclusion) => (
                    <div key={inclusion} className="filter-item">
                      <span className="filter-text">{inclusion}</span>
                      <FilterCount count={filterMatches?.inclusions[inclusion]} />
                      <button className="filter-delete" onClick={() => removeInclusion(inclusion)} title="Remove inclusion">
                        ×
                      </button>
                    </div>
                  ))
                ) : (
                  <div className="filter-empty">No inclusions configured</div>
                )}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

// Dropdown filter menu: a button showing the active filter count, opening a
// panel of checkboxes. Selections apply immediately; OK or clicking outside
// closes the panel.
interface FilterOption {
  id: string
  label: string
}

interface FilterMenuProps {
  options: FilterOption[]
  active: Set<string>
  onChange: (next: Set<string>) => void
}

function FilterMenu({ options, active, onChange }: FilterMenuProps) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  // Close when clicking outside the menu or pressing Escape
  useEffect(() => {
    if (!open) return
    const onClickOutside = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setOpen(false)
    }
    document.addEventListener('mousedown', onClickOutside)
    document.addEventListener('keydown', onKey)
    return () => {
      document.removeEventListener('mousedown', onClickOutside)
      document.removeEventListener('keydown', onKey)
    }
  }, [open])

  const toggle = (id: string) => {
    const next = new Set(active)
    if (next.has(id)) {
      next.delete(id)
    } else {
      next.add(id)
    }
    onChange(next)
  }

  const label = active.size === 0 ? 'Filters' : `${active.size} filter${active.size > 1 ? 's' : ''}`

  return (
    <div className="filter-menu" ref={ref}>
      <button
        className={`filter-menu-btn ${active.size > 0 ? 'filter-menu-btn-active' : ''}`}
        onClick={() => setOpen(!open)}
      >
        {label} <span className="filter-menu-arrow">▾</span>
      </button>
      {open && (
        <div className="filter-menu-panel">
          {options.map((opt) => (
            <label key={opt.id} className="checkbox-filter checkbox-filter-compact filter-menu-option">
              <input
                type="checkbox"
                checked={active.has(opt.id)}
                onChange={() => toggle(opt.id)}
              />
              <span>{opt.label}</span>
            </label>
          ))}
          <button className="btn filter-menu-ok" onClick={() => setOpen(false)}>
            OK
          </button>
        </div>
      )}
    </div>
  )
}

// Reusable search input component
interface SearchInputProps {
  value: string
  onChange: (value: string) => void
  placeholder?: string
  resultCount?: number
  totalCount?: number
}

function SearchInput({ value, onChange, placeholder = "Search...", resultCount, totalCount }: SearchInputProps) {
  return (
    <div className="search-container">
      <input
        type="text"
        className="search-input"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
      />
      {value && resultCount !== undefined && totalCount !== undefined && (
        <div className="search-results">
          Showing {resultCount} of {totalCount} results
        </div>
      )}
    </div>
  )
}

// Filter helper function
function matchesSearch(text: string, search: string): boolean {
  return text.toLowerCase().includes(search.toLowerCase())
}

interface CollectionsTabProps {
  collections: CollectionsResult | null
  loading: boolean
  expandedCollections: Set<string>
  toggleCollection: (name: string) => void
}

function CollectionsTab({ collections, loading, expandedCollections, toggleCollection }: CollectionsTabProps) {
  const [search, setSearch] = useState('')

  const filteredCollections = useMemo(() => {
    if (!collections || !search.trim()) {
      return collections?.collections || []
    }

    return collections.collections.filter(col => {
      // Match collection name
      if (matchesSearch(col.name, search)) return true
      // Match any mod name in collection
      return col.mods.some(mod => matchesSearch(mod.name, search))
    })
  }, [collections, search])

  if (loading && !collections) {
    return <div className="loading">Loading collections...</div>
  }

  if (!collections) {
    return null
  }

  return (
    <>
      <div className="stats">
        <div className="stat-card">
          <div className="stat-value">{collections.stats.collectionCount}</div>
          <div className="stat-label">
            Collections
            <span className="help-badge" data-tooltip="Number of Penumbra collections found">?</span>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-value">{collections.stats.usedMods}/{collections.stats.totalMods}</div>
          <div className="stat-label">
            Mods Used/total
            <span className="help-badge" data-tooltip="Mods in at least one collection / all mods in folder">?</span>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-value">{collections.stats.usedDiskSizeHuman}/{collections.stats.totalDiskSizeHuman}</div>
          <div className="stat-label">
            Space Used/total
            <span className="help-badge tooltip-left" data-tooltip="Disk space of used mods / all mods">?</span>
          </div>
        </div>
      </div>

      <div className="card">
        <h2>Collections</h2>
        <SearchInput
          value={search}
          onChange={setSearch}
          placeholder="Search collections or mods..."
          resultCount={filteredCollections.length}
          totalCount={collections.collections.length}
        />
        <div className="collections-list">
          {filteredCollections.map((col) => (
            <div
              key={col.name}
              className={`collection ${expandedCollections.has(col.name) ? 'expanded' : ''}`}
            >
              <div className="collection-header" onClick={() => toggleCollection(col.name)}>
                <span className="collection-name">{col.name}</span>
                <span className="collection-count">{col.mods.length} mods</span>
              </div>
              {expandedCollections.has(col.name) && (
                <div className="collection-mods">
                  {col.mods.map((mod) => (
                    <div key={mod.name} className="mod-item">
                      <span className="mod-name">{mod.name}</span>
                      <span className="mod-size">{mod.sizeHuman}</span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </>
  )
}

interface BackupTabProps {
  backup: BackupValidation | null
  loading: boolean
  backupRunning: boolean
  runBackup: () => void
}

function BackupTab({ backup, loading, backupRunning, runBackup }: BackupTabProps) {
  const [search, setSearch] = useState('')
  const [activeFilters, setActiveFilters] = useState<Set<string>>(new Set())

  const filteredItems = useMemo(() => {
    if (!backup) return []

    let items = backup.items

    // Filter menu (selected = keep matching items; several = union)
    if (activeFilters.size > 0) {
      items = items.filter(item =>
        (activeFilters.has('exclusions') && item.isFiltered) ||
        (activeFilters.has('inclusions') && item.isIncluded)
      )
    }

    // Filter by search
    if (search.trim()) {
      items = items.filter(item => {
        if (matchesSearch(item.mod.name, search)) return true
        return item.mod.collections.some(col => matchesSearch(col, search))
      })
    }

    return items
  }, [backup, search, activeFilters])

  if (loading && !backup) {
    return <div className="loading">Loading backup info...</div>
  }

  if (!backup) {
    return null
  }

  const modsToBackup = backup.items.filter(i => !i.isFiltered).length

  const filteredCount = backup.items.filter(i => i.isFiltered).length

  return (
    <div className="backup-tab">
      <div className="stats">
        <div className="stat-card">
          <div className="stat-value">{backup.items.length}</div>
          <div className="stat-label">
            Total Used Mods
            <span className="help-badge" data-tooltip="Mods in at least one collection, or matched by an inclusion filter">?</span>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-value">{modsToBackup}</div>
          <div className="stat-label">
            To Backup
            <span className="help-badge" data-tooltip="Mods that will be included in backup">?</span>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-value">{filteredCount}</div>
          <div className="stat-label">
            Filtered
            <span className="help-badge tooltip-left" data-tooltip="Mods excluded by your exclusion filters">?</span>
          </div>
        </div>
        <div className={`stat-card ${!backup.hasEnoughSpace ? 'stat-card-warning' : ''}`}>
          <div className="stat-value">{backup.estimatedSizeHuman} / {backup.availableSpaceHuman}</div>
          <div className="stat-label">
            Est. / Available
            <span className="help-badge tooltip-left" data-tooltip="Estimated backup size vs available disk space">?</span>
          </div>
        </div>
      </div>

      <div className="card card-backup">
        <h2>Mods to Backup</h2>
        <div className="search-row">
          <SearchInput
            value={search}
            onChange={setSearch}
            placeholder="Search mods or collections..."
            resultCount={filteredItems.length}
            totalCount={backup.items.length}
          />
          <FilterMenu
            options={[
              { id: 'exclusions', label: 'Exclusions' },
              { id: 'inclusions', label: 'Inclusions' },
            ]}
            active={activeFilters}
            onChange={setActiveFilters}
          />
        </div>
        <div className="backup-list">
          {filteredItems.map((item, index) => (
            <div key={item.mod.path || index} className={`backup-item ${item.isFiltered ? 'filtered' : ''}`}>
              <span className="mod-name">
                {item.mod.name}
                {item.isFiltered && <span style={{ color: 'var(--text-muted)', marginLeft: '0.5rem' }}>(filter exclusion: {item.filteredBy})</span>}
                {!item.isFiltered && item.isIncluded && <span style={{ color: 'var(--text-muted)', marginLeft: '0.5rem' }}>(filter inclusion: {item.includedBy})</span>}
              </span>
              <span className="mod-size">{item.mod.sizeHuman}</span>
            </div>
          ))}
        </div>
        <div className="actions">
          <button className="btn" onClick={runBackup} disabled={backupRunning || modsToBackup === 0 || !backup.hasEnoughSpace}>
            {backupRunning ? 'Running...' : `Backup ${modsToBackup} Mods`}
          </button>
          {!backup.hasEnoughSpace && (
            <span className="warning-text">Not enough disk space</span>
          )}
        </div>
      </div>
    </div>
  )
}

export default App
