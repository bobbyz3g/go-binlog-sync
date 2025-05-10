package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bobbyz3g/go-binlog-backup/pkg/config"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.NewConfigFromFile(*configPath)
	if err != nil {
		panic(err)
	}

	lg := config.NewLogger(cfg.Log)
	lg.Info("Server started",
		slog.String("host", cfg.Server.Host),
		slog.Int("port", cfg.Server.Port))

	err = http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lg.Info("Received request", slog.String("method", r.Method), slog.String("url", r.URL.String()))
		w.WriteHeader(http.StatusOK)
	}))
	if err != nil {
		lg.Error("Listen and server failed", slog.String("error", err.Error()))
	}
}
