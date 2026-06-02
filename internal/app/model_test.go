package app

import (
	"testing"

	"github.com/smasonuk/falken-vector/pkg/falkenvector"
)

func TestNewModelDefaultsToAskAgentMode(t *testing.T) {
	model := NewModel("/repo")
	if model.ActiveScreen != ScreenAsk {
		t.Fatalf("ActiveScreen = %q, want %q", model.ActiveScreen, ScreenAsk)
	}
	if !model.Ask.Agent {
		t.Fatal("Ask.Agent = false, want true")
	}
	if model.Search.Mode != falkenvector.RetrievalHybrid {
		t.Fatalf("Search.Mode = %q, want %q", model.Search.Mode, falkenvector.RetrievalHybrid)
	}
	if model.Ask.Mode != falkenvector.RetrievalHybrid {
		t.Fatalf("Ask.Mode = %q, want %q", model.Ask.Mode, falkenvector.RetrievalHybrid)
	}
}
