package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHasMatchingMod(t *testing.T) {
	t.Run("matching mod exists", func(t *testing.T) {
		dir := t.TempDir()

		// Create sort_order.json
		sortOrderPath := filepath.Join(dir, "sort_order.json")
		os.WriteFile(sortOrderPath, []byte(`{"Data": {"path1": "ModA", "path2": "ModB"}}`), 0644)

		// Create mods directory with one matching mod
		modsPath := filepath.Join(dir, "mods")
		os.MkdirAll(filepath.Join(modsPath, "ModA"), 0755)

		result := hasMatchingMod(sortOrderPath, modsPath)

		if !result {
			t.Error("expected true when matching mod exists")
		}
	})

	t.Run("no matching mods", func(t *testing.T) {
		dir := t.TempDir()

		// Create sort_order.json
		sortOrderPath := filepath.Join(dir, "sort_order.json")
		os.WriteFile(sortOrderPath, []byte(`{"Data": {"path1": "ModA", "path2": "ModB"}}`), 0644)

		// Create empty mods directory
		modsPath := filepath.Join(dir, "mods")
		os.MkdirAll(modsPath, 0755)

		result := hasMatchingMod(sortOrderPath, modsPath)

		if result {
			t.Error("expected false when no matching mods exist")
		}
	})

	t.Run("sort_order.json not found", func(t *testing.T) {
		dir := t.TempDir()
		result := hasMatchingMod(filepath.Join(dir, "nonexistent.json"), dir)

		if result {
			t.Error("expected false when sort_order.json doesn't exist")
		}
	})

	t.Run("mod path with subdirectory", func(t *testing.T) {
		dir := t.TempDir()

		// Create sort_order.json with nested path (like "folder/ModName")
		sortOrderPath := filepath.Join(dir, "sort_order.json")
		os.WriteFile(sortOrderPath, []byte(`{"Data": {"path1": "subfolder/ModA"}}`), 0644)

		// Create mods directory - filepath.Base should extract "ModA"
		modsPath := filepath.Join(dir, "mods")
		os.MkdirAll(filepath.Join(modsPath, "ModA"), 0755)

		result := hasMatchingMod(sortOrderPath, modsPath)

		if !result {
			t.Error("expected true when mod exists (extracted via filepath.Base)")
		}
	})

	t.Run("mod is file not directory", func(t *testing.T) {
		dir := t.TempDir()

		sortOrderPath := filepath.Join(dir, "sort_order.json")
		os.WriteFile(sortOrderPath, []byte(`{"Data": {"path1": "ModA"}}`), 0644)

		modsPath := filepath.Join(dir, "mods")
		os.MkdirAll(modsPath, 0755)
		// Create a file instead of directory
		os.WriteFile(filepath.Join(modsPath, "ModA"), []byte("not a dir"), 0644)

		result := hasMatchingMod(sortOrderPath, modsPath)

		if result {
			t.Error("expected false when mod path is a file, not directory")
		}
	})
}

func TestConfigStatus(t *testing.T) {
	t.Run("invalid penumbra path", func(t *testing.T) {
		cfg := &Config{
			Penumbra: PenumbraConfig{Path: "/nonexistent/path"},
			Mods:     ModsConfig{Path: "/nonexistent/mods"},
		}

		status := cfg.Status()

		if status.Valid {
			t.Error("expected invalid status")
		}
		if status.Penumbra != "Invalid Penumbra path" {
			t.Errorf("unexpected penumbra status: %s", status.Penumbra)
		}
	})

	t.Run("missing sort_order.json", func(t *testing.T) {
		dir := t.TempDir()
		penumbraPath := filepath.Join(dir, "penumbra")
		os.MkdirAll(penumbraPath, 0755)

		cfg := &Config{
			Penumbra: PenumbraConfig{Path: penumbraPath},
			Mods:     ModsConfig{Path: dir},
		}

		status := cfg.Status()

		if status.Valid {
			t.Error("expected invalid status")
		}
		if status.Penumbra != "sort_order.json not found in Penumbra path" {
			t.Errorf("unexpected penumbra status: %s", status.Penumbra)
		}
	})

	t.Run("missing collections folder", func(t *testing.T) {
		dir := t.TempDir()
		penumbraPath := filepath.Join(dir, "penumbra")
		os.MkdirAll(penumbraPath, 0755)
		os.WriteFile(filepath.Join(penumbraPath, "sort_order.json"), []byte(`{"Data":{}}`), 0644)

		cfg := &Config{
			Penumbra: PenumbraConfig{Path: penumbraPath},
			Mods:     ModsConfig{Path: dir},
		}

		status := cfg.Status()

		if status.Valid {
			t.Error("expected invalid status")
		}
		if status.Penumbra != "collections folder not found in Penumbra path" {
			t.Errorf("unexpected penumbra status: %s", status.Penumbra)
		}
	})

	t.Run("valid config", func(t *testing.T) {
		dir := t.TempDir()

		// Setup penumbra path
		penumbraPath := filepath.Join(dir, "penumbra")
		os.MkdirAll(penumbraPath, 0755)
		os.WriteFile(filepath.Join(penumbraPath, "sort_order.json"), []byte(`{"Data":{"p1":"ModA"}}`), 0644)
		os.MkdirAll(filepath.Join(penumbraPath, "collections"), 0755)

		// Setup mods path with matching mod
		modsPath := filepath.Join(dir, "mods")
		os.MkdirAll(filepath.Join(modsPath, "ModA"), 0755)

		cfg := &Config{
			Penumbra: PenumbraConfig{Path: penumbraPath},
			Mods:     ModsConfig{Path: modsPath},
		}

		status := cfg.Status()

		if !status.Valid {
			t.Errorf("expected valid status, got penumbra=%s, mods=%s", status.Penumbra, status.Mods)
		}
	})

	t.Run("no matching mods", func(t *testing.T) {
		dir := t.TempDir()

		// Setup penumbra path
		penumbraPath := filepath.Join(dir, "penumbra")
		os.MkdirAll(penumbraPath, 0755)
		os.WriteFile(filepath.Join(penumbraPath, "sort_order.json"), []byte(`{"Data":{"p1":"ModA"}}`), 0644)
		os.MkdirAll(filepath.Join(penumbraPath, "collections"), 0755)

		// Setup empty mods path
		modsPath := filepath.Join(dir, "mods")
		os.MkdirAll(modsPath, 0755)

		cfg := &Config{
			Penumbra: PenumbraConfig{Path: penumbraPath},
			Mods:     ModsConfig{Path: modsPath},
		}

		status := cfg.Status()

		if status.Valid {
			t.Error("expected invalid status when no mods match")
		}
		if status.Mods != "No mods from Penumbra found in Mods folder" {
			t.Errorf("unexpected mods status: %s", status.Mods)
		}
	})
}
