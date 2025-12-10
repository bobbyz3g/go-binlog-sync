package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/bobbyz3g/go-binlog-sync/pkg/worker"
)

type Server struct {
	lg  *slog.Logger
	cfg *worker.ServerConfig
	mux *http.ServeMux
}

func NewServer(lg *slog.Logger, cfg *worker.ServerConfig) *Server {
	return &Server{
		lg:  lg,
		cfg: cfg,
	}
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	mux := http.NewServeMux()
	s.registerHandler(mux)
	server := &http.Server{
		Addr:    net.JoinHostPort(s.cfg.Host, strconv.Itoa(s.cfg.Port)),
		Handler: mux,
	}

	errC := make(chan error)
	go func() {
		s.lg.Info("server started", slog.String("host", s.cfg.Host), slog.Int("port", s.cfg.Port))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errC <- fmt.Errorf("listen and serve: %w", err)
		}
	}()

	select {
	case err := <-errC:
		return err
	case <-ctx.Done():
	}

	s.lg.Info("shutting down server...")

	sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(sctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}
	return nil
}

func (s *Server) registerHandler(mux *http.ServeMux) {

}
