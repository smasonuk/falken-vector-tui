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
	if request.Question != "alpha" || request.Retrieval.Mode != falkenvector.RetrievalHybrid || request.Retrieval.TopK != TOPK {
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
	request, err := BuildAskRequest("/repo", AskViewModel{
		Question:          "alpha",
		Mode:              falkenvector.RetrievalLexical,
		TopKInput:         "5",
		Agent:             false,
		SelectedSourceIDs: map[string]bool{"/repo/docs": true},
	})
	if err != nil {
		t.Fatalf("BuildAskRequest: %v", err)
	}
	if !request.Agent || request.Retrieval.Mode != falkenvector.RetrievalHybrid || request.Retrieval.TopK != TOPK {
		t.Fatalf("request = %+v", request)
	}
	if len(request.Retrieval.SourceRoots) != 1 || request.Retrieval.SourceRoots[0] != "/repo/docs" {
		t.Fatalf("source roots = %+v", request.Retrieval.SourceRoots)
	}
	if !strings.Contains(request.SourceScopeNote, "selected sources") || !strings.Contains(request.SourceScopeNote, "search_index") {
		t.Fatalf("SourceScopeNote = %q, want selected-source agent instruction", request.SourceScopeNote)
	}
}

func TestBuildAskRequestNoSelectedSourcesDoesNotSetScopeNote(t *testing.T) {
	request, err := BuildAskRequest("/repo", AskViewModel{Question: "alpha"})
	if err != nil {
		t.Fatalf("BuildAskRequest: %v", err)
	}
	if request.SourceScopeNote != "" {
		t.Fatalf("SourceScopeNote = %q, want empty for unfiltered ask", request.SourceScopeNote)
	}
}

func TestBuildAskRequestAttachSelectedDocuments(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "docs", "a.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte("alpha\nbeta\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	request, err := BuildAskRequest(directory, AskViewModel{
		Question:                "alpha",
		AttachSelectedDocuments: true,
		SelectedSourceIDs:       map[string]bool{path: true},
		SourcePickerDocuments: []AskSourceDocument{{
			Path:        path,
			DisplayPath: filepath.Join("docs", "a.md"),
			SourceRoot:  directory,
		}},
	})
	if err != nil {
		t.Fatalf("BuildAskRequest: %v", err)
	}
	if len(request.AttachedDocuments) != 1 {
		t.Fatalf("attached = %+v, want one", request.AttachedDocuments)
	}
	if request.AttachedDocuments[0].Text != "alpha\nbeta\n" || request.AttachedDocuments[0].EndLine != 2 {
		t.Fatalf("attached doc = %+v", request.AttachedDocuments[0])
	}
	if request.SourceScopeNote != "" {
		t.Fatalf("SourceScopeNote = %q, want empty for attached documents", request.SourceScopeNote)
	}
}

func TestBuildAskRequestAttachSelectedDocumentsRequiresSelection(t *testing.T) {
	_, err := BuildAskRequest(t.TempDir(), AskViewModel{Question: "alpha", AttachSelectedDocuments: true})
	if err == nil {
		t.Fatal("BuildAskRequest succeeded, want selection error")
	}
	if !strings.Contains(err.Error(), "select at least one source") {
		t.Fatalf("err = %v", err)
	}
}

func TestAskSourcePickerBuildsTreeAndChecksDirectoryRecursively(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "repo")
	docsDir := filepath.Join(directory, "docs")
	srcDir := filepath.Join(directory, "src")
	aPath := filepath.Join(docsDir, "a.md")
	bPath := filepath.Join(docsDir, "b.md")
	mainPath := filepath.Join(srcDir, "main.go")
	result := BuildAskSourcePicker(directory, []falkenvector.IndexedDocument{
		{Path: aPath, SourceRoot: directory, SizeBytes: 8},
		{Path: bPath, SourceRoot: directory, SizeBytes: 16},
		{Path: mainPath, SourceRoot: directory, SizeBytes: 4},
	}, nil)
	if len(result.Roots) != 2 || result.Roots[0].ID != docsDir || result.Roots[1].ID != srcDir {
		t.Fatalf("roots = %+v", result.Roots)
	}
	if len(result.Children[docsDir]) != 2 || result.Children[docsDir][0].Label != "a.md" || result.Children[docsDir][1].Label != "b.md" {
		t.Fatalf("docs children = %+v", result.Children[docsDir])
	}

	model := AskViewModel{
		SourcePickerDocuments: result.Documents,
		SourcePickerRoots:     result.Roots,
		SourcePickerChildren:  result.Children,
		SourcePickerChecked:   result.Checked,
	}
	model = CheckAskSourcePickerItem(model, docsDir, true)
	selected := SourcePickerSelectedDocuments(model)
	if len(selected) != 2 || AskSourceDocumentsTokenEstimate(selected) != 6 {
		t.Fatalf("selected = %+v tokens=%d", selected, AskSourceDocumentsTokenEstimate(selected))
	}
	model = CheckAskSourcePickerItem(model, aPath, false)
	selected = SourcePickerSelectedDocuments(model)
	if len(selected) != 1 || selected[0].Path != bPath {
		t.Fatalf("selected after uncheck = %+v", selected)
	}
	if model.SourcePickerChecked[docsDir] {
		t.Fatalf("directory remained checked after child uncheck: %+v", model.SourcePickerChecked)
	}
}

func TestAskSourcePickerCollapsesOverlappingSourceRoots(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "test_vec")
	sourceDir := filepath.Join(directory, "source")
	alphaPath := filepath.Join(sourceDir, "alpha.md")
	betaPath := filepath.Join(sourceDir, "nested", "beta.md")
	result := BuildAskSourcePicker(directory, []falkenvector.IndexedDocument{
		{Path: alphaPath, SourceRoot: directory, SizeBytes: 8},
		{Path: betaPath, SourceRoot: sourceDir, SizeBytes: 12},
	}, nil)
	if len(result.Roots) != 1 || result.Roots[0].ID != sourceDir || result.Roots[0].Label != "source" {
		t.Fatalf("roots = %+v, want single source root", result.Roots)
	}
	if len(result.Children[directory]) != 0 {
		t.Fatalf("hidden current directory should not have rendered children: %+v", result.Children[directory])
	}
	if len(result.Children[sourceDir]) != 2 {
		t.Fatalf("source children = %+v, want merged docs from both indexed roots", result.Children[sourceDir])
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
