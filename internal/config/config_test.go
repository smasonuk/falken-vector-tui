package config

import (
	"testing"

	"github.com/smasonuk/falken-vector/pkg/falkenvector"
)

func TestParseArgs(t *testing.T) {
	cfg, err := ParseArgs([]string{"--state-dir", "/tmp/state", "--retrieval", "hybrid", "--top-k", "12"})
	if err != nil {
		t.Fatalf("ParseArgs: %v", err)
	}
	if cfg.StateDir != "/tmp/state" || cfg.Retrieval != falkenvector.RetrievalHybrid || cfg.TopK != 12 {
		t.Fatalf("cfg = %+v", cfg)
	}
}

func TestParseArgsRejectsInvalidRetrieval(t *testing.T) {
	if _, err := ParseArgs([]string{"--retrieval", "potato"}); err == nil {
		t.Fatal("ParseArgs succeeded, want error")
	}
}

func TestParseArgsTopKZeroUsesSDKDefault(t *testing.T) {
	cfg, err := ParseArgs([]string{"--top-k", "0"})
	if err != nil {
		t.Fatalf("ParseArgs: %v", err)
	}
	engine := ApplyToEngineConfig(cfg, falkenvector.EngineConfig{})
	if engine.Retrieval.TopK != 0 {
		t.Fatalf("TopK = %d, want SDK default zero", engine.Retrieval.TopK)
	}
}
