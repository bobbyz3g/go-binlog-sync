package worker

import (
	"context"
	"fmt"
	"log/slog"

	context2 "github.com/bobbyz3g/go-binlog-sync/pkg/context"
	"github.com/bobbyz3g/go-binlog-sync/pkg/filter"
)

type Worker struct {
	lg       *slog.Logger
	source   context2.SourceConfig
	dest     context2.DestinationConfig
	stateCfg context2.StateConfig
	filter   context2.FilterConfig
}

func NewWorker(lg *slog.Logger, source context2.SourceConfig, dest context2.DestinationConfig, stateCfg context2.StateConfig, filter context2.FilterConfig) *Worker {
	return &Worker{
		source:   source,
		dest:     dest,
		stateCfg: stateCfg,
		filter:   filter,
		lg:       lg,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	w.lg.Info("worker started")
	recorder, err := w.initStateRecorder(ctx)
	if err != nil {
		return err
	}
	if err := w.precheckSource(ctx); err != nil {
		return err
	}

	tableFilter := filter.NewTableFilter(w.filter)
	reader := NewBinlogReader(w.lg, w.source)
	events, err := reader.Read(ctx)
	if err != nil {
		return fmt.Errorf("read binlog: %w", err)
	}

	writer := NewEventWriter(w.lg, w.dest, recorder, tableFilter)
	if err := writer.Write(ctx, events); err != nil {
		return fmt.Errorf("write events: %w", err)
	}
	w.lg.Info("worker stopped")

	return nil
}
