package config

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/smasonuk/falken-vector/pkg/falkenvector"
)

type Config struct {
	StateDir  string
	Retrieval falkenvector.RetrievalMode
	TopK      int
}

func ParseArgs(args []string) (Config, error) {
	var cfg Config
	fs := flag.NewFlagSet("falken-vector-tui", flag.ContinueOnError)
	fs.StringVar(&cfg.StateDir, "state-dir", "", "state directory")
	retrieval := fs.String("retrieval", "", "retrieval mode")
	fs.IntVar(&cfg.TopK, "top-k", 0, "retrieval top-k")
	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}
	mode, err := ParseRetrievalMode(*retrieval)
	if err != nil {
		return Config{}, err
	}
	cfg.Retrieval = mode
	if cfg.TopK < 0 {
		return Config{}, errors.New("--top-k must be >= 0")
	}
	return cfg, nil
}

func ParseRetrievalMode(value string) (falkenvector.RetrievalMode, error) {
	switch falkenvector.RetrievalMode(strings.TrimSpace(value)) {
	case "":
		return "", nil
	case falkenvector.RetrievalLexical:
		return falkenvector.RetrievalLexical, nil
	case falkenvector.RetrievalVector:
		return falkenvector.RetrievalVector, nil
	case falkenvector.RetrievalHybrid:
		return falkenvector.RetrievalHybrid, nil
	default:
		return "", fmt.Errorf("--retrieval must be lexical, vector, or hybrid")
	}
}

func ApplyToEngineConfig(cfg Config, engine falkenvector.EngineConfig) falkenvector.EngineConfig {
	if strings.TrimSpace(cfg.StateDir) != "" {
		engine.StateDir = strings.TrimSpace(cfg.StateDir)
	}
	if cfg.Retrieval != "" {
		engine.Retrieval.Mode = cfg.Retrieval
	}
	if cfg.TopK > 0 {
		engine.Retrieval.TopK = cfg.TopK
	}
	return engine
}
