package app

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/smasonuk/falken-vector/pkg/falkenvector"
)

const defaultEmbeddingConcurrency = 4

func ParseExtensions(input string) []string {
	return ParseCSVList(input)
}

func ParseCSVList(input string) []string {
	parts := strings.Split(input, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func ValidateChunkSettings(sizeInput, overlapInput string) (int, int, error) {
	size, err := strconv.Atoi(strings.TrimSpace(sizeInput))
	if err != nil || size <= 0 {
		return 0, 0, errors.New("chunk size must be an integer > 0")
	}
	overlap, err := strconv.Atoi(strings.TrimSpace(overlapInput))
	if err != nil || overlap < 0 {
		return 0, 0, errors.New("chunk overlap must be an integer >= 0")
	}
	if overlap >= size {
		return 0, 0, errors.New("chunk overlap must be less than chunk size")
	}
	return size, overlap, nil
}

func ValidateChunker(mode falkenvector.ChunkerMode) error {
	switch mode {
	case falkenvector.ChunkerAuto, falkenvector.ChunkerFixed, falkenvector.ChunkerMarkdown, falkenvector.ChunkerText, falkenvector.ChunkerCode:
		return nil
	default:
		return fmt.Errorf("chunker must be auto, fixed, markdown, text, or code")
	}
}

func BuildIngestRequest(model IndexViewModel, dryRun bool) (falkenvector.IngestRequest, error) {
	size, overlap, err := ValidateChunkSettings(model.ChunkSizeInput, model.ChunkOverlapInput)
	if err != nil {
		return falkenvector.IngestRequest{}, err
	}
	if err := ValidateChunker(model.Chunker); err != nil {
		return falkenvector.IngestRequest{}, err
	}
	return falkenvector.IngestRequest{
		Root:                 ".",
		Extensions:           ParseExtensions(model.ExtensionsInput),
		ExcludeExtensions:    ParseExtensions(model.ExcludeExtensionsInput),
		ExcludeDirs:          ParseCSVList(model.ExcludeDirsInput),
		ChunkSize:            size,
		ChunkOverlap:         overlap,
		ChunkerMode:          model.Chunker,
		DryRun:               dryRun,
		SyncSource:           model.SyncSource,
		EmbeddingConcurrency: defaultEmbeddingConcurrency,
	}, nil
}

func FormatIngestResult(result falkenvector.IngestResult) string {
	if result.Directory == "" && result.Scanned == 0 {
		return ""
	}
	return fmt.Sprintf(
		"scanned: %d\nnew files: %d\nchanged files: %d\nunchanged files: %d\ndeleted files: %d\nfailed files: %d\nchunks embedded: %d",
		result.Scanned,
		result.NewFiles,
		result.ChangedFiles,
		result.UnchangedFiles,
		result.DeletedFiles,
		result.FailedFiles,
		result.ChunksEmbedded,
	)
}
