package app

import (
	"fmt"
	"strings"

	"github.com/smasonuk/falken-vector/pkg/falkenvector"
)

func BuildCompactRequest(model CompactViewModel, dryRun bool) (falkenvector.CompactRequest, error) {
	batchSize, err := ParsePositiveInt(model.BatchSizeInput, "batch size")
	if err != nil {
		return falkenvector.CompactRequest{}, err
	}
	return falkenvector.CompactRequest{
		DryRun:               dryRun,
		BatchSize:            batchSize,
		KeepBackup:           model.KeepBackup,
		EmbeddingConcurrency: defaultEmbeddingConcurrency,
	}, nil
}

func FormatCompactResult(result falkenvector.CompactResult) string {
	if result.ActiveChunks == 0 && result.InactiveChunks == 0 && result.ReembeddedChunks == 0 && result.UpdatedChunks == 0 && result.PendingRuns == 0 {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "active chunks: %d\n", result.ActiveChunks)
	fmt.Fprintf(&b, "inactive chunks: %d\n", result.InactiveChunks)
	fmt.Fprintf(&b, "re-embedded chunks: %d\n", result.ReembeddedChunks)
	fmt.Fprintf(&b, "updated chunks: %d\n", result.UpdatedChunks)
	if result.EmbeddingModel != "" {
		fmt.Fprintf(&b, "embedding model: %s\n", result.EmbeddingModel)
	}
	if result.PendingRuns != 0 {
		fmt.Fprintf(&b, "pending runs: %d\n", result.PendingRuns)
	}
	if len(result.Warnings) != 0 {
		b.WriteString("warnings:\n")
		for _, warning := range result.Warnings {
			fmt.Fprintf(&b, "- %s\n", warning)
		}
	}
	return strings.TrimSpace(b.String())
}
