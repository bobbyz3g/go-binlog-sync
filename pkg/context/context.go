package context

import (
	"log/slog"
	"os"
	"sync"

	"github.com/bobbyz3g/go-binlog-sync/pkg/logger"
	"sigs.k8s.io/yaml"
)

var (
	globalContext *SyncContext
	once          sync.Once
)

// SyncContext has the general, global state of migration. It is used by
// all components throughout the sync process.
type SyncContext struct {
	Log *slog.Logger
}

// InitContext initializes the global SyncContext.
// It should be called only once at the beginning of the application.
func InitContext(cfg *Config) {
	once.Do(func() {
		globalContext = &SyncContext{
			Log: logger.NewLogger(cfg.Log),
		}
	})
}

// Context returns the global SyncContext.
// It assumes InitContext has been called.
func Context() *SyncContext {
	return globalContext
}

// ServerConfig represents server listening configuration
type ServerConfig struct {
	// Host defines the host address to listen on
	Host string `json:"host" yaml:"host"`
	// Port defines the port number to listen on
	Port int `json:"port" yaml:"port"`
}

// SourceConfig defines configuration settings for a data source, including options required for initialization and connection.
type SourceConfig struct {
	// Flavor is "mysql" or "mariadb", if not set, use "mysql" default.
	Flavor      string `json:"flavor" yaml:"flavor"`
	ServerID    uint32 `json:"serverID" yaml:"serverID"`
	Host        string `json:"host" yaml:"host"`
	Port        uint16 `json:"port" yaml:"port"`
	User        string `json:"user" yaml:"user"`
	Password    string `json:"password" yaml:"password"`
	Binlog      string `json:"binlog" yaml:"binlog"`
	Position    uint32 `json:"position" yaml:"position"`
	GTIDSet     string `json:"gtidSet" yaml:"gtidSet"`
	GTIDEnabled bool   `json:"gtidEnabled" yaml:"gtidEnabled"`
}

// DestinationConfig defines configuration settings for a destination database connection.
type DestinationConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     uint16 `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
}

// Config represents the complete service configuration
type Config struct {
	// Log contains logging-related configuration
	Log logger.LogConfig `json:"log" yaml:"log"`
	// Server contains server-related configuration
	Server ServerConfig `json:"server" yaml:"server"`
	Source SourceConfig `json:"source" yaml:"source"`
	// Destination contains destination database connection configuration
	Destination DestinationConfig `json:"destination" yaml:"destination"`
}

func NewConfigFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &Config{
		Log: logger.LogConfig{
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
