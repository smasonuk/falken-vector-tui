package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/smasonuk/falken-vector/pkg/falkenvector"
)

func BuildAskRequest(directory string, model AskViewModel) (falkenvector.AskRequest, error) {
	if strings.TrimSpace(model.Question) == "" {
		return falkenvector.AskRequest{}, errors.New("question is required")
	}
	request := falkenvector.AskRequest{
		Question: strings.TrimSpace(model.Question),
		Retrieval: falkenvector.RetrievalOptions{
			Mode:        falkenvector.RetrievalHybrid,
			TopK:        TOPK,
			SourceRoots: SelectedAskSourceRoots(model),
		},
		Agent: true,
	}
	if model.AttachSelectedDocuments {
		documents := SelectedAskSourceDocuments(model)
		if len(documents) == 0 {
			return falkenvector.AskRequest{}, errors.New("select at least one source to attach documents whole")
		}
		attached, err := LoadAskAttachedDocuments(directory, documents)
		if err != nil {
			return falkenvector.AskRequest{}, err
		}
		request.AttachedDocuments = attached
	}
	return request, nil
}

func AskViewFromAnswer(answer falkenvector.Answer, directory string) AskViewModel {
	vm := AskViewModel{
		Answer: answer.Text,
	}
	for _, source := range answer.Sources {
		item := AskSourceViewFromSDK(source, directory)
		vm.SourceItems = append(vm.SourceItems, item)
		vm.Sources = append(vm.Sources, item.Label)
	}
	vm.Warnings = append(vm.Warnings, answer.CitationWarnings...)
	vm.Warnings = append(vm.Warnings, answer.CoverageWarnings...)
	vm.Warnings = append(vm.Warnings, answer.ThinSourceWarnings...)
	for _, call := range answer.ToolCalls {
		vm.ToolCalls = append(vm.ToolCalls, call.Name)
	}
	return vm
}

func AskSourceViewFromSDK(source falkenvector.Source, directory string) AskSourceView {
	displayPath := relativeDisplayPath(directory, source.Path)
	return AskSourceView{
		Label:       FormatAskSourceLabel(source.Number, displayPath, source.StartLine, source.EndLine),
		Path:        source.Path,
		DisplayPath: displayPath,
		StartLine:   source.StartLine,
		EndLine:     source.EndLine,
	}
}

func FormatAskSourceLabel(number int, path string, startLine, endLine int) string {
	return fmt.Sprintf("[%d] %s:%d-%d", number, path, startLine, endLine)
}

func LoadAskSourcePreview(directory string, source AskSourceView) (string, error) {
	path := sourcePath(directory, source.Path)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read source: %w", err)
	}
	lines := strings.SplitAfter(string(data), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return "", nil
	}
	start, end := normalizeSourceLines(source.StartLine, source.EndLine, len(lines))
	if start > len(lines) {
		return "", fmt.Errorf("source line %d is outside %s (%d lines)", source.StartLine, source.DisplayPath, len(lines))
	}
	var b strings.Builder
	width := len(strconv.Itoa(end))
	for lineNo := start; lineNo <= end; lineNo++ {
		line := strings.TrimSuffix(strings.TrimSuffix(lines[lineNo-1], "\n"), "\r")
		fmt.Fprintf(&b, "%*d  %s", width, lineNo, line)
		if lineNo != end {
			b.WriteByte('\n')
		}
	}
	return b.String(), nil
}

func sourcePath(directory, path string) string {
	cleanPath := filepath.Clean(path)
	if filepath.IsAbs(cleanPath) {
		return cleanPath
	}
	if strings.TrimSpace(directory) == "" {
		return cleanPath
	}
	return filepath.Join(directory, cleanPath)
}

func normalizeSourceLines(start, end, lineCount int) (int, int) {
	if start < 1 {
		start = 1
	}
	if end < start {
		end = start
	}
	if end > lineCount {
		end = lineCount
	}
	return start, end
}

func relativeDisplayPath(directory, path string) string {
	if strings.TrimSpace(path) == "" {
		return path
	}
	cleanPath := filepath.Clean(path)
	if !filepath.IsAbs(cleanPath) {
		return cleanPath
	}
	base := strings.TrimSpace(directory)
	if base == "" {
		return cleanPath
	}
	rel, err := filepath.Rel(base, cleanPath)
	if err != nil {
		return cleanPath
	}
	return rel
}

func FormatAskSidePanel(answer AskViewModel) string {
	var b strings.Builder
	if len(answer.Sources) != 0 {
		b.WriteString("Sources:\n")
		for _, source := range answer.Sources {
			fmt.Fprintf(&b, "%s\n", source)
		}
	}
	if len(answer.Warnings) != 0 {
		if b.Len() != 0 {
			b.WriteString("\n")
		}
		b.WriteString("Warnings:\n")
		for _, warning := range answer.Warnings {
			fmt.Fprintf(&b, "- %s\n", warning)
		}
	}
	if len(answer.ToolCalls) != 0 {
		if b.Len() != 0 {
			b.WriteString("\n")
		}
		b.WriteString("Tool calls:\n")
		for _, call := range answer.ToolCalls {
			fmt.Fprintf(&b, "- %s\n", call)
		}
	}
	return strings.TrimSpace(b.String())
}

func FormatAskNotesPanel(answer AskViewModel) string {
	var b strings.Builder
	if len(answer.Warnings) != 0 {
		b.WriteString("Warnings:\n")
		for _, warning := range answer.Warnings {
			fmt.Fprintf(&b, "- %s\n", warning)
		}
	}
	if len(answer.ToolCalls) != 0 {
		if b.Len() != 0 {
			b.WriteString("\n")
		}
		b.WriteString("Tool calls:\n")
		for _, call := range answer.ToolCalls {
			fmt.Fprintf(&b, "- %s\n", call)
		}
	}
	return strings.TrimSpace(b.String())
}
