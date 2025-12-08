package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bobbyz3g/go-binlog-sync/pkg/config"
	"github.com/bobbyz3g/go-binlog-sync/pkg/worker"
	"golang.org/x/sync/errgroup"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.NewConfigFromFile(*configPath)
	if err != nil {
		panic(err)
	}

	lg := config.NewLogger(cfg.Log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		lg.Info("received shutdown signal", slog.String("signal", sig.String()))
		cancel()
	}()

	server := NewServer(lg, &cfg.Server)
	w := worker.NewWorker(lg, cfg.Source)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return server.ListenAndServe(ctx)
	})
	g.Go(func() error {
		return w.Run(ctx)
	})

	if err := g.Wait(); err != nil {
		lg.Error("wait error", slog.String("err", err.Error()))
	}
}
