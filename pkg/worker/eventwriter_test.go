package worker

import (
	"context"
	"errors"
	"testing"

	"github.com/go-mysql-org/go-mysql/replication"
)

func TestDrainEventsProcessesQueuedEventAfterCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	events := make(chan *StreamEvent, 1)
	events <- &StreamEvent{Event: &replication.BinlogEvent{}}
	close(events)

	called := false
	err := drainEvents(ctx, events, func(processCtx context.Context, e *StreamEvent) error {
		called = true
		if processCtx.Err() != nil {
			t.Fatalf("process context canceled: %v", processCtx.Err())
		}
		if e == nil || e.Event == nil {
			t.Fatal("expected non-nil event")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("drain events: %v", err)
	}
	if !called {
		t.Fatal("expected queued event to be processed")
	}
}

func TestDrainEventsSkipsNilEvents(t *testing.T) {
	events := make(chan *StreamEvent, 3)
	events <- nil
	events <- &StreamEvent{}
	events <- &StreamEvent{Event: &replication.BinlogEvent{}}
	close(events)

	calls := 0
	err := drainEvents(context.Background(), events, func(processCtx context.Context, e *StreamEvent) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("drain events: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 handled event, got %d", calls)
	}
}

func TestDrainEventsReturnsHandlerError(t *testing.T) {
	wantErr := errors.New("boom")

	events := make(chan *StreamEvent, 1)
	events <- &StreamEvent{Event: &replication.BinlogEvent{}}
	close(events)

	err := drainEvents(context.Background(), events, func(processCtx context.Context, e *StreamEvent) error {
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}
