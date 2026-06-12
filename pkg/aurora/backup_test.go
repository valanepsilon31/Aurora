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

	t.Run("kept when one collection is not excluded", func(t *testing.T) {
		col1 := &repository.PenumbraCollection{Name: "FilteredCollection"}
		col2 := &repository.PenumbraCollection{Name: "OtherCollection"}
		mod := &repository.PenumbraMod{
			Name:        "MyMod",
			Path:        "some/path",
			Collections: []*repository.PenumbraCollection{col1, col2},
		}
		filters := []string{"Filtered"}

		filtered, _ := isModFiltered(mod, filters)

		// The mod is still used by OtherCollection: keep it
		if filtered {
			t.Error("expected not filtered when a non-excluded collection still uses the mod")
		}
	})

	t.Run("excluded when every collection matches a filter", func(t *testing.T) {
		col1 := &repository.PenumbraCollection{Name: "xx-layle"}
		col2 := &repository.PenumbraCollection{Name: "xx-marielle"}
		mod := &repository.PenumbraMod{
			Name:        "MyMod",
			Path:        "some/path",
			Collections: []*repository.PenumbraCollection{col1, col2},
		}
		filters := []string{"xx-layle", "xx-marielle"}

		filtered, matchedFilter := isModFiltered(mod, filters)

		if !filtered {
			t.Error("expected filtered when every referencing collection is excluded")
		}
		if matchedFilter != "xx-layle" {
			t.Errorf("expected first collection's match 'xx-layle', got %s", matchedFilter)
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

		opts := NewBackupOptions(folders, threads, CompressionMax, "", quiet)

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
		opts := NewBackupOptions([]string{}, 1, CompressionMax, "", false)

		if opts.Quiet {
			t.Error("expected Quiet false")
		}
	})

	t.Run("compression presets map to levels", func(t *testing.T) {
		if got := NewBackupOptions(nil, 1, CompressionNormal, "", true).Level; got != 5 {
			t.Errorf("expected normal preset Level 5, got %d", got)
		}
		if got := NewBackupOptions(nil, 1, CompressionMax, "", true).Level; got != 9 {
			t.Errorf("expected max preset Level 9, got %d", got)
		}
		if got := NewBackupOptions(nil, 1, "", "", true).Level; got != 5 {
			t.Errorf("expected empty preset to default to Level 5 (normal), got %d", got)
		}
		if got := NewBackupOptions(nil, 1, "garbage", "", true).Level; got != 5 {
			t.Errorf("expected unknown preset to default to Level 5 (normal), got %d", got)
		}
	})
}

