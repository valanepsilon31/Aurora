package repository

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSortOrder(t *testing.T) {
	t.Run("valid sort_order.json", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "sort_order.json")
		os.WriteFile(path, []byte(`{"Data": {"path/to/mod1": "ModA", "path/to/mod2": "ModB"}}`), 0644)

		result, err := LoadSortOrder(path)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Data) != 2 {
			t.Errorf("expected 2 entries, got %d", len(result.Data))
		}
		if result.Data["path/to/mod1"] != "ModA" {
			t.Errorf("expected ModA, got %s", result.Data["path/to/mod1"])
		}
	})

	t.Run("with BOM", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "sort_order.json")
		content := append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"Data": {"p1": "Mod1"}}`)...)
		os.WriteFile(path, content, 0644)

		result, err := LoadSortOrder(path)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Data["p1"] != "Mod1" {
			t.Errorf("expected Mod1, got %s", result.Data["p1"])
		}
	})

	t.Run("empty data", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "sort_order.json")
		os.WriteFile(path, []byte(`{"Data": {}}`), 0644)

		result, err := LoadSortOrder(path)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Data) != 0 {
			t.Errorf("expected empty data, got %d entries", len(result.Data))
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := LoadSortOrder("/nonexistent/path.json")

		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "sort_order.json")
		os.WriteFile(path, []byte(`{invalid}`), 0644)

		_, err := LoadSortOrder(path)

		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}
