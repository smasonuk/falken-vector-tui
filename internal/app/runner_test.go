package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/smasonuk/falken-vector/pkg/falkenvector"
)

func TestOperationRunnerStartCompleteAndBusy(t *testing.T) {
	runner := NewOperationRunner()
	release := make(chan struct{})
	if err := runner.Start(OperationStatus, func(context.Context, falkenvector.EventSink) OperationResult {
		<-release
		return OperationResult{}
	}); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if busy, kind := runner.Busy(); !busy || kind != OperationStatus {
		t.Fatalf("Busy = %t, %s; want status busy", busy, kind)
	}
	if err := runner.Start(OperationSearch, func(context.Context, falkenvector.EventSink) OperationResult { return OperationResult{} }); !errors.Is(err, ErrOperationBusy) {
		t.Fatalf("second Start error = %v, want ErrOperationBusy", err)
	}
	close(release)
	select {
	case result := <-runner.Results():
		if result.Kind != OperationStatus {
			t.Fatalf("result kind = %s", result.Kind)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for result")
	}
	if busy, kind := runner.Busy(); busy || kind != OperationNone {
		t.Fatalf("Busy after result = %t, %s; want idle", busy, kind)
	}
}

func TestOperationRunnerCancel(t *testing.T) {
	runner := NewOperationRunner()
	if err := runner.Start(OperationIngest, func(ctx context.Context, sink falkenvector.EventSink) OperationResult {
		sink(falkenvector.Event{Type: falkenvector.EventRunStarted, Message: "ingest"})
		<-ctx.Done()
		return OperationResult{Err: ctx.Err()}
	}); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !runner.Cancel() {
		t.Fatal("Cancel = false, want true")
	}
	select {
	case event := <-runner.Events():
		if event.Type != falkenvector.EventRunStarted {
			t.Fatalf("event = %+v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}
	select {
	case result := <-runner.Results():
		if !errors.Is(result.Err, context.Canceled) {
			t.Fatalf("result err = %v, want canceled", result.Err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for cancel result")
	}
}