func TestInBackupSet(t *testing.T) {
	col := &repository.PenumbraCollection{Name: "SomeCollection"}

	t.Run("collection mod is selected", func(t *testing.T) {
		mod := &repository.PenumbraMod{Name: "Mod", Collections: []*repository.PenumbraCollection{col}}
		selected, excludedBy, includedBy := inBackupSet(mod, nil, nil)
		if !selected || excludedBy != "" || includedBy != "" {
			t.Errorf("expected plain selection, got selected=%v excludedBy=%q includedBy=%q", selected, excludedBy, includedBy)
		}
	})

	t.Run("unreferenced mod is dropped by default", func(t *testing.T) {
		mod := &repository.PenumbraMod{Name: "Orphan"}
		selected, _, _ := inBackupSet(mod, nil, nil)
		if selected {
			t.Error("expected unreferenced mod not selected without inclusion")
		}
	})

	t.Run("inclusion selects unreferenced mod", func(t *testing.T) {
		mod := &repository.PenumbraMod{Name: "Orphan"}
		selected, _, includedBy := inBackupSet(mod, nil, []string{"Orph"})
		if !selected || includedBy != "Orph" {
			t.Errorf("expected inclusion to select, got selected=%v includedBy=%q", selected, includedBy)
		}
	})

	t.Run("inclusion matches by path prefix", func(t *testing.T) {
		mod := &repository.PenumbraMod{Name: "Other", Path: "special/path"}
		selected, _, includedBy := inBackupSet(mod, nil, []string{"special"})
		if !selected || includedBy != "special" {
			t.Errorf("expected path inclusion, got selected=%v includedBy=%q", selected, includedBy)
		}
	})

	t.Run("inclusion wins over exclusion", func(t *testing.T) {
		mod := &repository.PenumbraMod{Name: "Orphan"}
		selected, excludedBy, includedBy := inBackupSet(mod, []string{"Orphan"}, []string{"Orphan"})
		if !selected {
			t.Error("expected inclusion to win over exclusion")
		}
		if excludedBy != "" || includedBy != "Orphan" {
			t.Errorf("expected only the decisive inclusion reported, got excludedBy=%q includedBy=%q", excludedBy, includedBy)
		}
	})

	t.Run("inclusion rescues excluded collection mod", func(t *testing.T) {
		mod := &repository.PenumbraMod{Name: "Mod", Collections: []*repository.PenumbraCollection{col}}
		selected, excludedBy, includedBy := inBackupSet(mod, []string{"Mod"}, []string{"Mod"})
		if !selected {
			t.Error("expected inclusion to rescue excluded collection mod")
		}
		if excludedBy != "" || includedBy != "Mod" {
			t.Errorf("expected inclusion reported as decisive, got excludedBy=%q includedBy=%q", excludedBy, includedBy)
		}
	})

	t.Run("inclusion is ignored for collection mods", func(t *testing.T) {
		mod := &repository.PenumbraMod{Name: "Mod", Collections: []*repository.PenumbraCollection{col}}
		_, _, includedBy := inBackupSet(mod, nil, []string{"Mod"})
		if includedBy != "" {
			t.Errorf("expected no inclusion mark on collection mod, got %q", includedBy)
		}
	})

	t.Run("exclusion drops collection mod", func(t *testing.T) {
		mod := &repository.PenumbraMod{Name: "Mod", Collections: []*repository.PenumbraCollection{col}}
		selected, excludedBy, _ := inBackupSet(mod, []string{"Mod"}, nil)
		if selected || excludedBy != "Mod" {
			t.Errorf("expected exclusion, got selected=%v excludedBy=%q", selected, excludedBy)
		}
	})
}

func TestFilterCaseInsensitive(t *testing.T) {
	t.Run("exclusion matches regardless of case", func(t *testing.T) {
		mod := &repository.PenumbraMod{Name: "ShadowKnight", Path: "some/path"}
		filtered, matched := isModFiltered(mod, []string{"shadow"})
		if !filtered || matched != "shadow" {
			t.Errorf("expected case-insensitive exclusion match, got filtered=%v matched=%q", filtered, matched)
		}
	})

	t.Run("exclusion matches uppercase filter on lowercase mod", func(t *testing.T) {
		mod := &repository.PenumbraMod{Name: "shadowknight", Path: "some/path"}
		filtered, _ := isModFiltered(mod, []string{"SHADOW"})
		if !filtered {
			t.Error("expected uppercase filter to match lowercase mod name")
		}
	})

	t.Run("inclusion matches regardless of case", func(t *testing.T) {
		mod := &repository.PenumbraMod{Name: "OrphanMod"}
		included, matched := isModIncluded(mod, []string{"orphan"})
		if !included || matched != "orphan" {
			t.Errorf("expected case-insensitive inclusion match, got included=%v matched=%q", included, matched)
		}
	})

	t.Run("no match on different prefix", func(t *testing.T) {
		mod := &repository.PenumbraMod{Name: "ShadowKnight", Path: "some/path"}
		if filtered, _ := isModFiltered(mod, []string{"light"}); filtered {
			t.Error("expected no match for unrelated filter")
		}
	})

	t.Run("prefix longer than name does not match", func(t *testing.T) {
		mod := &repository.PenumbraMod{Name: "Mod", Path: "p"}
		if filtered, _ := isModFiltered(mod, []string{"ModWithLongerName"}); filtered {
			t.Error("expected no match when filter is longer than name")
		}
	})
}

func TestNewBackupOptionsOutputDir(t *testing.T) {
	t.Run("empty output dir keeps relative path", func(t *testing.T) {
		opts := NewBackupOptions(nil, 1, CompressionNormal, "", true)
		if opts.OutputPath != BackupOutputPath {
			t.Errorf("expected %q, got %q", BackupOutputPath, opts.OutputPath)
		}
	})

	t.Run("output dir is joined into the path", func(t *testing.T) {
		opts := NewBackupOptions(nil, 1, CompressionNormal, "/tmp/backups", true)
		want := "/tmp/backups/" + BackupOutputPath
		if opts.OutputPath != want {
			t.Errorf("expected %q, got %q", want, opts.OutputPath)
		}
	})
}
