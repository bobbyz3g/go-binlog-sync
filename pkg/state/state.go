package state

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Mode represents the checkpoint mode.
type Mode string

const (
	ModeGTID Mode = "gtid"
	ModePos  Mode = "pos"
)

var ErrNotFound = errors.New("state not found")

// State records the last completed binlog position or GTID set.
type State struct {
	SourceID   string    `json:"source_id" yaml:"source_id"`
	Flavor     string    `json:"flavor" yaml:"flavor"`
	Mode       Mode      `json:"mode" yaml:"mode"`
	GTIDSet    string    `json:"gtid_set,omitempty" yaml:"gtid_set,omitempty"`
	BinlogFile string    `json:"binlog_file,omitempty" yaml:"binlog_file,omitempty"`
	BinlogPos  uint64    `json:"binlog_pos,omitempty" yaml:"binlog_pos,omitempty"`
	ServerID   uint32    `json:"server_id,omitempty" yaml:"server_id,omitempty"`
	UpdatedAt  time.Time `json:"updated_at" yaml:"updated_at"`
	Version    uint64    `json:"version" yaml:"version"`
}

// Store persists and loads State.
type Store interface {
	Load(ctx context.Context) (*State, error)
	Save(ctx context.Context, st *State) error
}

// Touch updates UpdatedAt and increments Version.
func (s *State) Touch(now time.Time) {
	if s == nil {
		return
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	s.UpdatedAt = now
	if s.Version == 0 {
		s.Version = 1
	} else {
		s.Version++
	}
}

// SourceKey returns a default source identifier.
func SourceKey(host string, port uint16, serverID uint32) string {
	host = strings.TrimSpace(host)
	if host == "" {
		host = "unknown"
	}
	return fmt.Sprintf("%s:%d#%d", host, port, serverID)
}

func timeNowUTC() time.Time {
	return time.Now().UTC()
}
