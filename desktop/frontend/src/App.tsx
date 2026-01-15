import { useState, useEffect, useMemo } from 'react'

// Wails runtime bindings
declare global {
  interface Window {
    go: {
      main: {
        App: {
          GetConfig: () => Promise<ConfigResult>
          UpdateConfig: (penumbraPath: string, modsPath: string) => Promise<void>
          IsConfigValid: () => Promise<boolean>
          AddFilter: (filter: string) => Promise<void>
          RemoveFilter: (filter: string) => Promise<void>
          SetConcurrency: (concurrency: number) => Promise<void>
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
  configFile: string
  penumbraPath: string
  modsPath: string
  filters: string[]
  concurrency: number
  status: {
    valid: boolean
    penumbraStatus: string
    modsStatus: string
  }
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
}

interface BackupValidation {
  items: BackupItem[]
  totalSize: number
  totalSizeHuman: string
  estimatedSize: number
  estimatedSizeHuman: string
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
  const [isEditing, setIsEditing] = useState(false)

  // Backup state
  const [backupRunning, setBackupRunning] = useState(false)
  const [backupProgress, setBackupProgress] = useState<BackupProgress | null>(null)
  const [backupResult, setBackupResult] = useState<BackupResult | null>(null)

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
        if (progress.done) {
          setBackupRunning(false)
        }
      })
      return () => unsubscribe()
    }
  }, [])

  useEffect(() => {
    if (activeTab === 'config' && config?.status.valid && !collections) {
      loadCollections()
    } else if (activeTab === 'collections' && config?.status.valid) {
      loadCollections()
    } else if (activeTab === 'backup' && config?.status.valid) {
      loadBackup()
    }
  }, [activeTab, config?.status.valid])

  const loadConfig = async () => {
    try {
      setLoading(true)
      setError(null)
      const cfg = await window.go.main.App.GetConfig()
      setConfig(cfg)
      setPenumbraPath(cfg.penumbraPath)
      setModsPath(cfg.modsPath)
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
      await window.go.main.App.UpdateConfig(penumbraPath, modsPath)
      await loadConfig()
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
      setBackupProgress({ percent: 0, current: 'Starting...', done: false })
      const result = await window.go.main.App.RunBackup(config?.concurrency ?? 0)
      setBackupResult(result)
      setBackupProgress({ percent: 100, current: 'Complete!', done: true })
    } catch (err) {
      setError(`Backup failed: ${err}`)
      setShowBackupModal(false)
      setBackupProgress(null)
    } finally {
      setBackupRunning(false)
    }
  }

  const closeBackupModal = () => {
    setShowBackupModal(false)
    setBackupProgress(null)
  }

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
            {backupRunning ? (
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
            ) : backupResult ? (
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
                <button className="btn" onClick={closeBackupModal}>OK</button>
              </>
            ) : null}
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
            filterSuggestions={[...new Set([
              ...(collections?.collections.map(c => c.name) || []),
              ...(collections?.mods.map(m => m.name) || [])
            ])]}
            setPenumbraPath={setPenumbraPath}
            setModsPath={setModsPath}
            setIsEditing={setIsEditing}
            saveConfig={saveConfig}
            loadConfig={loadConfig}
            addFilter={async (filter) => {
              await window.go.main.App.AddFilter(filter)
              await loadConfig()
              setBackup(null)
            }}
            removeFilter={async (filter) => {
              await window.go.main.App.RemoveFilter(filter)
              await loadConfig()
              setBackup(null)
            }}
            setConcurrency={async (concurrency) => {
              await window.go.main.App.SetConcurrency(concurrency)
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
  filterSuggestions: string[]
  setPenumbraPath: (path: string) => void
  setModsPath: (path: string) => void
  setIsEditing: (editing: boolean) => void
  saveConfig: () => void
  loadConfig: () => void
  addFilter: (filter: string) => Promise<void>
  removeFilter: (filter: string) => Promise<void>
  setConcurrency: (concurrency: number) => Promise<void>
}

function ConfigTab({
  config,
  loading,
  isEditing,
  penumbraPath,
  modsPath,
  filterSuggestions,
  setPenumbraPath,
  setModsPath,
  setIsEditing,
  saveConfig,
  loadConfig,
  addFilter,
  removeFilter,
  setConcurrency,
}: ConfigTabProps) {
  const [newFilter, setNewFilter] = useState('')
  const [concurrencyValue, setConcurrencyValue] = useState(config?.concurrency ?? 0)

  // Sync concurrency value when config loads
  useEffect(() => {
    if (config) {
      setConcurrencyValue(config.concurrency)
    }
  }, [config?.concurrency])

  const handleConcurrencyChange = async (value: number) => {
    setConcurrencyValue(value)
    await setConcurrency(value)
  }

  const handleAddFilter = async () => {
    if (newFilter.trim()) {
      await addFilter(newFilter.trim())
      setNewFilter('')
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleAddFilter()
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
            <div className="field">
              <span className="field-label">Config File</span>
              <span className="field-value">{config?.configFile}</span>
            </div>
            <div className="field">
              <span className="field-label">Penumbra Path</span>
              <span className="field-value">{config?.penumbraPath || '(not set)'}</span>
            </div>
            <div className="field">
              <span className="field-label">Penumbra Status</span>
              <span className={`field-value ${config?.status.penumbraStatus === 'OK' ? 'status-ok' : 'status-error'}`}>
                {config?.status.penumbraStatus}
              </span>
            </div>
            <div className="field">
              <span className="field-label">Mods Path</span>
              <span className="field-value">{config?.modsPath || '(not set)'}</span>
            </div>
            <div className="field">
              <span className="field-label">Mods Status</span>
              <span className={`field-value ${config?.status.modsStatus === 'OK' ? 'status-ok' : 'status-error'}`}>
                {config?.status.modsStatus}
              </span>
            </div>
            <div className="field">
              <span className="field-label">
                Concurrency
                <span className="help-badge tooltip-right" data-tooltip="Number of parallel threads for backup compression.&#10;&#10;0 = auto (uses all CPU cores).&#10;Higher values split the backup into multiple files.&#10;&#10;Warning: 0 (auto) will use all CPU cores and may&#10;slow down your machine significantly during backup.">?</span>
              </span>
              <span className="field-value">
                <input
                  type="number"
                  min="0"
                  max="32"
                  value={concurrencyValue}
                  onChange={(e) => handleConcurrencyChange(parseInt(e.target.value) || 0)}
                  className="concurrency-input"
                />
                
              </span>
            </div>
            <div className="actions">
              <button className="btn" onClick={() => setIsEditing(true)}>
                Edit Configuration
              </button>
              <button className="btn btn-secondary" onClick={loadConfig}>
                Refresh
              </button>
            </div>
          </>
        )}
      </div>

      {config?.status.valid && (
        <div className="card card-filters">
          <h2>Backup Filters</h2>
          <p className="filter-hint">Mods matching these patterns will be excluded from backups</p>

          <div className="filter-add">
            <input
              type="text"
              value={newFilter}
              onChange={(e) => setNewFilter(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Enter filter pattern or select a mod/collection..."
              list="filter-suggestions"
            />
            <datalist id="filter-suggestions">
              {filterSuggestions.map((name) => (
                <option key={name} value={name} />
              ))}
            </datalist>
            <button className="btn" onClick={handleAddFilter} disabled={!newFilter.trim()}>
              Add
            </button>
          </div>

          <div className="filter-list">
            {config?.filters && config.filters.length > 0 ? (
              config.filters.map((filter) => (
                <div key={filter} className="filter-item">
                  <span className="filter-text">{filter}</span>
                  <button className="filter-delete" onClick={() => removeFilter(filter)} title="Remove filter">
                    ×
                  </button>
                </div>
              ))
            ) : (
              <div className="filter-empty">No filters configured</div>
            )}
          </div>
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
            <span className="help-badge" data-tooltip="Disk space of used mods / all mods">?</span>
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
  const [showFilteredOnly, setShowFilteredOnly] = useState(false)

  const filteredItems = useMemo(() => {
    if (!backup) return []

    let items = backup.items

    // Filter by checkbox
    if (showFilteredOnly) {
      items = items.filter(item => item.isFiltered)
    }

    // Filter by search
    if (search.trim()) {
      items = items.filter(item => {
        if (matchesSearch(item.mod.name, search)) return true
        return item.mod.collections.some(col => matchesSearch(col, search))
      })
    }

    return items
  }, [backup, search, showFilteredOnly])

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
            <span className="help-badge" data-tooltip="Mods in at least one collection">?</span>
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
            <span className="help-badge" data-tooltip="Mods excluded by your filters">?</span>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-value">{backup.estimatedSizeHuman}</div>
          <div className="stat-label">
            Est. Size
            <span className="help-badge" data-tooltip="Estimated compressed backup size (~25%)">?</span>
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
          <label className="checkbox-filter">
            <input
              type="checkbox"
              checked={showFilteredOnly}
              onChange={(e) => setShowFilteredOnly(e.target.checked)}
            />
            <span>Show filtered only</span>
          </label>
        </div>
        <div className="backup-list">
          {filteredItems.map((item, index) => (
            <div key={item.mod.path || index} className={`backup-item ${item.isFiltered ? 'filtered' : ''}`}>
              <span className="mod-name">
                {item.mod.name}
                {item.isFiltered && <span style={{ color: 'var(--text-muted)', marginLeft: '0.5rem' }}>(filtered: {item.filteredBy})</span>}
              </span>
              <span className="mod-size">{item.mod.sizeHuman}</span>
            </div>
          ))}
        </div>
        <div className="actions">
          <button className="btn" onClick={runBackup} disabled={backupRunning || modsToBackup === 0}>
            {backupRunning ? 'Running...' : `Backup ${modsToBackup} Mods`}
          </button>
        </div>
      </div>
    </div>
  )
}

export default App
