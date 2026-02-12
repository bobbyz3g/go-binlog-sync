package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	pkgctx "github.com/bobbyz3g/go-binlog-sync/pkg/context"
	"github.com/bobbyz3g/go-binlog-sync/pkg/sql"
	"github.com/bobbyz3g/go-binlog-sync/pkg/state"
	"github.com/go-mysql-org/go-mysql/client"
	"github.com/go-mysql-org/go-mysql/replication"
)

type EventWriter struct {
	lg          *slog.Logger
	cfg         pkgctx.DestinationConfig
	conn        *client.Conn
	columnCache map[string][]string
	recorder    *state.Recorder
}

func NewEventWriter(lg *slog.Logger, cfg pkgctx.DestinationConfig, recorder *state.Recorder) *EventWriter {
	return &EventWriter{
		lg:          lg,
		cfg:         cfg,
		columnCache: make(map[string][]string),
		recorder:    recorder,
	}
}

func (w *EventWriter) Write(ctx context.Context, events <-chan *StreamEvent) error {
	if err := w.connect(ctx); err != nil {
		return err
	}
	defer w.conn.Close()
	defer w.flushState()

	for {
		select {
		case <-ctx.Done():
			return nil
		case e, ok := <-events:
			if !ok {
				return nil
			}
			if e == nil || e.Event == nil {
				continue
			}
			if err := w.applyEvent(ctx, e.Event); err != nil {
				return err
			}
			if err := w.recordState(ctx, e); err != nil {
				return err
			}
		}
	}
}

func (w *EventWriter) connect(ctx context.Context) error {
	if w.conn != nil {
		return nil
	}
	if strings.TrimSpace(w.cfg.Host) == "" {
		return errors.New("destination host is empty")
	}
	port := w.cfg.Port
	if port == 0 {
		port = 3306
	}
	addr := fmt.Sprintf("%s:%d", w.cfg.Host, port)
	conn, err := client.ConnectWithContext(ctx, addr, w.cfg.User, w.cfg.Password, "", 10*time.Second)
	if err != nil {
		return fmt.Errorf("connect destination: %w", err)
	}
	w.conn = conn
	return nil
}

func (w *EventWriter) applyEvent(ctx context.Context, e *replication.BinlogEvent) error {
	switch ev := e.Event.(type) {
	case *replication.RowsEvent:
		return w.applyRowsEvent(ctx, ev)
	case *replication.QueryEvent:
		return w.applyQueryEvent(ctx, ev)
	default:
		return nil
	}
}

func (w *EventWriter) applyQueryEvent(ctx context.Context, ev *replication.QueryEvent) error {
	query := strings.TrimSpace(string(ev.Query))
	if query == "" {
		return nil
	}
	switch strings.ToLower(query) {
	case "begin", "commit", "rollback":
		return nil
	}
	if len(ev.Schema) > 0 && !sql.IsCreateDatabase(ev.Query) {
		if err := w.conn.UseDB(string(ev.Schema)); err != nil {
			return fmt.Errorf("use schema %s(%s): %w", string(ev.Schema), ev.Query, err)
		}
	}
	return w.exec(ctx, query)
}

