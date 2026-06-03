package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/smasonuk/falken-vector/pkg/falkenvector"
)

const fullDocumentTokenWarningThreshold = 100000

type AskSourcePickerResult struct {
	Documents  []AskSourceDocument
	Roots      []AskSourceTreeItem
	Children   map[string][]AskSourceTreeItem
	SelectedID string
	Expanded   map[string]bool
	Checked    map[string]bool
}

func BuildAskSourcePicker(directory string, documents []falkenvector.IndexedDocument, checked map[string]bool) AskSourcePickerResult {
	docs := normalizeAskSourceDocuments(directory, documents)
	roots, children := buildAskSourceTree(directory, docs)
	expanded := make(map[string]bool, len(roots))
	selectedID := ""
	for _, root := range roots {
		expanded[root.ID] = true
		if selectedID == "" {
			selectedID = root.ID
		}
	}
	return AskSourcePickerResult{
		Documents:  docs,
		Roots:      roots,
		Children:   children,
		SelectedID: selectedID,
		Expanded:   expanded,
		Checked:    normalizeAskSourceCheckedWithExpansion(roots, children, checked),
	}
}

func ApplyAskSourcePickerResult(model AskViewModel, result AskSourcePickerResult) AskViewModel {
	model.SourcePickerLoading = false
	model.SourcePickerError = ""
	model.SourcePickerDocuments = result.Documents
	model.SourcePickerRoots = result.Roots
	model.SourcePickerChildren = result.Children
	model.SourcePickerSelectedID = result.SelectedID
	model.SourcePickerExpanded = result.Expanded
	model.SourcePickerChecked = result.Checked
	return model
}

func BeginAskSourcePickerLoad(model AskViewModel) AskViewModel {
	model.SourcePickerOpen = true
	model.SourcePickerLoading = true
	model.SourcePickerError = ""
	model.SourcePickerChecked = cloneBoolMap(model.SelectedSourceIDs)
	model.SourcePickerAttachSelectedDocuments = model.AttachSelectedDocuments
	if model.SourcePickerExpanded == nil {
		model.SourcePickerExpanded = map[string]bool{}
	}
	return model
}

func CloseAskSourcePicker(model AskViewModel) AskViewModel {
	model.SourcePickerOpen = false
	model.SourcePickerLoading = false
	model.SourcePickerError = ""
	model.SourcePickerChecked = cloneBoolMap(model.SelectedSourceIDs)
	model.SourcePickerAttachSelectedDocuments = model.AttachSelectedDocuments
	return model
}

func ApplyAskSourcePickerSelection(model AskViewModel) AskViewModel {
	model.SelectedSourceIDs = normalizeAskSourceChecked(model.SourcePickerRoots, model.SourcePickerChildren, model.SourcePickerChecked)
	model.AttachSelectedDocuments = model.SourcePickerAttachSelectedDocuments
	model.SourcePickerOpen = false
	model.SourcePickerLoading = false
	model.SourcePickerError = ""
	return model
}

func ClearAskSourcePickerDraft(model AskViewModel) AskViewModel {
	model.SourcePickerChecked = map[string]bool{}
	model.SourcePickerAttachSelectedDocuments = false
	return model
}

func SelectAskSourcePickerItem(model AskViewModel, id string) AskViewModel {
	model.SourcePickerSelectedID = id
	return model
}

func ExpandAskSourcePickerItem(model AskViewModel, id string, expanded bool) AskViewModel {
	next := cloneBoolMap(model.SourcePickerExpanded)
	if expanded {
		next[id] = true
	} else {
		delete(next, id)
	}
	model.SourcePickerExpanded = next
	return model
}

func CheckAskSourcePickerItem(model AskViewModel, id string, checked bool) AskViewModel {
	next := cloneBoolMap(model.SourcePickerChecked)
	setAskSourceCheckedRecursive(next, model.SourcePickerChildren, id, checked)
	model.SourcePickerChecked = normalizeAskSourceChecked(model.SourcePickerRoots, model.SourcePickerChildren, next)
	return model
}

func SetAskSourcePickerAttachDocuments(model AskViewModel, value bool) AskViewModel {
	model.SourcePickerAttachSelectedDocuments = value
	return model
}

