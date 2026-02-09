package worker

import (
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
)

// StreamEvent wraps a binlog event with checkpoint metadata.
type StreamEvent struct {
	Event    *replication.BinlogEvent
	Position mysql.Position
	GTIDSet  string
}