func (w *EventWriter) applyRowsEvent(ctx context.Context, ev *replication.RowsEvent) error {
	if ev.Table == nil {
		return errors.New("rows event missing table map")
	}
	schema := string(ev.Table.Schema)
	table := string(ev.Table.Table)
	if schema == "" || table == "" {
		return fmt.Errorf("rows event missing schema/table (schema=%q table=%q)", schema, table)
	}
	columns, err := w.columnsForTable(schema, table, int(ev.ColumnCount), ev.Table)
	if err != nil {
		return err
	}
	qualifiedTable := sql.QuoteTable(schema, table)

	switch ev.Type() {
	case replication.EnumRowsEventTypeInsert:
		for i, row := range ev.Rows {
			cols, vals, err := sql.PickColumns(columns, row, sql.SkipColumns(ev, i))
			if err != nil {
				return err
			}
			stmt, args, err := sql.BuildInsertStatement(qualifiedTable, cols, vals)
			if err != nil {
				return err
			}
			if err := w.exec(ctx, stmt, args...); err != nil {
				return err
			}
		}
	case replication.EnumRowsEventTypeDelete:
		for i, row := range ev.Rows {
			cols, vals, err := sql.PickColumns(columns, row, sql.SkipColumns(ev, i))
			if err != nil {
				return err
			}
			stmt, args, err := sql.BuildDeleteStatement(qualifiedTable, cols, vals)
			if err != nil {
				return err
			}
			if err := w.exec(ctx, stmt, args...); err != nil {
				return err
			}
		}
	case replication.EnumRowsEventTypeUpdate:
		if len(ev.Rows)%2 != 0 {
			return fmt.Errorf("update rows event has odd row count: %d", len(ev.Rows))
		}
		for i := 0; i+1 < len(ev.Rows); i += 2 {
			before := ev.Rows[i]
			after := ev.Rows[i+1]
			beforeCols, beforeVals, err := sql.PickColumns(columns, before, sql.SkipColumns(ev, i))
			if err != nil {
				return err
			}
			afterCols, afterVals, err := sql.PickColumns(columns, after, sql.SkipColumns(ev, i+1))
			if err != nil {
				return err
			}
			stmt, args, err := sql.BuildUpdateStatement(qualifiedTable, afterCols, afterVals, beforeCols, beforeVals)
			if err != nil {
				return err
			}
			if err := w.exec(ctx, stmt, args...); err != nil {
				return err
			}
		}
	default:
		w.lg.Debug("unsupported rows event type", slog.String("type", ev.Type().String()))
	}

	return nil
}

func (w *EventWriter) columnsForTable(schema, table string, columnCount int, tableMap *replication.TableMapEvent) ([]string, error) {
	key := schema + "." + table
	if cols, ok := w.columnCache[key]; ok && len(cols) == columnCount {
		return cols, nil
	}
	if tableMap != nil && len(tableMap.ColumnName) == columnCount {
		cols := make([]string, 0, columnCount)
		for _, name := range tableMap.ColumnName {
			if len(name) == 0 {
				cols = nil
				break
			}
			cols = append(cols, string(name))
		}
		if len(cols) == columnCount {
			w.columnCache[key] = cols
			return cols, nil
		}
	}
	result, err := w.conn.Execute(
		"SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA=? AND TABLE_NAME=? ORDER BY ORDINAL_POSITION",
		schema,
		table,
	)
	if err != nil {
		return nil, fmt.Errorf("fetch columns for %s.%s: %w", schema, table, err)
	}
	defer result.Close()
	if result.Resultset == nil {
		return nil, fmt.Errorf("fetch columns for %s.%s: empty resultset", schema, table)
	}
	cols := make([]string, 0, result.Resultset.RowNumber())
	for i := 0; i < result.Resultset.RowNumber(); i++ {
		v, err := result.Resultset.GetValue(i, 0)
		if err != nil {
			return nil, fmt.Errorf("read columns for %s.%s: %w", schema, table, err)
		}
		switch name := v.(type) {
		case string:
			cols = append(cols, name)
		case []byte:
			cols = append(cols, string(name))
		default:
			return nil, fmt.Errorf("read columns for %s.%s: unexpected type %T", schema, table, v)
		}
	}
	if len(cols) != columnCount {
		return nil, fmt.Errorf("column count mismatch for %s.%s: got %d expected %d", schema, table, len(cols), columnCount)
	}
	w.columnCache[key] = cols
	return cols, nil
}

func (w *EventWriter) exec(ctx context.Context, query string, args ...interface{}) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	result, err := w.conn.Execute(query, args...)
	if result != nil {
		result.Close()
	}
	if err != nil {
		return fmt.Errorf("execute query: %w", err)
	}
	return nil
}

func (w *EventWriter) recordState(ctx context.Context, e *StreamEvent) error {
	if w.recorder == nil || e == nil {
		return nil
	}
	upd := state.Update{
		BinlogFile: e.Position.Name,
		BinlogPos:  uint64(e.Position.Pos),
		GTIDSet:    e.GTIDSet,
	}
	return w.recorder.Record(ctx, upd)
}

func (w *EventWriter) flushState() {
	if w.recorder == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := w.recorder.Flush(ctx); err != nil {
		w.lg.Error("flush state failed", slog.String("err", err.Error()))
	}
}