func SelectedAskSourceRoots(model AskViewModel) []string {
	return checkedAskSourceRoots(model.SelectedSourceIDs)
}

func SelectedAskSourceDocuments(model AskViewModel) []AskSourceDocument {
	return selectedAskSourceDocuments(model.SourcePickerDocuments, model.SelectedSourceIDs)
}

func SourcePickerSelectedDocuments(model AskViewModel) []AskSourceDocument {
	return selectedAskSourceDocuments(model.SourcePickerDocuments, model.SourcePickerChecked)
}

func AskSourceDocumentsTokenEstimate(documents []AskSourceDocument) int {
	total := 0
	for _, doc := range documents {
		total += doc.TokenEstimate
	}
	return total
}

func FormatAskSourceSelectionSummary(model AskViewModel) string {
	documents := SelectedAskSourceDocuments(model)
	if len(model.SelectedSourceIDs) == 0 || len(documents) == 0 {
		if len(model.SelectedSourceIDs) != 0 {
			return fmt.Sprintf("Sources: %d selected", len(model.SelectedSourceIDs))
		}
		if model.AttachSelectedDocuments {
			return "Sources: none selected | attach whole docs"
		}
		return "Sources: all indexed"
	}
	summary := fmt.Sprintf("Sources: %d selected", len(documents))
	if model.AttachSelectedDocuments {
		summary += fmt.Sprintf(" | attach whole docs | ~%d tokens", AskSourceDocumentsTokenEstimate(documents))
	}
	return summary
}

func FormatAskSourcePickerStats(model AskViewModel) string {
	documents := SourcePickerSelectedDocuments(model)
	if len(documents) == 0 {
		if model.SourcePickerAttachSelectedDocuments {
			return "No sources selected. Select at least one source to attach documents whole."
		}
		return "No sources selected. Ask will use the full indexed corpus."
	}
	if model.SourcePickerAttachSelectedDocuments {
		return fmt.Sprintf("Selected documents: %d | rough tokens: %d", len(documents), AskSourceDocumentsTokenEstimate(documents))
	}
	return fmt.Sprintf("Selected documents: %d", len(documents))
}

func AskSourcePickerWarning(model AskViewModel) string {
	if !model.SourcePickerAttachSelectedDocuments {
		return ""
	}
	tokens := AskSourceDocumentsTokenEstimate(SourcePickerSelectedDocuments(model))
	if tokens <= fullDocumentTokenWarningThreshold {
		return ""
	}
	return fmt.Sprintf("Large context warning: roughly %d tokens.", tokens)
}

func LoadAskAttachedDocuments(directory string, documents []AskSourceDocument) ([]falkenvector.AttachedDocument, error) {
	out := make([]falkenvector.AttachedDocument, 0, len(documents))
	for _, doc := range documents {
		path := sourcePath(directory, doc.Path)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read attached source %s: %w", doc.DisplayPath, err)
		}
		text := string(data)
		lineCount := countAskDocumentLines(text)
		if lineCount <= 0 {
			lineCount = 1
		}
		out = append(out, falkenvector.AttachedDocument{
			Path:       path,
			SourceRoot: doc.SourceRoot,
			Text:       text,
			StartLine:  1,
			EndLine:    lineCount,
		})
	}
	return out, nil
}

