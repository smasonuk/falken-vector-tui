package app

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/smasonuk/falken-vector/pkg/falkenvector"
)

func TestFormatStatusNilLastIndexed(t *testing.T) {
	got := FormatStatus(StatusViewModel{
		Directory: "/repo",
		Documents: falkenvector.StatusDocumentCounts{
			Indexed: 1,
			Error:   2,
			Deleted: 3,
		},
		Chunks: falkenvector.StatusChunkCounts{Active: 4, Inactive: 5},
	})
	for _, want := range []string{"Directory: /repo", "Documents: indexed 1", "Chunks:    active 4", "Last indexed: never"} {
		if !strings.Contains(got, want) {
			t.Fatalf("FormatStatus = %q, want %q", got, want)
		}
	}
}

func TestFormatStatusPopulatedLastIndexed(t *testing.T) {
	at := time.Date(2026, 5, 25, 18, 12, 0, 0, time.Local)
	got := FormatStatus(StatusViewModel{Directory: "/repo", LastIndexedAt: &at})
	if !strings.Contains(got, "Last indexed: 2026-05-25 18:12") {
		t.Fatalf("FormatStatus = %q", got)
	}
}

func TestIsNoIndexError(t *testing.T) {
	err := errors.Join(errors.New("outer"), falkenvector.ErrNoIndex)
	if !IsNoIndexError(err) {
		t.Fatalf("IsNoIndexError(%v) = false, want true", err)
	}
}
