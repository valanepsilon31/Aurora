package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHasModFolders(t *testing.T) {
	t.Run("has subdirectories", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, "ModA"), 0755)

		if !hasModFolders(dir) {
			t.Error("expected true when subdirectories exist")
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		dir := t.TempDir()

		if hasModFolders(dir) {
			t.Error("expected false when directory is empty")
		}
	})

	t.Run("only files no directories", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "not_a_mod.txt"), []byte("file"), 0644)

		if hasModFolders(dir) {
			t.Error("expected false when only files exist")
		}
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		if hasModFolders("/nonexistent/path") {
			t.Error("expected false for nonexistent directory")
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

	t.Run("missing collections folder", func(t *testing.T) {
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
		if status.Penumbra != "collections folder not found in Penumbra path" {
			t.Errorf("unexpected penumbra status: %s", status.Penumbra)
		}
	})

	t.Run("valid config", func(t *testing.T) {
		dir := t.TempDir()

		penumbraPath := filepath.Join(dir, "penumbra")
		os.MkdirAll(penumbraPath, 0755)
		os.MkdirAll(filepath.Join(penumbraPath, "collections"), 0755)

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

	t.Run("no mod folders", func(t *testing.T) {
		dir := t.TempDir()

		penumbraPath := filepath.Join(dir, "penumbra")
		os.MkdirAll(penumbraPath, 0755)
		os.MkdirAll(filepath.Join(penumbraPath, "collections"), 0755)

		modsPath := filepath.Join(dir, "mods")
		os.MkdirAll(modsPath, 0755)

		cfg := &Config{
			Penumbra: PenumbraConfig{Path: penumbraPath},
			Mods:     ModsConfig{Path: modsPath},
		}

		status := cfg.Status()

		if status.Valid {
			t.Error("expected invalid status when no mod folders exist")
		}
		if status.Mods != "No mod folders found in Mods path" {
			t.Errorf("unexpected mods status: %s", status.Mods)
		}
	})
}
