package app

import (
	"os"
	"path/filepath"
	"testing"

	tuiconfig "github.com/smasonuk/falken-vector-tui/internal/config"
	"github.com/smasonuk/falken-vector/pkg/falkenvector"
)

func TestResolvePreferencesPath(t *testing.T) {
	cwd := t.TempDir()
	getenv := func(key string) string {
		if key == "FALKENGO_STATE_DIR" {
			return "env-state"
		}
		return ""
	}

	got := ResolvePreferencesPath(tuiconfig.Config{}, func(string) string { return "" }, cwd)
	want := filepath.Join(cwd, ".falkengo", "tui", "preferences.json")
	if got != want {
		t.Fatalf("default path = %q, want %q", got, want)
	}

	got = ResolvePreferencesPath(tuiconfig.Config{}, getenv, cwd)
	want = filepath.Join(cwd, "env-state", "tui", "preferences.json")
	if got != want {
		t.Fatalf("env path = %q, want %q", got, want)
	}

	got = ResolvePreferencesPath(tuiconfig.Config{StateDir: "flag-state"}, getenv, cwd)
	want = filepath.Join(cwd, "flag-state", "tui", "preferences.json")
	if got != want {
		t.Fatalf("flag path = %q, want %q", got, want)
	}
}

func TestLoadPreferencesMissingFileUsesDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".falkengo", "tui", "preferences.json")
	prefs, err := LoadPreferences(path)
	if err != nil {
		t.Fatalf("LoadPreferences: %v", err)
	}
	if prefs.Version != preferencesVersion {
		t.Fatalf("Version = %d, want %d", prefs.Version, preferencesVersion)
	}
}

func TestSaveLoadPreferencesRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".falkengo", "tui", "preferences.json")
	model := NewModel("/repo")
	model.Index.ExtensionsInput = "go,md"
	model.Index.ExcludeExtensionsInput = "tmp,log"
	model.Index.ExcludeDirsInput = "fixtures,dist"
	model.Index.Chunker = falkenvector.ChunkerMarkdown
	model.Index.ChunkSizeInput = "900"
	model.Index.ChunkOverlapInput = "90"
	model.Index.SyncSource = true
	model.Compact.KeepBackup = true
	model.Compact.BatchSizeInput = "50"

	if err := SavePreferences(path, PreferencesFromModel(model)); err != nil {
		t.Fatalf("SavePreferences: %v", err)
	}
	loaded, err := LoadPreferences(path)
	if err != nil {
		t.Fatalf("LoadPreferences: %v", err)
	}
	applied := ApplyPreferences(NewModel("/repo"), loaded)
	if applied.Index.ExcludeExtensionsInput != "tmp,log" {
		t.Fatalf("ExcludeExtensionsInput = %q", applied.Index.ExcludeExtensionsInput)
	}
	if applied.Index.ExcludeDirsInput != "fixtures,dist" {
		t.Fatalf("ExcludeDirsInput = %q", applied.Index.ExcludeDirsInput)
	}
	if applied.Index.Chunker != falkenvector.ChunkerMarkdown || !applied.Index.SyncSource {
		t.Fatalf("Index preferences not restored: %+v", applied.Index)
	}
	if applied.Search.Mode != falkenvector.RetrievalHybrid || applied.Search.TopKInput != "8" {
		t.Fatalf("Search defaults changed: %+v", applied.Search)
	}
	if applied.Ask.Mode != falkenvector.RetrievalHybrid || applied.Ask.TopKInput != "8" || !applied.Ask.Agent {
		t.Fatalf("Ask preferences not restored: %+v", applied.Ask)
	}
	if !applied.Compact.KeepBackup || applied.Compact.BatchSizeInput != "50" {
		t.Fatalf("Compact preferences not restored: %+v", applied.Compact)
	}
}

func TestLoadPreferencesCorruptFileIsNonFatal(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".falkengo", "tui", "preferences.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	prefs, err := LoadPreferences(path)
	if err == nil {
		t.Fatal("LoadPreferences err = nil, want corrupt JSON error")
	}
	if prefs.Version != preferencesVersion {
		t.Fatalf("Version = %d, want %d", prefs.Version, preferencesVersion)
	}
}

func TestApplyPreferencesDoesNotOverwriteVolatileState(t *testing.T) {
	model := NewModel("/repo")
	model.Search.Question = "draft search"
	model.Search.Results = []SearchResultView{{Title: "result"}}
	model.Search.Selected = 1
	model.Ask.Question = "draft ask"
	model.Ask.Answer = "answer"
	model.Ask.Sources = []string{"source"}
	model.Compact.Confirming = true
	model.Compact.DryRunComplete = true
	model.Events = []EventLine{{Text: "event"}}
	model.Busy = true
	model.LastError = "existing error"

	source := NewModel("/repo")
	source.Index.ExcludeExtensionsInput = "tmp"
	source.Compact.KeepBackup = true

	applied := ApplyPreferences(model, PreferencesFromModel(source))
	if applied.Search.Question != "draft search" || len(applied.Search.Results) != 1 || applied.Search.Selected != 1 {
		t.Fatalf("search volatile state changed: %+v", applied.Search)
	}
	if applied.Ask.Question != "draft ask" || applied.Ask.Answer != "answer" || len(applied.Ask.Sources) != 1 {
		t.Fatalf("ask volatile state changed: %+v", applied.Ask)
	}
	if !applied.Compact.Confirming || !applied.Compact.DryRunComplete {
		t.Fatalf("compact volatile state changed: %+v", applied.Compact)
	}
	if len(applied.Events) != 1 || !applied.Busy || applied.LastError != "existing error" {
		t.Fatalf("runtime state changed: %+v", applied)
	}
	if applied.Index.ExcludeExtensionsInput != "tmp" || applied.Search.Mode != falkenvector.RetrievalHybrid || !applied.Ask.Agent || !applied.Compact.KeepBackup {
		t.Fatalf("stable preferences were not applied: %+v", applied)
	}
}
