package worker

import (
	"context"
	"fmt"
	"log/slog"

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
	syncCfg := replication.BinlogSyncerConfig{
		Flavor:   "mysql",
		ServerID: b.cfg.ServerID,
		Host:     b.cfg.Host,
		Port:     b.cfg.Port,
		User:     b.cfg.User,
		Password: b.cfg.Password,
		Logger:   b.lg,
	}
	syncer := replication.NewBinlogSyncer(syncCfg)
	// TODO: gtid and get previous position
	streamer, err := syncer.StartSync(mysql.Position{Name: b.cfg.Binlog, Pos: b.cfg.Position})
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
