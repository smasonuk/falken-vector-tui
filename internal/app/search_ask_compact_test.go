package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/smasonuk/falken-vector/pkg/falkenvector"
)

func TestBuildQueryRequest(t *testing.T) {
	request, err := BuildQueryRequest(SearchViewModel{
		Question:  " alpha ",
		Mode:      falkenvector.RetrievalLexical,
		TopKInput: "99",
	})
	if err != nil {
		t.Fatalf("BuildQueryRequest: %v", err)
	}
	if request.Question != "alpha" || request.Retrieval.Mode != falkenvector.RetrievalHybrid || request.Retrieval.TopK != 8 {
		t.Fatalf("request = %+v", request)
	}
}

func TestBuildQueryRequestBlocksEmptyQuestion(t *testing.T) {
	if _, err := BuildQueryRequest(SearchViewModel{Mode: falkenvector.RetrievalLexical, TopKInput: "8"}); err == nil {
		t.Fatal("BuildQueryRequest succeeded, want error")
	}
}

func TestFormatSearchResultAndNoIndexMessage(t *testing.T) {
	chunk := falkenvector.RetrievedChunk{Source: falkenvector.Source{Path: "a.go", StartLine: 1, EndLine: 3}, Score: 0.75, Text: "alpha"}
	if got := FormatSearchResultTitle(0, chunk); !strings.Contains(got, "score=0.750 a.go:1-3") {
		t.Fatalf("title = %q", got)
	}
	if got := NoIndexSearchMessage(); !strings.Contains(got, "No index found") {
		t.Fatalf("message = %q", got)
	}
}

func TestBuildAskRequest(t *testing.T) {
	request, err := BuildAskRequest(AskViewModel{Question: "alpha", Mode: falkenvector.RetrievalLexical, TopKInput: "5", Agent: false})
	if err != nil {
		t.Fatalf("BuildAskRequest: %v", err)
	}
	if !request.Agent || request.Retrieval.Mode != falkenvector.RetrievalHybrid || request.Retrieval.TopK != 8 {
		t.Fatalf("request = %+v", request)
	}
}

func TestAskViewFromAnswerUsesRelativeSourcePaths(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "repo")
	answer := falkenvector.Answer{
		Sources: []falkenvector.Source{
			{
				Number:    1,
				Path:      filepath.Join(directory, "internal", "app", "ask.go"),
				StartLine: 10,
				EndLine:   12,
			},
		},
	}
	view := AskViewFromAnswer(answer, directory)
	if len(view.Sources) != 1 {
		t.Fatalf("Sources = %d, want 1", len(view.Sources))
	}
	if strings.Contains(view.Sources[0], directory) {
		t.Fatalf("source path = %q, want relative path", view.Sources[0])
	}
	if !strings.Contains(view.Sources[0], filepath.Join("internal", "app", "ask.go")+":10-12") {
		t.Fatalf("source path = %q", view.Sources[0])
	}
	if len(view.SourceItems) != 1 {
		t.Fatalf("SourceItems = %d, want 1", len(view.SourceItems))
	}
	if view.SourceItems[0].Path != filepath.Join(directory, "internal", "app", "ask.go") {
		t.Fatalf("SourceItems[0].Path = %q", view.SourceItems[0].Path)
	}
}

func TestLoadAskSourcePreviewReadsLineSpan(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "source.go")
	if err := os.WriteFile(path, []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	source := AskSourceView{
		Path:        "source.go",
		DisplayPath: "source.go",
		StartLine:   1,
		EndLine:     3,
	}
	text, err := LoadAskSourcePreview(directory, source)
	if err != nil {
		t.Fatalf("LoadAskSourcePreview: %v", err)
	}
	if !strings.Contains(text, "1  package main") || !strings.Contains(text, "3  func main() {}") {
		t.Fatalf("preview = %q", text)
	}
}

func TestLoadAskSourcePreviewReportsMissingFile(t *testing.T) {
	_, err := LoadAskSourcePreview(t.TempDir(), AskSourceView{Path: "missing.go", DisplayPath: "missing.go", StartLine: 1, EndLine: 1})
	if err == nil {
		t.Fatal("LoadAskSourcePreview err = nil, want error")
	}
	if !strings.Contains(err.Error(), "read source") {
		t.Fatalf("err = %q", err)
	}
}

func TestBuildCompactRequest(t *testing.T) {
	request, err := BuildCompactRequest(CompactViewModel{BatchSizeInput: "100", KeepBackup: true}, true)
	if err != nil {
		t.Fatalf("BuildCompactRequest: %v", err)
	}
	if !request.DryRun || request.BatchSize != 100 || !request.KeepBackup || request.EmbeddingConcurrency != defaultEmbeddingConcurrency {
		t.Fatalf("request = %+v", request)
	}
}
