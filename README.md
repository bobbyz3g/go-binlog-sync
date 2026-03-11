# go-binlog-sync (gbs)

[中文说明](README.zh-CN.md)

A MySQL/MariaDB binlog synchronization tool. It reads binlog events from a source MySQL instance and applies them to a destination instance in real time, enabling data replication.

## Features

- **Real-time replication** -- streams binlog events and applies INSERT, UPDATE, DELETE, and DDL operations to the destination
- **GTID & position-based replication** -- supports both GTID mode and traditional binlog file/position mode
- **MySQL & MariaDB** -- works with both flavors
- **State persistence** -- checkpoint progress to a local file or a MySQL table, allowing safe restart without replaying events
- **Source precheck** -- validates source configuration (`log_bin`, `binlog_format`, `server_id`, GTID settings) before starting
- **Graceful shutdown** -- handles SIGINT/SIGTERM with proper state flushing

## Requirements

- Go 1.26
- Source MySQL must have binlog enabled with `binlog_format=ROW`

## Build

```bash
make build    # builds linux/arm64 and linux/amd64 binaries to bin/
make test     # runs all tests
make clean    # removes build artifacts
```

To build for the current platform:

```bash
go build -o gbs ./cmd/gbs
```

## Configuration

Copy `config-template.yaml` and edit it:

```bash
cp config-template.yaml config.yaml
```

```yaml
log:
  level: info             # debug, info, warn, error

server:
  host: 0.0.0.0
  port: 8081

source:
  serverID: 32            # unique replication server ID
  host: 127.0.0.1
  port: 3306
  user: root
  password: ""
  gtidEnabled: true       # true for GTID mode, false for position mode
  gtidSet: ""             # starting GTID set (GTID mode)
  binlog: ""              # starting binlog file (position mode)
  position: 0             # starting binlog position (position mode)

destination:
  host: 127.0.0.1
  port: 3306
  user: root
  password: ""

filter:
  whitelist:
    databases: []         # database names
    tables: []            # table names or db.table
  blacklist:
    databases: []
    tables: []

state:
  enabled: false
  type: file              # file or mysql
  everyEvents: 100        # checkpoint every N events
  filePath: gbs.state.json
  mysql:
    host: 127.0.0.1
    port: 3306
    user: root
    password: ""
    database: ""
    table: gbs_sync_state
    sourceID: ""
```

Filter notes:

- If both whitelist and blacklist are empty, all databases/tables are synced.
- Blacklist entries always win over whitelist entries.
- Table entries accept `db.table` or `table` (matches any database).
- DDL filtering is best-effort for common database/table statements; row events always obey the filter.

## Usage

```bash
gbs -config config.yaml
```

## Monitoring

The HTTP server also exposes Prometheus metrics at:

- `GET /metrics`
- Full URL example: `http://127.0.0.1:8081/metrics` (based on `server.host` and `server.port`)

Example Prometheus scrape config:

```yaml
scrape_configs:
  - job_name: gbs
    static_configs:
      - targets: ["127.0.0.1:8081"]
```

Core metrics:

- `gbs_worker_up`
- `gbs_binlog_events_read_total{event_type}`
- `gbs_binlog_read_errors_total`
- `gbs_replication_lag_seconds`
- `gbs_events_filtered_total{kind}`
- `gbs_events_applied_total{event_type}`
- `gbs_event_apply_errors_total{stage}`
- `gbs_event_apply_duration_seconds{event_type}`
- `gbs_state_checkpoint_total{result}`
- `gbs_state_last_checkpoint_timestamp_seconds`

Suggested initial alerts:

- `gbs_worker_up == 0` for 1 minute
- high increase rate on `gbs_binlog_read_errors_total`
- high increase rate on `gbs_event_apply_errors_total`
- `gbs_replication_lag_seconds` continuously above your SLO threshold
- no increase in `gbs_state_checkpoint_total{result="success"}` for a long window

### Source MySQL setup

Ensure the source MySQL instance has the required settings:

```sql
-- Check binlog is enabled
SHOW VARIABLES LIKE 'log_bin';            -- must be ON

-- Check binlog format
SHOW VARIABLES LIKE 'binlog_format';      -- must be ROW

-- Check binlog row image
SHOW VARIABLES LIKE 'binlog_row_image';   -- FULL recommended

-- For GTID mode (MySQL)
SHOW VARIABLES LIKE 'gtid_mode';          -- must be ON
SHOW VARIABLES LIKE 'enforce_gtid_consistency'; -- must be ON

-- For GTID mode (MariaDB)
SHOW VARIABLES LIKE 'gtid_strict_mode';   -- must be ON
```

The replication user needs at minimum `REPLICATION SLAVE` and `REPLICATION CLIENT` privileges on the source:

```sql
GRANT REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'repl_user'@'%';
```

## Architecture

```
Source MySQL
    |
    v
BinlogReader  -- streams binlog events via replication protocol
    |
    v
Event Channel
    |
    v
EventWriter   -- applies events to destination
    |           ├── RowsEvent  -> INSERT / UPDATE / DELETE
    |           ├── QueryEvent -> DDL
    |           └── Other      -> skip
    v
StateRecorder -- persists checkpoint (file or MySQL)
```

## License

[BSD 3-Clause](LICENSE)
