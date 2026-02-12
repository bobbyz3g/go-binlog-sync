package state

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bobbyz3g/go-binlog-sync/pkg/sql"
	"github.com/go-mysql-org/go-mysql/client"
	"github.com/go-mysql-org/go-mysql/mysql"
)

type MySQLConfig struct {
	Addr     string
	User     string
	Password string
	Database string
	Table    string
	SourceID string
	Timeout  time.Duration
}

type MySQLStore struct {
	cfg MySQLConfig
}

func NewMySQLStore(cfg MySQLConfig) (*MySQLStore, error) {
	if strings.TrimSpace(cfg.Addr) == "" {
		return nil, errors.New("mysql addr is empty")
	}
	if strings.TrimSpace(cfg.Table) == "" {
		return nil, errors.New("mysql table is empty")
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}
	return &MySQLStore{cfg: cfg}, nil
}

func (s *MySQLStore) EnsureTable(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	conn, err := s.connect(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	table, err := qualifiedTable(s.cfg.Database, s.cfg.Table)
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
  source_id   VARCHAR(191) NOT NULL,
  flavor      VARCHAR(32)  NOT NULL,
  mode        ENUM('gtid','pos') NOT NULL,
  gtid_set    TEXT NULL,
  binlog_file VARCHAR(255) NULL,
  binlog_pos  BIGINT UNSIGNED NULL,
  server_id   BIGINT UNSIGNED NULL,
  updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  version     BIGINT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (source_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`, table)
	result, err := conn.Execute(query)
	if result != nil {
		result.Close()
	}
	if err != nil {
		return fmt.Errorf("ensure state table: %w", err)
	}
	return nil
}

func (s *MySQLStore) Load(ctx context.Context) (*State, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	conn, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	table, err := qualifiedTable(s.cfg.Database, s.cfg.Table)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(s.cfg.SourceID) == "" {
		return nil, errors.New("mysql source_id is empty")
	}
	query := fmt.Sprintf("SELECT source_id, flavor, mode, gtid_set, binlog_file, binlog_pos, server_id, updated_at, version FROM %s WHERE source_id=?", table)
	result, err := conn.Execute(query, s.cfg.SourceID)
	if err != nil {
		return nil, fmt.Errorf("query state: %w", err)
	}
	if result != nil {
		defer result.Close()
	}
	if result == nil || result.Resultset == nil || result.Resultset.RowNumber() == 0 {
		return nil, ErrNotFound
	}

	row := result.Resultset
	state := &State{}
	state.SourceID, err = readString(row, 0, 0)
	if err != nil {
		return nil, err
	}
	state.Flavor, err = readString(row, 0, 1)
	if err != nil {
		return nil, err
	}
	mode, err := readString(row, 0, 2)
	if err != nil {
		return nil, err
	}
	state.Mode = Mode(strings.ToLower(mode))
	state.GTIDSet, err = readString(row, 0, 3)
	if err != nil {
		return nil, err
	}
	state.BinlogFile, err = readString(row, 0, 4)
	if err != nil {
		return nil, err
	}
	state.BinlogPos, err = readUint64(row, 0, 5)
	if err != nil {
		return nil, err
	}
	serverID, err := readUint64(row, 0, 6)
	if err != nil {
		return nil, err
	}
	state.ServerID = uint32(serverID)
	state.UpdatedAt, err = readTime(row, 0, 7)
	if err != nil {
		return nil, err
	}
	state.Version, err = readUint64(row, 0, 8)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (s *MySQLStore) Save(ctx context.Context, st *State) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if st == nil {
		return errors.New("state is nil")
	}
	conn, err := s.connect(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	table, err := qualifiedTable(s.cfg.Database, s.cfg.Table)
	if err != nil {
		return err
	}
	sourceID := strings.TrimSpace(st.SourceID)
	if sourceID == "" {
		sourceID = strings.TrimSpace(s.cfg.SourceID)
	}
	if sourceID == "" {
		return errors.New("mysql source_id is empty")
	}

	st.Touch(timeNowUTC())
	query := fmt.Sprintf(`INSERT INTO %s
		(source_id, flavor, mode, gtid_set, binlog_file, binlog_pos, server_id, updated_at, version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			flavor=VALUES(flavor),
			mode=VALUES(mode),
			gtid_set=VALUES(gtid_set),
			binlog_file=VALUES(binlog_file),
			binlog_pos=VALUES(binlog_pos),
			server_id=VALUES(server_id),
			updated_at=VALUES(updated_at),
			version=VALUES(version)`, table)
	result, err := conn.Execute(
		query,
		sourceID,
		st.Flavor,
		string(st.Mode),
		st.GTIDSet,
		st.BinlogFile,
		st.BinlogPos,
		st.ServerID,
		st.UpdatedAt,
		st.Version,
	)
	if result != nil {
		result.Close()
	}
	if err != nil {
		return fmt.Errorf("save state: %w", err)
	}
	return nil
}

func (s *MySQLStore) connect(ctx context.Context) (*client.Conn, error) {
	if s == nil {
		return nil, errors.New("mysql store is nil")
	}
	cfg := s.cfg
	return client.ConnectWithContext(ctx, cfg.Addr, cfg.User, cfg.Password, cfg.Database, cfg.Timeout)
}

func qualifiedTable(db, table string) (string, error) {
	table = strings.TrimSpace(table)
	if table == "" {
		return "", errors.New("mysql table is empty")
	}
	if strings.Contains(table, ".") {
		parts := strings.Split(table, ".")
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid table name %q", table)
		}
		return sql.QuoteIdentifier(parts[0]) + "." + sql.QuoteIdentifier(parts[1]), nil
	}
	if strings.TrimSpace(db) != "" {
		return sql.QuoteIdentifier(db) + "." + sql.QuoteIdentifier(table), nil
	}
	return sql.QuoteIdentifier(table), nil
}

func readString(rs *mysql.Resultset, row, col int) (string, error) {
	if rs == nil {
		return "", errors.New("empty resultset")
	}
	value, err := rs.GetValue(row, col)
	if err != nil {
		return "", err
	}
	switch v := value.(type) {
	case nil:
		return "", nil
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		return "", fmt.Errorf("unexpected value type %T", value)
	}
}

func readUint64(rs *mysql.Resultset, row, col int) (uint64, error) {
	if rs == nil {
		return 0, errors.New("empty resultset")
	}
	value, err := rs.GetValue(row, col)
	if err != nil {
		return 0, err
	}
	switch v := value.(type) {
	case nil:
		return 0, nil
	case uint64:
		return v, nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("negative value %d", v)
		}
		return uint64(v), nil
	case []byte:
		return strconv.ParseUint(string(v), 10, 64)
	case string:
		return strconv.ParseUint(v, 10, 64)
	default:
		return 0, fmt.Errorf("unexpected value type %T", value)
	}
}

func readTime(rs *mysql.Resultset, row, col int) (time.Time, error) {
	s, err := readString(rs, row, col)
	if err != nil {
		return time.Time{}, err
	}
	if strings.TrimSpace(s) == "" {
		return time.Time{}, nil
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("parse time %q", s)
}