func normalizeAskSourceDocuments(directory string, documents []falkenvector.IndexedDocument) []AskSourceDocument {
	out := make([]AskSourceDocument, 0, len(documents))
	seen := make(map[string]struct{}, len(documents))
	for _, doc := range documents {
		path := strings.TrimSpace(doc.Path)
		if path == "" {
			continue
		}
		path = filepath.Clean(sourcePath(directory, path))
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		sourceRoot := strings.TrimSpace(doc.SourceRoot)
		if sourceRoot != "" {
			sourceRoot = filepath.Clean(sourcePath(directory, sourceRoot))
		}
		out = append(out, AskSourceDocument{
			Path:          path,
			DisplayPath:   relativeDisplayPath(directory, path),
			SourceRoot:    sourceRoot,
			SizeBytes:     doc.SizeBytes,
			TokenEstimate: estimateAskSourceTokens(doc.SizeBytes),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Path < out[j].Path
	})
	return out
}

func buildAskSourceTree(directory string, documents []AskSourceDocument) ([]AskSourceTreeItem, map[string][]AskSourceTreeItem) {
	children := map[string][]AskSourceTreeItem{}
	rootSeen := map[string]bool{}
	var roots []AskSourceTreeItem
	base := normalizedAskSourceBase(directory)
	docRoots := askSourceDocumentRoots(base, documents)
	for _, doc := range documents {
		root := docRoots[doc.Path]
		if root == "" {
			continue
		}
		hideRoot := base != "" && root == base
		if !hideRoot && !rootSeen[root] {
			rootSeen[root] = true
			roots = append(roots, AskSourceTreeItem{ID: root, Label: askSourceRootLabel(directory, root), IsBranch: true})
		}
		rel, err := filepath.Rel(root, doc.Path)
		if err != nil || rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." || filepath.IsAbs(rel) {
			rel = filepath.Base(doc.Path)
		}
		parent := root
		parts := strings.Split(filepath.Clean(rel), string(filepath.Separator))
		for i, part := range parts {
			if part == "." || part == "" {
				continue
			}
			id := filepath.Join(parent, part)
			isBranch := i < len(parts)-1
			item := AskSourceTreeItem{ID: id, Label: part, IsBranch: isBranch}
			if hideRoot && i == 0 {
				if !rootSeen[id] {
					rootSeen[id] = true
					roots = append(roots, item)
				} else {
					mergeAskSourceItem(roots, item)
				}
			} else {
				appendAskSourceChild(children, parent, item)
			}
			parent = id
		}
	}
	sortAskSourceItems(roots)
	for parent, items := range children {
		sortAskSourceItems(items)
		children[parent] = items
	}
	return roots, children
}

func askSourceDocumentRoots(base string, documents []AskSourceDocument) map[string]string {
	candidates := make([]string, 0, len(documents))
	byPath := make(map[string]string, len(documents))
	for _, doc := range documents {
		root := askSourceRootForDocument(base, doc)
		byPath[doc.Path] = root
		if root != "" {
			candidates = append(candidates, root)
		}
	}
	sort.Strings(candidates)
	for path, root := range byPath {
		byPath[path] = shallowestCoveringAskSourceRoot(root, candidates)
	}
	return byPath
}

func askSourceRootForDocument(base string, doc AskSourceDocument) string {
	if base != "" && askSourcePathWithin(base, doc.Path) {
		return base
	}
	if doc.SourceRoot != "" && askSourcePathWithin(doc.SourceRoot, doc.Path) {
		return doc.SourceRoot
	}
	return filepath.Dir(doc.Path)
}

func normalizedAskSourceBase(directory string) string {
	base := strings.TrimSpace(directory)
	if base == "" {
		return ""
	}
	return filepath.Clean(base)
}

func shallowestCoveringAskSourceRoot(root string, candidates []string) string {
	if root == "" {
		return ""
	}
	best := root
	for _, candidate := range candidates {
		if candidate == "" || candidate == root {
			continue
		}
		if askSourcePathWithin(candidate, root) && len(candidate) < len(best) {
			best = candidate
		}
	}
	return best
}

func askSourceRootLabel(directory string, root string) string {
	display := relativeDisplayPath(directory, root)
	if display == "." || display == "" {
		display = filepath.Base(root)
	}
	if display == "." || display == string(filepath.Separator) || display == "" {
		display = root
	}
	return display
}

func mergeAskSourceItem(items []AskSourceTreeItem, item AskSourceTreeItem) {
	for i, existing := range items {
		if existing.ID != item.ID {
			continue
		}
		if item.IsBranch && !existing.IsBranch {
			items[i].IsBranch = true
		}
		return
	}
}

func appendAskSourceChild(children map[string][]AskSourceTreeItem, parent string, item AskSourceTreeItem) {
	for i, existing := range children[parent] {
		if existing.ID != item.ID {
			continue
		}
		if item.IsBranch && !existing.IsBranch {
			children[parent][i].IsBranch = true
		}
		return
	}
	children[parent] = append(children[parent], item)
}

