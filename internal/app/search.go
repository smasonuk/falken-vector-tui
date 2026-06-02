package app

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/smasonuk/falken-vector/pkg/falkenvector"
)

func ParsePositiveInt(input string, field string) (int, error) {
	value, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be an integer > 0", field)
	}
	return value, nil
}

func ValidateRetrievalMode(mode falkenvector.RetrievalMode) error {
	switch mode {
	case falkenvector.RetrievalLexical, falkenvector.RetrievalVector, falkenvector.RetrievalHybrid:
		return nil
	default:
		return fmt.Errorf("retrieval mode must be lexical, vector, or hybrid")
	}
}

func BuildQueryRequest(model SearchViewModel) (falkenvector.QueryRequest, error) {
	if strings.TrimSpace(model.Question) == "" {
		return falkenvector.QueryRequest{}, errors.New("question is required")
	}
	return falkenvector.QueryRequest{
		Question: strings.TrimSpace(model.Question),
		Retrieval: falkenvector.RetrievalOptions{
			Mode: falkenvector.RetrievalHybrid,
			TopK: TOPK,
		},
	}, nil
}

func SearchViewFromResult(result falkenvector.QueryResult) SearchViewModel {
	items := make([]SearchResultView, 0, len(result.Chunks))
	for i, chunk := range result.Chunks {
		items = append(items, SearchResultView{
			Title:   FormatSearchResultTitle(i, chunk),
			Preview: FormatChunkPreview(chunk),
			Chunk:   chunk,
		})
	}
	return SearchViewModel{
		Question:  result.Question,
		Mode:      falkenvector.RetrievalMode(result.QueryPlan.Mode),
		TopKInput: strconv.Itoa(TOPK),
		Results:   items,
		QueryPlan: FormatQueryPlan(result.QueryPlan),
	}
}

func FormatSearchResultTitle(index int, chunk falkenvector.RetrievedChunk) string {
	return fmt.Sprintf("[%d] score=%.3f %s:%d-%d", index+1, chunk.Score, chunk.Path, chunk.StartLine, chunk.EndLine)
}

func FormatChunkPreview(chunk falkenvector.RetrievedChunk) string {
	return fmt.Sprintf("%s:%d-%d\n\n%s", chunk.Path, chunk.StartLine, chunk.EndLine, chunk.Text)
}

func FormatQueryPlan(plan falkenvector.QueryPlan) string {
	var b strings.Builder
	if plan.Mode != "" {
		fmt.Fprintf(&b, "mode: %s\n", plan.Mode)
	}
	if len(plan.Queries) != 0 {
		b.WriteString("queries:\n")
		for _, query := range plan.Queries {
			fmt.Fprintf(&b, "- %s\n", query)
		}
	}
	if plan.Warning != "" {
		fmt.Fprintf(&b, "warning: %s\n", plan.Warning)
	}
	return strings.TrimSpace(b.String())
}

func NoIndexSearchMessage() string {
	return "No index found for this directory. Go to Index and run \"Index this directory\"."
}
