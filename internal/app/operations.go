package app

import (
	"context"

	tuiconfig "github.com/smasonuk/falken-vector-tui/internal/config"
	"github.com/smasonuk/falken-vector/pkg/falkenvector"
)

type EngineFactory func(falkenvector.EventSink) (*falkenvector.Engine, error)

func EngineFactoryFromConfig(cfg tuiconfig.Config) EngineFactory {
	return func(sink falkenvector.EventSink) (*falkenvector.Engine, error) {
		return NewEngineFromEnvironmentWithConfig(cfg, sink)
	}
}

func StartStatus(runner *OperationRunner, factory EngineFactory, directory string) error {
	return runner.Start(OperationStatus, func(ctx context.Context, sink falkenvector.EventSink) OperationResult {
		engine, err := factory(sink)
		if err != nil {
			return OperationResult{Err: err}
		}
		defer engine.Close()
		status, err := engine.Status(ctx)
		if IsNoIndexError(err) {
			vm := NoIndexStatus(directory)
			return OperationResult{Status: &vm}
		}
		if err != nil {
			return OperationResult{Err: err}
		}
		vm := StatusViewFromSDK(directory, status)
		return OperationResult{Status: &vm}
	})
}

func StartIngest(runner *OperationRunner, factory EngineFactory, model IndexViewModel, dryRun bool) error {
	request, err := BuildIngestRequest(model, dryRun)
	if err != nil {
		return err
	}
	return runner.Start(OperationIngest, func(ctx context.Context, sink falkenvector.EventSink) OperationResult {
		engine, err := factory(sink)
		if err != nil {
			return OperationResult{Err: err}
		}
		defer engine.Close()
		result, err := engine.Ingest(ctx, request)
		return OperationResult{Ingest: &result, RefreshStatus: err == nil && !dryRun, Err: err}
	})
}

func StartSearch(runner *OperationRunner, factory EngineFactory, model SearchViewModel) error {
	request, err := BuildQueryRequest(model)
	if err != nil {
		return err
	}
	return runner.Start(OperationSearch, func(ctx context.Context, sink falkenvector.EventSink) OperationResult {
		engine, err := factory(sink)
		if err != nil {
			return OperationResult{Err: err}
		}
		defer engine.Close()
		result, err := engine.Query(ctx, request)
		if IsNoIndexError(err) {
			return OperationResult{Err: err}
		}
		if err != nil {
			return OperationResult{Err: err}
		}
		vm := model
		converted := SearchViewFromResult(result)
		vm.Results = converted.Results
		vm.QueryPlan = converted.QueryPlan
		vm.Selected = 0
		return OperationResult{Search: &vm}
	})
}

func StartAsk(runner *OperationRunner, factory EngineFactory, directory string, model AskViewModel) error {
	request, err := BuildAskRequest(model)
	if err != nil {
		return err
	}
	return runner.Start(OperationAsk, func(ctx context.Context, sink falkenvector.EventSink) OperationResult {
		engine, err := factory(sink)
		if err != nil {
			return OperationResult{Err: err}
		}
		defer engine.Close()
		answer, err := engine.Ask(ctx, request)
		if err != nil {
			return OperationResult{Err: err}
		}
		vm := model
		converted := AskViewFromAnswer(answer, directory)
		vm.Answer = converted.Answer
		vm.Sources = converted.Sources
		vm.SourceItems = converted.SourceItems
		vm.SelectedSource = 0
		vm.SourceDialogOpen = false
		vm.SourceDialogTitle = ""
		vm.SourceDialogText = ""
		vm.SourceDialogError = ""
		vm.Warnings = converted.Warnings
		vm.ToolCalls = converted.ToolCalls
		return OperationResult{Ask: &vm}
	})
}

func StartCompact(runner *OperationRunner, factory EngineFactory, model CompactViewModel, dryRun bool) error {
	request, err := BuildCompactRequest(model, dryRun)
	if err != nil {
		return err
	}
	return runner.Start(OperationCompact, func(ctx context.Context, sink falkenvector.EventSink) OperationResult {
		engine, err := factory(sink)
		if err != nil {
			return OperationResult{Err: err}
		}
		defer engine.Close()
		result, err := engine.Compact(ctx, request)
		return OperationResult{Compact: &result, RefreshStatus: err == nil && !dryRun, Err: err}
	})
}
