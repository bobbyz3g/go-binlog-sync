package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	context2 "github.com/bobbyz3g/go-binlog-sync/pkg/context"
	"github.com/bobbyz3g/go-binlog-sync/pkg/state"
)

func (w *Worker) initStateRecorder(ctx context.Context) (*state.Recorder, error) {
	cfg := w.stateCfg
	if !cfg.Enabled {
		return nil, nil
	}

	store, err := w.buildStateStore()
	if err != nil {
		return nil, err
	}
	if mysqlStore, ok := store.(*state.MySQLStore); ok {
		if err := mysqlStore.EnsureTable(ctx); err != nil {
			return nil, err
		}
	}

	saved, err := store.Load(ctx)
	if err != nil && !errors.Is(err, state.ErrNotFound) {
		return nil, err
	}
	if saved != nil {
		if err := w.applySavedState(saved); err != nil {
			w.lg.Warn("ignore invalid state", slog.String("err", err.Error()))
			saved = nil
		} else {
			w.lg.Info("state loaded", slog.String("mode", string(saved.Mode)))
		}
	}

	base := w.baseState(saved)
	every := cfg.EveryEvents
	if every <= 0 {
		every = 1
	}
	return state.NewRecorder(store, *base, every), nil
}

func (w *Worker) buildStateStore() (state.Store, error) {
	cfg := w.stateCfg
	switch strings.ToLower(strings.TrimSpace(cfg.Type)) {
	case "", "file":
		path := strings.TrimSpace(cfg.FilePath)
		if path == "" {
			return nil, errors.New("state file path is empty")
		}
		return state.NewFileStore(path), nil
	case "mysql":
		my := cfg.MySQL
		host := strings.TrimSpace(my.Host)
		if host == "" {
			return nil, errors.New("state mysql host is empty")
		}
		port := my.Port
		if port == 0 {
			port = 3306
		}
		addr := fmt.Sprintf("%s:%d", host, port)
		sourceID := w.stateSourceID()
		return state.NewMySQLStore(state.MySQLConfig{
			Addr:     addr,
			User:     my.User,
			Password: my.Password,
			Database: my.Database,
			Table:    my.Table,
			SourceID: sourceID,
		})
	default:
		return nil, fmt.Errorf("invalid state type %q", cfg.Type)
	}
}

func (w *Worker) applySavedState(saved *state.State) error {
	if saved == nil {
		return nil
	}
	mode := state.Mode(strings.ToLower(string(saved.Mode)))
	switch mode {
	case state.ModeGTID:
		if strings.TrimSpace(saved.GTIDSet) == "" {
			return errors.New("state gtid set is empty")
		}
		w.source.GTIDEnabled = true
		w.source.GTIDSet = saved.GTIDSet
	case state.ModePos:
		if strings.TrimSpace(saved.BinlogFile) == "" || saved.BinlogPos == 0 {
			return errors.New("state binlog position is empty")
		}
		if saved.BinlogPos > uint64(^uint32(0)) {
			return fmt.Errorf("state binlog position overflow: %d", saved.BinlogPos)
		}
		w.source.GTIDEnabled = false
		w.source.Binlog = saved.BinlogFile
		w.source.Position = uint32(saved.BinlogPos)
	default:
		return fmt.Errorf("invalid state mode %q", saved.Mode)
	}
	return nil
}

func (w *Worker) baseState(saved *state.State) *state.State {
	sourceID := w.stateSourceID()
	flavor := normalizeFlavor(w.source.Flavor)
	if saved != nil {
		base := *saved
		if strings.TrimSpace(base.SourceID) == "" {
			base.SourceID = sourceID
		}
		if strings.TrimSpace(base.Flavor) == "" {
			base.Flavor = flavor
		}
		if base.ServerID == 0 {
			base.ServerID = w.source.ServerID
		}
		if base.Mode == "" {
			base.Mode = modeFromSource(w.source)
		}
		if base.Mode == state.ModeGTID && base.GTIDSet == "" {
			base.GTIDSet = w.source.GTIDSet
		}
		if base.Mode == state.ModePos && base.BinlogFile == "" {
			base.BinlogFile = w.source.Binlog
		}
		if base.Mode == state.ModePos && base.BinlogPos == 0 {
			base.BinlogPos = uint64(w.source.Position)
		}
		return &base
	}
	return &state.State{
		SourceID:   sourceID,
		Flavor:     flavor,
		Mode:       modeFromSource(w.source),
		GTIDSet:    w.source.GTIDSet,
		BinlogFile: w.source.Binlog,
		BinlogPos:  uint64(w.source.Position),
		ServerID:   w.source.ServerID,
	}
}

func (w *Worker) stateSourceID() string {
	if strings.TrimSpace(w.stateCfg.MySQL.SourceID) != "" {
		return strings.TrimSpace(w.stateCfg.MySQL.SourceID)
	}
	return state.SourceKey(w.source.Host, w.source.Port, w.source.ServerID)
}

func modeFromSource(source context2.SourceConfig) state.Mode {
	if source.GTIDEnabled {
		return state.ModeGTID
	}
	return state.ModePos
}

func normalizeFlavor(flavor string) string {
	switch strings.ToLower(strings.TrimSpace(flavor)) {
	case "", "mysql":
		return "mysql"
	case "mariadb":
		return "mariadb"
	default:
		return strings.ToLower(strings.TrimSpace(flavor))
	}
}
