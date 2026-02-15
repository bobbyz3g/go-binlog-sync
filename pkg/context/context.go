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

// FilterList defines table/database filters for sync.
type FilterList struct {
	Databases []string `json:"databases" yaml:"databases"`
	Tables    []string `json:"tables" yaml:"tables"`
}

// FilterConfig defines whitelist/blacklist settings for sync.
type FilterConfig struct {
	Whitelist FilterList `json:"whitelist" yaml:"whitelist"`
	Blacklist FilterList `json:"blacklist" yaml:"blacklist"`
}

// StateMySQLConfig defines MySQL table storage for sync state.
type StateMySQLConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     uint16 `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	Database string `json:"database" yaml:"database"`
	Table    string `json:"table" yaml:"table"`
	SourceID string `json:"sourceID" yaml:"sourceID"`
}

// StateConfig defines state persistence configuration.
type StateConfig struct {
	// Enabled toggles state persistence.
	Enabled bool `json:"enabled" yaml:"enabled"`
	// Type selects the state backend (file/mysql).
	Type string `json:"type" yaml:"type"`
	// EveryEvents controls how many events between checkpoints.
	EveryEvents int `json:"everyEvents" yaml:"everyEvents"`
	// FilePath is the state file path when Type=file.
	FilePath string `json:"filePath" yaml:"filePath"`
	// MySQL defines the state table connection when Type=mysql.
	MySQL StateMySQLConfig `json:"mysql" yaml:"mysql"`
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
	// Filter contains table/database filtering configuration.
	Filter FilterConfig `json:"filter" yaml:"filter"`
	// State contains sync state persistence configuration
	State StateConfig `json:"state" yaml:"state"`
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
		State: StateConfig{
			Enabled:     false,
			Type:        "file",
			EveryEvents: 100,
			FilePath:    "gbs.state.json",
			MySQL: StateMySQLConfig{
				Port:  3306,
				Table: "gbs_sync_state",
			},
		},
	}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}
