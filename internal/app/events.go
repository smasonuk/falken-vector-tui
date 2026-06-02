package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/smasonuk/falken-vector/pkg/falkenvector"
)

func FormatSDKEvent(event falkenvector.Event) string {
	switch event.Type {
	case falkenvector.EventRunStarted:
		return joinEvent("run.started", event.Message)
	case falkenvector.EventRunCompleted:
		return joinEvent("run.completed", event.Message)
	case falkenvector.EventRunFailed:
		return joinEvent("run.failed", firstNonEmpty(event.Error, event.Message))
	case falkenvector.EventRetrievalStarted:
		if event.Retrieval == nil {
			return "retrieval.started"
		}
		return fmt.Sprintf("retrieval.started mode=%s top_k=%d", event.Retrieval.Mode, event.Retrieval.TopK)
	case falkenvector.EventRetrievalCompleted:
		if event.Retrieval == nil {
			return "retrieval.completed"
		}
		return fmt.Sprintf("retrieval.completed results=%d", event.Retrieval.ResultCount)
	case falkenvector.EventRetrievalFailed:
		return joinEvent("retrieval.failed", firstNonEmpty(event.Error, event.Message))
	case falkenvector.EventIngestStarted:
		if event.Ingest == nil {
			return "ingest.started"
		}
		return joinEvent("ingest.started", event.Ingest.Directory)
	case falkenvector.EventIngestProgress:
		if event.Ingest == nil {
			return "ingest.progress"
		}
		return formatIngestProgress(*event.Ingest)
	case falkenvector.EventIngestCompleted:
		if event.Ingest == nil {
			return "ingest.completed"
		}
		return fmt.Sprintf("ingest.completed scanned=%d changed=%d new=%d failed=%d", event.Ingest.Scanned, event.Ingest.ChangedFiles, event.Ingest.NewFiles, event.Ingest.FailedFiles)
	case falkenvector.EventEmbeddingRequest:
		if event.Embedding == nil {
			return "embedding.request"
		}
		return fmt.Sprintf("embedding.request input_length=%d", event.Embedding.InputLength)
	case falkenvector.EventEmbeddingResponse:
		if event.Embedding == nil {
			return "embedding.response"
		}
		return fmt.Sprintf("embedding.response model=%s dimensions=%d", event.Embedding.Model, event.Embedding.Dimensions)
	case falkenvector.EventEmbeddingFailed:
		return joinEvent("embedding.failed", firstNonEmpty(event.Error, event.Message))
	case falkenvector.EventLLMRequest:
		if event.LLM == nil {
			return "llm.request"
		}
		return fmt.Sprintf("llm.request model=%s", event.LLM.Model)
	case falkenvector.EventLLMResponse:
		if event.LLM == nil {
			return "llm.response"
		}
		return fmt.Sprintf("llm.response model=%s finish=%s", event.LLM.Model, event.LLM.FinishReason)
	case falkenvector.EventLLMDelta:
		if event.LLM == nil || event.LLM.Label == "" {
			return "llm.delta"
		}
		return fmt.Sprintf("llm.delta label=%s", event.LLM.Label)
	case falkenvector.EventLLMFailed:
		return joinEvent("llm.failed", firstNonEmpty(event.Error, event.Message))
	case falkenvector.EventToolCall:
		if event.ToolCall == nil {
			return "tool.call"
		}
		return joinEvent("tool.call", event.ToolCall.Name)
	case falkenvector.EventToolResult:
		if event.ToolResult == nil {
			return "tool.result"
		}
		if event.ToolResult.Error != "" {
			return fmt.Sprintf("tool.result %s error=%s", event.ToolResult.Name, event.ToolResult.Error)
		}
		return joinEvent("tool.result", event.ToolResult.Name)
	case falkenvector.EventWarning:
		return joinEvent("warning", event.Message)
	default:
		return joinEvent(string(event.Type), firstNonEmpty(event.Message, event.Error))
	}
}

func EventLineFromSDK(event falkenvector.Event) EventLine {
	at := event.At
	if at.IsZero() {
		at = time.Now()
	}
	return EventLine{At: at, Text: FormatSDKEvent(event)}
}

func formatIngestProgress(event falkenvector.IngestEvent) string {
	var parts []string
	parts = append(parts, "ingest.progress")
	if event.Action != "" {
		parts = append(parts, event.Action)
	}
	if event.CurrentPath != "" {
		parts = append(parts, event.CurrentPath)
	}
	if event.CurrentFile != 0 || event.TotalFiles != 0 {
		parts = append(parts, fmt.Sprintf("%d/%d", event.CurrentFile, event.TotalFiles))
	}
	if event.Scanned != 0 {
		parts = append(parts, fmt.Sprintf("scanned=%d", event.Scanned))
	}
	return strings.Join(parts, " ")
}

func joinEvent(prefix string, detail string) string {
	detail = strings.TrimSpace(detail)
	if detail == "" {
		return prefix
	}
	return prefix + " " + detail
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
