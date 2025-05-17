package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/bobbyz3g/go-binlog-backup/pkg/config"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
)

type BinlogReader struct {
	lg  *slog.Logger
	cfg config.SourceConfig
}

func NewBinlogReader(lg *slog.Logger, source config.SourceConfig) *BinlogReader {
	return &BinlogReader{
		cfg: source,
		lg:  lg,
	}
}

func (b *BinlogReader) Read(ctx context.Context) (chan *replication.BinlogEvent, error) {
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

	events := make(chan *replication.BinlogEvent)

	go func() {
		for {
			select {
			case <-ctx.Done():
				syncer.Close()
				b.lg.Info("BinlogReader stopped")
				return
			default:
			}
			e, err := streamer.GetEvent(ctx)
			if err != nil {
				b.lg.Error("BinlogReader GetEvent failed", slog.String("err", err.Error()))
				continue
			}
			events <- e
		}
	}()

	return events, nil
}
