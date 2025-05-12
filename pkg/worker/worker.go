package worker

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/bobbyz3g/go-binlog-backup/pkg/config"
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
	w.lg.Info("Worker started")
	reader := NewBinlogReader(w.lg, w.cfg)
	events, err := reader.Read(ctx)
	if err != nil {
		return fmt.Errorf("read binlog: %w", err)
	}

reading:
	for {
		select {
		case <-ctx.Done():
			w.lg.Info("Worker stopped, err: %w", ctx.Err())
			break reading
		case e := <-events:
			e.Dump(os.Stdout)
		}
	}

	return nil
}
