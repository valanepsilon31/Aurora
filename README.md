# Aurora

[![Release](https://img.shields.io/github/v/release/valanepsilon31/Aurora?style=for-the-badge)](https://github.com/valanepsilon31/Aurora/releases)
[![Build](https://img.shields.io/github/actions/workflow/status/valanepsilon31/Aurora/build-and-release.yml?style=for-the-badge)](https://github.com/valanepsilon31/Aurora/actions)
[![License](https://img.shields.io/badge/License-MIT-blue?style=for-the-badge)](LICENSE)
[![Ko-Fi](https://img.shields.io/badge/Ko--fi-F16061?style=for-the-badge&logo=ko-fi&logoColor=white)](https://ko-fi.com/valara9930)


Aurora is a simple backup tool for [Penumbra](https://github.com/xivdev/Penumbra) mods. It scans your collections, finds which mods you're actually using, and backs them up into a several zip files.

## Important Notes

**Aurora backs up your mod files, not your game data.** For mods configs like penumbra, glamourer etc but also real ingame stuff like hotbars, macros, use [JustBackup](https://github.com/NightmareXIV/JustBackup) instead.

> **Disclaimer:** I built this tool for myself but wanted to share it with the community. It's provided as-is - use at your own risk. I'm not responsible for any data loss or issues that may occur. Always keep your own backups!

> **macOS:** The macOS builds (both Desktop and CLI) have not been tested. They may or may not work - feedback welcome!

---

## How to Use

### 1. Configure Your Paths

Set your Penumbra config folder, your mods folder, and (optionally) where backup archives are written — by default they land next to the app.

A few more knobs live here:

- **Concurrency** — how many threads compress in parallel (0 = all CPU cores).
- **Compression** — `Normal` (default, fast) or `Max` (smallest archives, about twice as slow for ~5% smaller files).

![Config Tab](docs/desktop-config.jpg)

### 2. Filter What Gets Backed Up

Backups include every mod referenced by at least one collection. Two filter lists refine that:

- **Exclusions** — matching mods are skipped. A mod is excluded when its name matches, or when *every* collection using it matches (so a mod still needed by another collection stays in).
- **Inclusions** — matching mods are *always* backed up, even when no collection references them. Inclusions win over exclusions.

Filters match by prefix, case-insensitive, with autocomplete from your mods and collections. Each filter chip shows how many mods it currently matches — a red `(0)` means the filter matches nothing.

### 3. Browse Your Collections

See all your Penumbra collections at a glance. Check which mods are in use and how much space they take.

![Collections Tab](docs/desktop-collections.jpg)

### 4. Create Your Backup

Preview what will be backed up — entries show why they're in or out (`filter exclusion: ...` / `filter inclusion: ...`), and the filter menu narrows the list to just those. Then hit the button.

![Backup Tab](docs/desktop-backup.jpg)

### 5. Watch the Progress

Progress is byte-accurate (no stalling at 50% then jumping to 100%), and once done, **Open folder** takes you straight to the archives.

![Progress](docs/desktop-progress.jpg)

---

## Download

Grab the latest release for your platform from the [Releases](https://github.com/valanepsilon31/Aurora/releases) page:

| Platform | Desktop App | CLI Only |
|----------|-------------|----------|
| Windows | `aurora-desktop-windows-amd64.zip` | `aurora-cli-windows-amd64.zip` |
| macOS (Apple Silicon) | `aurora-desktop-darwin-arm64.zip` | `aurora-cli-darwin-arm64.tar.gz` |
| Linux | `aurora-desktop-linux-amd64.tar.gz` | `aurora-cli-linux-amd64.tar.gz` |
---

## CLI (Optional)

Prefer the command line? Aurora has you covered:

```bash
# Check your config
aurora config

# See your collections
aurora penumbra

# Preview backup
aurora backup --validate

# Faster with multiple threads
aurora backup --threads 4
```

---

## Troubleshooting

If something isn't working correctly, check the `aurora.log` file located next to the executable. It contains detailed information about what Aurora is doing and any errors that occur.

---

## License

MIT