func sortAskSourceItems(items []AskSourceTreeItem) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].IsBranch != items[j].IsBranch {
			return items[i].IsBranch
		}
		return strings.ToLower(items[i].Label) < strings.ToLower(items[j].Label)
	})
}

func normalizeAskSourceChecked(roots []AskSourceTreeItem, children map[string][]AskSourceTreeItem, checked map[string]bool) map[string]bool {
	valid := askSourceValidIDs(roots, children)
	next := make(map[string]bool, len(checked))
	for id, value := range checked {
		if value && valid[id] {
			next[id] = true
		}
	}
	recomputeAskSourceBranchChecks(next, roots, children)
	return next
}

func normalizeAskSourceCheckedWithExpansion(roots []AskSourceTreeItem, children map[string][]AskSourceTreeItem, checked map[string]bool) map[string]bool {
	valid := askSourceValidIDs(roots, children)
	next := make(map[string]bool, len(checked))
	for id, value := range checked {
		if value && valid[id] {
			setAskSourceCheckedRecursive(next, children, id, true)
		}
	}
	recomputeAskSourceBranchChecks(next, roots, children)
	return next
}

func askSourceValidIDs(roots []AskSourceTreeItem, children map[string][]AskSourceTreeItem) map[string]bool {
	valid := make(map[string]bool)
	var walk func(AskSourceTreeItem)
	walk = func(item AskSourceTreeItem) {
		if valid[item.ID] {
			return
		}
		valid[item.ID] = true
		for _, child := range children[item.ID] {
			walk(child)
		}
	}
	for _, root := range roots {
		walk(root)
	}
	return valid
}

func setAskSourceCheckedRecursive(checked map[string]bool, children map[string][]AskSourceTreeItem, id string, value bool) {
	if value {
		checked[id] = true
	} else {
		delete(checked, id)
	}
	for _, child := range children[id] {
		setAskSourceCheckedRecursive(checked, children, child.ID, value)
	}
}

func recomputeAskSourceBranchChecks(checked map[string]bool, roots []AskSourceTreeItem, children map[string][]AskSourceTreeItem) {
	var walk func(AskSourceTreeItem) bool
	walk = func(item AskSourceTreeItem) bool {
		kids := children[item.ID]
		if len(kids) == 0 {
			return checked[item.ID]
		}
		all := true
		for _, child := range kids {
			if !walk(child) {
				all = false
			}
		}
		if all {
			checked[item.ID] = true
		} else {
			delete(checked, item.ID)
		}
		return all
	}
	for _, root := range roots {
		walk(root)
	}
}

func checkedAskSourceRoots(checked map[string]bool) []string {
	ids := make([]string, 0, len(checked))
	for id, value := range checked {
		if value {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		covered := false
		for _, existing := range out {
			if askSourcePathWithin(existing, id) {
				covered = true
				break
			}
		}
		if !covered {
			out = append(out, id)
		}
	}
	return out
}

func selectedAskSourceDocuments(documents []AskSourceDocument, checked map[string]bool) []AskSourceDocument {
	roots := checkedAskSourceRoots(checked)
	if len(roots) == 0 {
		return nil
	}
	out := make([]AskSourceDocument, 0, len(documents))
	for _, doc := range documents {
		for _, root := range roots {
			if askSourcePathWithin(root, doc.Path) {
				out = append(out, doc)
				break
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].DisplayPath < out[j].DisplayPath
	})
	return out
}

func estimateAskSourceTokens(sizeBytes int64) int {
	if sizeBytes <= 0 {
		return 0
	}
	tokens := int(sizeBytes / 4)
	if tokens == 0 {
		return 1
	}
	return tokens
}

func countAskDocumentLines(text string) int {
	if text == "" {
		return 0
	}
	lines := strings.Count(text, "\n")
	if !strings.HasSuffix(text, "\n") {
		lines++
	}
	return lines
}

func askSourcePathWithin(root string, path string) bool {
	if strings.TrimSpace(root) == "" || strings.TrimSpace(path) == "" {
		return false
	}
	root = filepath.Clean(root)
	path = filepath.Clean(path)
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel))
}

func cloneBoolMap(values map[string]bool) map[string]bool {
	if len(values) == 0 {
		return map[string]bool{}
	}
	out := make(map[string]bool, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}
