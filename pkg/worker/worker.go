package worker

import (
	"context"
	"fmt"
	"log/slog"

	context2 "github.com/bobbyz3g/go-binlog-sync/pkg/context"
)

type Worker struct {
	lg     *slog.Logger
	source context2.SourceConfig
	dest   context2.DestinationConfig
}

func NewWorker(lg *slog.Logger, source context2.SourceConfig, dest context2.DestinationConfig) *Worker {
	return &Worker{
		source: source,
		dest:   dest,
		lg:     lg,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	w.lg.Info("worker started")
	if err := w.precheckSource(ctx); err != nil {
		return err
	}

	reader := NewBinlogReader(w.lg, w.source)
	events, err := reader.Read(ctx)
	if err != nil {
		return fmt.Errorf("read binlog: %w", err)
	}

	writer := NewEventWriter(w.lg, w.dest)
	if err := writer.Write(ctx, events); err != nil {
		return fmt.Errorf("write events: %w", err)
	}
	w.lg.Info("worker stopped")

	return nil
}
