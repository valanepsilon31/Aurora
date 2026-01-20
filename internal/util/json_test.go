package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadJSONFile(t *testing.T) {
	t.Run("valid JSON", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.json")
		os.WriteFile(path, []byte(`{"name": "test", "value": 42}`), 0644)

		var result struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}
		err := ReadJSONFile(path, &result)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Name != "test" {
			t.Errorf("expected name 'test', got '%s'", result.Name)
		}
		if result.Value != 42 {
			t.Errorf("expected value 42, got %d", result.Value)
		}
	})

	t.Run("JSON with BOM", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "bom.json")
		// UTF-8 BOM + JSON content
		content := append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"name": "bom-test"}`)...)
		os.WriteFile(path, content, 0644)

		var result struct {
			Name string `json:"name"`
		}
		err := ReadJSONFile(path, &result)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Name != "bom-test" {
			t.Errorf("expected name 'bom-test', got '%s'", result.Name)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		var result struct{}
		err := ReadJSONFile("/nonexistent/path.json", &result)

		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "invalid.json")
		os.WriteFile(path, []byte(`{invalid json}`), 0644)

		var result struct{}
		err := ReadJSONFile(path, &result)

		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}
