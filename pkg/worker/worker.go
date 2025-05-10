package worker

import (
	"context"
	"log/slog"

	"github.com/bobbyz3g/go-binlog-backup/pkg/config"
)

type Worker struct {
	lg  *slog.Logger
	cfg config.SourceConfig
}

func (w *Worker) Run(ctx context.Context) error {
	w.lg.Debug("Worker started")
	return nil
}
