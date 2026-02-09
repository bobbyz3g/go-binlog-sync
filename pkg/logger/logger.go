package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// LogConfig represents logging configuration
type LogConfig struct {
	// Level defines the logging level (e.g., debug, info, warn, error)
	Level string `json:"level" yaml:"level"`
	// FilePath defines the path where log files will be stored
	FilePath string `json:"filePath" yaml:"filePath"`
	// Format defines the log format (e.g., json, text)
	Format string `json:"format" yaml:"format"`
}

// parseLogLevel converts string level to slog.Level
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// NewLogger creates a new slog.Logger based on LogConfig
func NewLogger(c LogConfig) *slog.Logger {
	var (
		w   io.Writer
		err error
	)
	if c.FilePath != "" {
		w, err = os.OpenFile(c.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(err.Error())
		}
	} else {
		w = os.Stdout
	}

	opts := &slog.HandlerOptions{
		Level:     parseLogLevel(c.Level),
		AddSource: true,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key != slog.SourceKey {
				return a
			}
			src, ok := a.Value.Any().(*slog.Source)
			if !ok || src == nil {
				return a
			}
			short := *src
			short.File = filepath.Base(short.File)
			a.Value = slog.AnyValue(&short)
			return a
		},
	}

	var handler slog.Handler
	if c.Format == "json" {
		handler = slog.NewJSONHandler(w, opts)
	} else {
		handler = slog.NewTextHandler(w, opts)
	}

	return slog.New(handler)
}
