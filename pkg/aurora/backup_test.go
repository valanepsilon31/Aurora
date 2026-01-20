package aurora

import (
	"aurora/internal/repository"
	"testing"
)

func TestIsModFiltered(t *testing.T) {
	t.Run("no filters", func(t *testing.T) {
		mod := &repository.PenumbraMod{Name: "TestMod", Path: "some/path"}
		filters := []string{}

		filtered, matchedFilter := isModFiltered(mod, filters)

		if filtered {
			t.Error("expected not filtered when no filters")
		}
		if matchedFilter != "" {
			t.Errorf("expected empty matchedFilter, got %s", matchedFilter)
		}
	})

	t.Run("filter by name prefix", func(t *testing.T) {
		mod := &repository.PenumbraMod{Name: "TestMod", Path: "some/path"}
		filters := []string{"Test"}

		filtered, matchedFilter := isModFiltered(mod, filters)

		if !filtered {
			t.Error("expected filtered when name matches prefix")
		}
		if matchedFilter != "Test" {
			t.Errorf("expected matchedFilter 'Test', got %s", matchedFilter)
		}
	})

	t.Run("filter by path prefix", func(t *testing.T) {
		mod := &repository.PenumbraMod{Name: "MyMod", Path: "filtered/path"}
		filters := []string{"filtered"}

		filtered, matchedFilter := isModFiltered(mod, filters)

		if !filtered {
			t.Error("expected filtered when path matches prefix")
		}
		if matchedFilter != "filtered" {
			t.Errorf("expected matchedFilter 'filtered', got %s", matchedFilter)
		}
	})

	t.Run("filter by single collection name", func(t *testing.T) {
		col := &repository.PenumbraCollection{Name: "FilteredCollection"}
		mod := &repository.PenumbraMod{
			Name:        "MyMod",
			Path:        "some/path",
			Collections: []*repository.PenumbraCollection{col},
		}
		filters := []string{"Filtered"}

		filtered, matchedFilter := isModFiltered(mod, filters)

		if !filtered {
			t.Error("expected filtered when single collection name matches prefix")
		}
		if matchedFilter != "Filtered" {
			t.Errorf("expected matchedFilter 'Filtered', got %s", matchedFilter)
		}
	})

	t.Run("multiple collections - no filter", func(t *testing.T) {
		col1 := &repository.PenumbraCollection{Name: "FilteredCollection"}
		col2 := &repository.PenumbraCollection{Name: "OtherCollection"}
		mod := &repository.PenumbraMod{
			Name:        "MyMod",
			Path:        "some/path",
			Collections: []*repository.PenumbraCollection{col1, col2},
		}
		filters := []string{"Filtered"}

		filtered, _ := isModFiltered(mod, filters)

		// Should NOT filter when mod is in multiple collections
		if filtered {
			t.Error("expected not filtered when mod is in multiple collections")
		}
	})

	t.Run("no match", func(t *testing.T) {
		mod := &repository.PenumbraMod{Name: "MyMod", Path: "my/path"}
		filters := []string{"Other", "Different"}

		filtered, matchedFilter := isModFiltered(mod, filters)

		if filtered {
			t.Error("expected not filtered when no filter matches")
		}
		if matchedFilter != "" {
			t.Errorf("expected empty matchedFilter, got %s", matchedFilter)
		}
	})

	t.Run("first matching filter wins", func(t *testing.T) {
		mod := &repository.PenumbraMod{Name: "TestMod", Path: "TestPath"}
		filters := []string{"Test", "TestMod"}

		filtered, matchedFilter := isModFiltered(mod, filters)

		if !filtered {
			t.Error("expected filtered")
		}
		if matchedFilter != "Test" {
			t.Errorf("expected first matching filter 'Test', got %s", matchedFilter)
		}
	})
}

func TestNewBackupOptions(t *testing.T) {
	t.Run("creates options with correct values", func(t *testing.T) {
		folders := []string{"/path/to/mod1", "/path/to/mod2"}
		threads := 4
		quiet := true

		opts := NewBackupOptions(folders, threads, quiet)

		if opts.OutputPath != BackupOutputPath {
			t.Errorf("expected OutputPath %s, got %s", BackupOutputPath, opts.OutputPath)
		}
		if len(opts.Files) != 2 {
			t.Errorf("expected 2 files, got %d", len(opts.Files))
		}
		if opts.MaxThreads != 4 {
			t.Errorf("expected MaxThreads 4, got %d", opts.MaxThreads)
		}
		if opts.Level != 9 {
			t.Errorf("expected Level 9, got %d", opts.Level)
		}
		if !opts.UseZipFormat {
			t.Error("expected UseZipFormat true")
		}
		if !opts.Quiet {
			t.Error("expected Quiet true")
		}
	})

	t.Run("quiet false", func(t *testing.T) {
		opts := NewBackupOptions([]string{}, 1, false)

		if opts.Quiet {
			t.Error("expected Quiet false")
		}
	})
}
