package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pkgctx "github.com/bobbyz3g/go-binlog-sync/pkg/context"
)

func TestRegisterHandlerMetrics(t *testing.T) {
	server := NewServer(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		&pkgctx.ServerConfig{Host: "127.0.0.1", Port: 8081},
	)
	mux := http.NewServeMux()
	server.registerHandler(mux)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	resp := httptest.NewRecorder()

	mux.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}
	if !strings.Contains(resp.Header().Get("Content-Type"), "text/plain") {
		t.Fatalf("expected text/plain content type, got %q", resp.Header().Get("Content-Type"))
	}
	if !strings.Contains(resp.Body.String(), "gbs_worker_up") {
		t.Fatalf("metrics response does not contain gbs_worker_up")
	}
}
