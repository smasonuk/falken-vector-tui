package app

import (
	"os"

	tuiconfig "github.com/smasonuk/falken-vector-tui/internal/config"
	"github.com/smasonuk/falken-vector/pkg/falkenvector"
)

func NewEngineFromEnvironment(eventSink falkenvector.EventSink) (*falkenvector.Engine, error) {
	return NewEngineFromEnvironmentWithConfig(tuiconfig.Config{}, eventSink)
}

func NewEngineFromEnvironmentWithConfig(cfg tuiconfig.Config, eventSink falkenvector.EventSink) (*falkenvector.Engine, error) {
	engineConfig, err := falkenvector.EngineConfigFromEnvE(os.Getenv)
	if err != nil {
		return nil, err
	}
	engineConfig = tuiconfig.ApplyToEngineConfig(cfg, engineConfig)
	engineConfig.Events = eventSink
	return falkenvector.NewEngine(engineConfig)
}
