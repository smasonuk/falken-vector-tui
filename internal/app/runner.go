package app

import (
	"context"
	"errors"
	"sync"

	"github.com/smasonuk/falken-vector/pkg/falkenvector"
)

var ErrOperationBusy = errors.New("operation already running")

type OperationResult struct {
	Kind OperationKind

	Status  *StatusViewModel
	Ingest  *falkenvector.IngestResult
	Search  *SearchViewModel
	Ask     *AskViewModel
	Compact *falkenvector.CompactResult

	RefreshStatus bool
	Err           error
}

type OperationRunner struct {
	mu      sync.Mutex
	busy    bool
	kind    OperationKind
	cancel  context.CancelFunc
	results chan OperationResult
	events  chan falkenvector.Event
}

func NewOperationRunner() *OperationRunner {
	return &OperationRunner{
		results: make(chan OperationResult, 16),
		events:  make(chan falkenvector.Event, 256),
	}
}

func (r *OperationRunner) Results() <-chan OperationResult {
	if r == nil {
		return nil
	}
	return r.results
}

func (r *OperationRunner) Events() <-chan falkenvector.Event {
	if r == nil {
		return nil
	}
	return r.events
}

func (r *OperationRunner) Busy() (bool, OperationKind) {
	if r == nil {
		return false, OperationNone
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.busy, r.kind
}

func (r *OperationRunner) Start(kind OperationKind, fn func(context.Context, falkenvector.EventSink) OperationResult) error {
	if r == nil {
		return errors.New("operation runner is nil")
	}
	r.mu.Lock()
	if r.busy {
		r.mu.Unlock()
		return ErrOperationBusy
	}
	ctx, cancel := context.WithCancel(context.Background())
	r.busy = true
	r.kind = kind
	r.cancel = cancel
	r.mu.Unlock()

	go func() {
		result := fn(ctx, r.emitEvent)
		if result.Kind == OperationNone {
			result.Kind = kind
		}
		r.mu.Lock()
		r.busy = false
		r.kind = OperationNone
		r.cancel = nil
		r.mu.Unlock()
		r.results <- result
	}()
	return nil
}

func (r *OperationRunner) Cancel() bool {
	if r == nil {
		return false
	}
	r.mu.Lock()
	cancel := r.cancel
	r.mu.Unlock()
	if cancel == nil {
		return false
	}
	cancel()
	return true
}

func (r *OperationRunner) emitEvent(event falkenvector.Event) {
	select {
	case r.events <- event:
	default:
	}
}
