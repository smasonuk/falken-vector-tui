package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tuiconfig "github.com/smasonuk/falken-vector-tui/internal/config"
	"github.com/smasonuk/falken-vector/pkg/falkenvector"
)

const preferencesVersion = 1

type Preferences struct {
	Version int                `json:"version"`
	Index   IndexPreferences   `json:"index,omitempty"`
	Compact CompactPreferences `json:"compact,omitempty"`
}

type IndexPreferences struct {
	ExtensionsInput        *string                   `json:"extensions,omitempty"`
	ExcludeExtensionsInput *string                   `json:"exclude_extensions,omitempty"`
	ExcludeDirsInput       *string                   `json:"exclude_dirs,omitempty"`
	Chunker                *falkenvector.ChunkerMode `json:"chunker,omitempty"`
	ChunkSizeInput         *string                   `json:"chunk_size,omitempty"`
	ChunkOverlapInput      *string                   `json:"chunk_overlap,omitempty"`
	SyncSource             *bool                     `json:"sync_source,omitempty"`
}

type CompactPreferences struct {
	KeepBackup     *bool   `json:"keep_backup,omitempty"`
	BatchSizeInput *string `json:"batch_size,omitempty"`
}

func ResolvePreferencesPath(cfg tuiconfig.Config, getenv func(string) string, cwd string) string {
	if getenv == nil {
		getenv = os.Getenv
	}
	stateDir := strings.TrimSpace(cfg.StateDir)
	if stateDir == "" {
		stateDir = strings.TrimSpace(getenv("FALKENGO_STATE_DIR"))
	}
	if stateDir == "" {
		stateDir = filepath.Join(cwd, ".falkengo")
	} else if !filepath.IsAbs(stateDir) {
		stateDir = filepath.Join(cwd, stateDir)
	}
	return filepath.Join(filepath.Clean(stateDir), "tui", "preferences.json")
}

func LoadPreferences(path string) (Preferences, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Preferences{Version: preferencesVersion}, nil
	}
	if err != nil {
		return Preferences{Version: preferencesVersion}, fmt.Errorf("load preferences: %w", err)
	}
	var prefs Preferences
	if err := json.Unmarshal(data, &prefs); err != nil {
		return Preferences{Version: preferencesVersion}, fmt.Errorf("load preferences: %w", err)
	}
	if prefs.Version == 0 {
		prefs.Version = preferencesVersion
	}
	return prefs, nil
}

func SavePreferences(path string, prefs Preferences) error {
	prefs.Version = preferencesVersion
	data, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		return fmt.Errorf("save preferences: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("save preferences: %w", err)
	}
	file, err := os.CreateTemp(dir, ".preferences-*.tmp")
	if err != nil {
		return fmt.Errorf("save preferences: %w", err)
	}
	tempPath := file.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tempPath)
		}
	}()
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return fmt.Errorf("save preferences: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("save preferences: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("save preferences: %w", err)
	}
	cleanup = false
	return nil
}

func PreferencesFromModel(model AppModel) Preferences {
	return Preferences{
		Version: preferencesVersion,
		Index: IndexPreferences{
			ExtensionsInput:        stringPtr(model.Index.ExtensionsInput),
			ExcludeExtensionsInput: stringPtr(model.Index.ExcludeExtensionsInput),
			ExcludeDirsInput:       stringPtr(model.Index.ExcludeDirsInput),
			Chunker:                chunkerPtr(model.Index.Chunker),
			ChunkSizeInput:         stringPtr(model.Index.ChunkSizeInput),
			ChunkOverlapInput:      stringPtr(model.Index.ChunkOverlapInput),
			SyncSource:             boolPtr(model.Index.SyncSource),
		},
		Compact: CompactPreferences{
			KeepBackup:     boolPtr(model.Compact.KeepBackup),
			BatchSizeInput: stringPtr(model.Compact.BatchSizeInput),
		},
	}
}

func ApplyPreferences(model AppModel, prefs Preferences) AppModel {
	if prefs.Index.ExtensionsInput != nil {
		model.Index.ExtensionsInput = *prefs.Index.ExtensionsInput
	}
	if prefs.Index.ExcludeExtensionsInput != nil {
		model.Index.ExcludeExtensionsInput = *prefs.Index.ExcludeExtensionsInput
	}
	if prefs.Index.ExcludeDirsInput != nil {
		model.Index.ExcludeDirsInput = *prefs.Index.ExcludeDirsInput
	}
	if prefs.Index.Chunker != nil && ValidateChunker(*prefs.Index.Chunker) == nil {
		model.Index.Chunker = *prefs.Index.Chunker
	}
	if prefs.Index.ChunkSizeInput != nil {
		model.Index.ChunkSizeInput = *prefs.Index.ChunkSizeInput
	}
	if prefs.Index.ChunkOverlapInput != nil {
		model.Index.ChunkOverlapInput = *prefs.Index.ChunkOverlapInput
	}
	if prefs.Index.SyncSource != nil {
		model.Index.SyncSource = *prefs.Index.SyncSource
	}
	if prefs.Compact.KeepBackup != nil {
		model.Compact.KeepBackup = *prefs.Compact.KeepBackup
	}
	if prefs.Compact.BatchSizeInput != nil {
		model.Compact.BatchSizeInput = *prefs.Compact.BatchSizeInput
	}
	return model
}

func stringPtr(value string) *string {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}

func chunkerPtr(value falkenvector.ChunkerMode) *falkenvector.ChunkerMode {
	return &value
}
