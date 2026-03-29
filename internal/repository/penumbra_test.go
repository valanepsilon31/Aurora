package repository

import (
	"aurora/internal/config"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMods(t *testing.T) {
	t.Run("loads mod directories sorted by name", func(t *testing.T) {
		modsDir := t.TempDir()
		penumbraDir := t.TempDir()
		os.MkdirAll(filepath.Join(penumbraDir, "collections"), 0755)

		// Create mod directories with files
		for _, name := range []string{"ModC", "ModA", "ModB"} {
			modPath := filepath.Join(modsDir, name)
			os.MkdirAll(modPath, 0755)
			os.WriteFile(filepath.Join(modPath, "data.txt"), []byte("content"), 0644)
		}

		cfg := &config.Config{
			Penumbra: config.PenumbraConfig{Path: penumbraDir},
			Mods:     config.ModsConfig{Path: modsDir},
		}

		repo := NewPenumbraRepository(cfg)

		if len(repo.Mods) != 3 {
			t.Fatalf("expected 3 mods, got %d", len(repo.Mods))
		}
		if repo.Mods[0].Name != "ModA" {
			t.Errorf("expected first mod to be ModA, got %s", repo.Mods[0].Name)
		}
		if repo.Mods[1].Name != "ModB" {
			t.Errorf("expected second mod to be ModB, got %s", repo.Mods[1].Name)
		}
		if repo.Mods[2].Name != "ModC" {
			t.Errorf("expected third mod to be ModC, got %s", repo.Mods[2].Name)
		}
	})

	t.Run("skips empty mod directories", func(t *testing.T) {
		modsDir := t.TempDir()
		penumbraDir := t.TempDir()
		os.MkdirAll(filepath.Join(penumbraDir, "collections"), 0755)

		// Create one mod with content and one empty
		os.MkdirAll(filepath.Join(modsDir, "HasContent"), 0755)
		os.WriteFile(filepath.Join(modsDir, "HasContent", "data.txt"), []byte("content"), 0644)
		os.MkdirAll(filepath.Join(modsDir, "EmptyMod"), 0755)

		cfg := &config.Config{
			Penumbra: config.PenumbraConfig{Path: penumbraDir},
			Mods:     config.ModsConfig{Path: modsDir},
		}

		repo := NewPenumbraRepository(cfg)

		if len(repo.Mods) != 1 {
			t.Fatalf("expected 1 mod, got %d", len(repo.Mods))
		}
		if repo.Mods[0].Name != "HasContent" {
			t.Errorf("expected HasContent, got %s", repo.Mods[0].Name)
		}
	})

	t.Run("skips files in mods directory", func(t *testing.T) {
		modsDir := t.TempDir()
		penumbraDir := t.TempDir()
		os.MkdirAll(filepath.Join(penumbraDir, "collections"), 0755)

		// Create a file (not directory) in mods path
		os.WriteFile(filepath.Join(modsDir, "not_a_mod.txt"), []byte("file"), 0644)

		// Create one actual mod directory
		os.MkdirAll(filepath.Join(modsDir, "RealMod"), 0755)
		os.WriteFile(filepath.Join(modsDir, "RealMod", "data.txt"), []byte("content"), 0644)

		cfg := &config.Config{
			Penumbra: config.PenumbraConfig{Path: penumbraDir},
			Mods:     config.ModsConfig{Path: modsDir},
		}

		repo := NewPenumbraRepository(cfg)

		if len(repo.Mods) != 1 {
			t.Fatalf("expected 1 mod, got %d", len(repo.Mods))
		}
		if repo.Mods[0].Name != "RealMod" {
			t.Errorf("expected RealMod, got %s", repo.Mods[0].Name)
		}
	})

	t.Run("mod path equals mod name", func(t *testing.T) {
		modsDir := t.TempDir()
		penumbraDir := t.TempDir()
		os.MkdirAll(filepath.Join(penumbraDir, "collections"), 0755)

		os.MkdirAll(filepath.Join(modsDir, "TestMod"), 0755)
		os.WriteFile(filepath.Join(modsDir, "TestMod", "data.txt"), []byte("content"), 0644)

		cfg := &config.Config{
			Penumbra: config.PenumbraConfig{Path: penumbraDir},
			Mods:     config.ModsConfig{Path: modsDir},
		}

		repo := NewPenumbraRepository(cfg)

		if len(repo.Mods) != 1 {
			t.Fatalf("expected 1 mod, got %d", len(repo.Mods))
		}
		if repo.Mods[0].Path != "TestMod" {
			t.Errorf("expected Path to equal folder name 'TestMod', got %s", repo.Mods[0].Path)
		}
	})
}
