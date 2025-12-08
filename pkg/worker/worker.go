package worker

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/bobbyz3g/go-binlog-sync/pkg/config"
)

type Worker struct {
	lg  *slog.Logger
	cfg config.SourceConfig
}

func NewWorker(lg *slog.Logger, cfg config.SourceConfig) *Worker {
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
