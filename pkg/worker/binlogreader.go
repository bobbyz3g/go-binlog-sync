package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	context2 "github.com/bobbyz3g/go-binlog-sync/pkg/context"
	"github.com/bobbyz3g/go-binlog-sync/pkg/metrics"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
)

type BinlogReader struct {
	lg  *slog.Logger
	cfg context2.SourceConfig
}

func NewBinlogReader(lg *slog.Logger, source context2.SourceConfig) *BinlogReader {
	return &BinlogReader{
		cfg: source,
		lg:  lg,
	}
}

func (b *BinlogReader) Read(ctx context.Context) (chan *StreamEvent, error) {
	var flavor string
	switch strings.ToLower(b.cfg.Flavor) {
	case "mysql", "":
		flavor = mysql.MySQLFlavor
	case "mariadb":
		flavor = mysql.MariaDBFlavor

	default:
		return nil, errors.New("invalid flavor")
	}

	syncCfg := replication.BinlogSyncerConfig{
		Flavor:   flavor,
		ServerID: b.cfg.ServerID,
		Host:     b.cfg.Host,
		Port:     b.cfg.Port,
		User:     b.cfg.User,
		Password: b.cfg.Password,
		Logger:   b.lg,
	}

	syncer := replication.NewBinlogSyncer(syncCfg)
	var (
		streamer *replication.BinlogStreamer
		err      error
	)

	if b.cfg.GTIDEnabled {
		gset, err := mysql.ParseGTIDSet(flavor, b.cfg.GTIDSet)
		if err != nil {
			return nil, fmt.Errorf("parse GTID set: %w", err)
		}
		streamer, err = syncer.StartSyncGTID(gset)
	} else {
		streamer, err = syncer.StartSync(mysql.Position{Name: b.cfg.Binlog, Pos: b.cfg.Position})
	}

	if err != nil {
		return nil, fmt.Errorf("start sync: %w", err)
	}

	events := make(chan *StreamEvent)

	go func() {
		for {
			select {
			case <-ctx.Done():
				syncer.Close()
				b.lg.Info("binlog reader stopped")
				close(events)
				return
			default:
			}
			e, err := streamer.GetEvent(ctx)
			if err != nil {
				metrics.IncBinlogReadErrors()
				b.lg.Error("get event failed", slog.String("err", err.Error()))
				continue
			}
			metrics.IncBinlogEventsReadTotal(binlogEventTypeLabel(e))
			if e != nil && e.Header != nil && e.Header.Timestamp > 0 {
				lag := time.Since(time.Unix(int64(e.Header.Timestamp), 0)).Seconds()
				metrics.SetReplicationLag(lag)
			}
			events <- &StreamEvent{
				Event:    e,
				Position: syncer.GetNextPosition(),
				GTIDSet:  extractGTIDSet(e),
			}
		}
	}()

	return events, nil
}

func extractGTIDSet(e *replication.BinlogEvent) string {
	if e == nil {
		return ""
	}
	switch ev := e.Event.(type) {
	case *replication.QueryEvent:
		if ev.GSet != nil {
			return ev.GSet.String()
		}
	case *replication.XIDEvent:
		if ev.GSet != nil {
			return ev.GSet.String()
		}
	}
	return ""
}

func binlogEventTypeLabel(e *replication.BinlogEvent) string {
	if e == nil || e.Header == nil {
		return "unknown"
	}
	return strings.ToLower(e.Header.EventType.String())
}
