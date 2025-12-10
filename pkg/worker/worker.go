package worker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

type Worker struct {
	lg  *slog.Logger
	cfg SourceConfig
}

func NewWorker(lg *slog.Logger, cfg SourceConfig) *Worker {
	return &Worker{
		cfg: cfg,
		lg:  lg,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	w.lg.Info("worker started")
	reader := NewBinlogReader(w.lg, w.cfg)
	events, err := reader.Read(ctx)
	if err != nil {
		return fmt.Errorf("read binlog: %w", err)
	}

	for e := range events {
		e.Dump(os.Stdout)
	}
	w.lg.Info("worker stopped")

	return nil
}
