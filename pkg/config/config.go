package config

import (
	"io"
	"log/slog"
	"os"
	"sigs.k8s.io/yaml"
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

// ServerConfig represents server listening configuration
type ServerConfig struct {
	// Host defines the host address to listen on
	Host string `json:"host" yaml:"host"`
	// Port defines the port number to listen on
	Port int `json:"port" yaml:"port"`
}

// Config represents the complete service configuration
type Config struct {
	// Log contains logging-related configuration
	Log LogConfig `json:"log" yaml:"log"`
	// Server contains server-related configuration
	Server ServerConfig `json:"server" yaml:"server"`
}

func NewConfigFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &Config{
		Log: LogConfig{
			Level: "info",
		},
		Server: ServerConfig{
			Host: "127.0.0.1",
			Port: 8080,
		},
	}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
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
		w       io.Writer
		err     error
		handler slog.Handler
	)
	if c.FilePath != "" {
		w, err = os.OpenFile(c.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(err.Error())
		}
	} else {
		w = os.Stderr
	}

	opts := &slog.HandlerOptions{
		Level: parseLogLevel(c.Level),
	}

	if c.Format == "json" {
		handler = slog.NewJSONHandler(w, opts)
	} else {
		handler = slog.NewTextHandler(w, opts)
	}

	return slog.New(handler)
}
