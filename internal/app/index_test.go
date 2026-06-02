package app

import (
	"reflect"
	"strings"
	"testing"

	"github.com/smasonuk/falken-vector/pkg/falkenvector"
)

func TestParseExtensions(t *testing.T) {
	got := ParseExtensions("go, md,txt")
	want := []string{"go", "md", "txt"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseExtensions = %+v, want %+v", got, want)
	}
	if got := ParseExtensions(""); got != nil {
		t.Fatalf("ParseExtensions empty = %+v, want nil", got)
	}
}

func TestParseCSVList(t *testing.T) {
	got := ParseCSVList("fixtures, dist,generated")
	want := []string{"fixtures", "dist", "generated"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseCSVList = %+v, want %+v", got, want)
	}
}

func TestBuildIngestRequest(t *testing.T) {
	request, err := BuildIngestRequest(IndexViewModel{
		ExtensionsInput:        "go,md",
		ExcludeExtensionsInput: "tmp, log",
		ExcludeDirsInput:       "fixtures, dist",
		Chunker:                falkenvector.ChunkerMarkdown,
		ChunkSizeInput:         "1200",
		ChunkOverlapInput:      "200",
		SyncSource:             true,
	}, true)
	if err != nil {
		t.Fatalf("BuildIngestRequest: %v", err)
	}
	if request.Root != "." || !request.DryRun || !request.SyncSource || request.ChunkerMode != falkenvector.ChunkerMarkdown {
		t.Fatalf("request = %+v", request)
	}
	if request.EmbeddingConcurrency != defaultEmbeddingConcurrency {
		t.Fatalf("EmbeddingConcurrency = %d, want %d", request.EmbeddingConcurrency, defaultEmbeddingConcurrency)
	}
	if !reflect.DeepEqual(request.Extensions, []string{"go", "md"}) {
		t.Fatalf("extensions = %+v", request.Extensions)
	}
	if !reflect.DeepEqual(request.ExcludeExtensions, []string{"tmp", "log"}) {
		t.Fatalf("exclude extensions = %+v", request.ExcludeExtensions)
	}
	if !reflect.DeepEqual(request.ExcludeDirs, []string{"fixtures", "dist"}) {
		t.Fatalf("exclude dirs = %+v", request.ExcludeDirs)
	}
}

func TestBuildIngestRequestValidation(t *testing.T) {
	_, err := BuildIngestRequest(IndexViewModel{Chunker: falkenvector.ChunkerAuto, ChunkSizeInput: "10", ChunkOverlapInput: "10"}, true)
	if err == nil || !strings.Contains(err.Error(), "less than chunk size") {
		t.Fatalf("error = %v, want overlap validation", err)
	}
	_, err = BuildIngestRequest(IndexViewModel{Chunker: falkenvector.ChunkerMode("bogus"), ChunkSizeInput: "10", ChunkOverlapInput: "1"}, true)
	if err == nil || !strings.Contains(err.Error(), "chunker") {
		t.Fatalf("error = %v, want chunker validation", err)
	}
}

func TestFormatIngestResult(t *testing.T) {
	got := FormatIngestResult(falkenvector.IngestResult{Scanned: 3, NewFiles: 1, ChangedFiles: 2, ChunksEmbedded: 4})
	if !strings.Contains(got, "scanned: 3") || !strings.Contains(got, "chunks embedded: 4") {
		t.Fatalf("FormatIngestResult = %q", got)
	}
}
