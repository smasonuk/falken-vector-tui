package app

import (
	"strings"
	"testing"

	"github.com/smasonuk/falken-vector/pkg/falkenvector"
)

func TestFormatSDKEventCoversKnownTypes(t *testing.T) {
	cases := []struct {
		event falkenvector.Event
		want  string
	}{
		{falkenvector.Event{Type: falkenvector.EventRunStarted, Message: "ingest"}, "run.started ingest"},
		{falkenvector.Event{Type: falkenvector.EventRunCompleted, Message: "done"}, "run.completed done"},
		{falkenvector.Event{Type: falkenvector.EventRunFailed, Error: "bad"}, "run.failed bad"},
		{falkenvector.Event{Type: falkenvector.EventRetrievalStarted, Retrieval: &falkenvector.RetrievalEvent{Mode: "lexical", TopK: 8}}, "retrieval.started mode=lexical top_k=8"},
		{falkenvector.Event{Type: falkenvector.EventRetrievalCompleted, Retrieval: &falkenvector.RetrievalEvent{ResultCount: 3}}, "retrieval.completed results=3"},
		{falkenvector.Event{Type: falkenvector.EventRetrievalFailed, Error: "missing"}, "retrieval.failed missing"},
		{falkenvector.Event{Type: falkenvector.EventIngestStarted, Ingest: &falkenvector.IngestEvent{Directory: "/repo"}}, "ingest.started /repo"},
		{falkenvector.Event{Type: falkenvector.EventIngestProgress, Ingest: &falkenvector.IngestEvent{Action: "indexing", CurrentPath: "README.md", CurrentFile: 1, TotalFiles: 2}}, "ingest.progress indexing README.md 1/2"},
		{falkenvector.Event{Type: falkenvector.EventIngestCompleted, Ingest: &falkenvector.IngestEvent{Scanned: 2, NewFiles: 1}}, "ingest.completed scanned=2 changed=0 new=1 failed=0"},
		{falkenvector.Event{Type: falkenvector.EventEmbeddingRequest, Embedding: &falkenvector.EmbeddingEvent{InputLength: 12, InputPreview: "secret"}}, "embedding.request input_length=12"},
		{falkenvector.Event{Type: falkenvector.EventEmbeddingResponse, Embedding: &falkenvector.EmbeddingEvent{Model: "m", Dimensions: 3}}, "embedding.response model=m dimensions=3"},
		{falkenvector.Event{Type: falkenvector.EventEmbeddingFailed, Error: "bad"}, "embedding.failed bad"},
		{falkenvector.Event{Type: falkenvector.EventLLMRequest, LLM: &falkenvector.LLMEvent{Model: "chat", Messages: []falkenvector.LLMMessage{{Content: "secret"}}}}, "llm.request model=chat"},
		{falkenvector.Event{Type: falkenvector.EventLLMResponse, LLM: &falkenvector.LLMEvent{Model: "chat", FinishReason: "stop", Text: "secret"}}, "llm.response model=chat finish=stop"},
		{falkenvector.Event{Type: falkenvector.EventLLMDelta, LLM: &falkenvector.LLMEvent{Label: "answer", Text: "secret"}}, "llm.delta label=answer"},
		{falkenvector.Event{Type: falkenvector.EventLLMFailed, Error: "bad"}, "llm.failed bad"},
		{falkenvector.Event{Type: falkenvector.EventToolCall, ToolCall: &falkenvector.ToolCallEvent{Name: "search_index", Arguments: "secret"}}, "tool.call search_index"},
		{falkenvector.Event{Type: falkenvector.EventToolResult, ToolResult: &falkenvector.ToolResultEvent{Name: "search_index", Content: "secret"}}, "tool.result search_index"},
		{falkenvector.Event{Type: falkenvector.EventWarning, Message: "careful"}, "warning careful"},
	}
	for _, tt := range cases {
		if got := FormatSDKEvent(tt.event); got != tt.want {
			t.Fatalf("FormatSDKEvent(%s) = %q, want %q", tt.event.Type, got, tt.want)
		}
	}
}

func TestFormatSDKEventUnknownAndEmptyAreSafe(t *testing.T) {
	if got := FormatSDKEvent(falkenvector.Event{Type: falkenvector.EventType("new.event"), Message: "ok"}); got != "new.event ok" {
		t.Fatalf("unknown = %q", got)
	}
	if got := FormatSDKEvent(falkenvector.Event{}); strings.Contains(got, "secret") {
		t.Fatalf("empty event leaked content: %q", got)
	}
}
