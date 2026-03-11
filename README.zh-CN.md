# go-binlog-sync (gbs)

[English](README.md)

一个 MySQL/MariaDB binlog 同步工具。它从源 MySQL 实例实时读取 binlog 事件，并将这些事件应用到目标实例，以实现数据复制。

## 功能特性

- **实时复制**：持续拉取 binlog 事件，并将 INSERT、UPDATE、DELETE 和 DDL 操作应用到目标库
- **支持 GTID 与位点复制**：同时支持 GTID 模式和传统的 binlog 文件/位置模式
- **兼容 MySQL 与 MariaDB**：两种数据库都可用
- **状态持久化**：可将同步进度保存到本地文件或 MySQL 表，重启后可安全续传，避免重复回放
- **源库预检查**：启动前校验源库配置，如 `log_bin`、`binlog_format`、`server_id` 和 GTID 相关设置
- **优雅退出**：处理 `SIGINT`/`SIGTERM`，退出前尽量完成状态刷新

## 环境要求

- Go 1.26
- 源 MySQL 需要开启 binlog，且 `binlog_format=ROW`

## 构建

```bash
make build    # 构建 linux/arm64 和 linux/amd64 二进制到 bin/
make test     # 运行全部测试
make clean    # 清理构建产物
```

如果只想为当前平台构建：

```bash
go build -o gbs ./cmd/gbs
```

## 配置

复制 `config-template.yaml` 后按需修改：

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
  serverID: 32            # 唯一的复制 server ID
  host: 127.0.0.1
  port: 3306
  user: root
  password: ""
  gtidEnabled: true       # GTID 模式为 true，位点模式为 false
  gtidSet: ""             # 起始 GTID 集合（GTID 模式）
  binlog: ""              # 起始 binlog 文件名（位点模式）
  position: 0             # 起始 binlog 位置（位点模式）

destination:
  host: 127.0.0.1
  port: 3306
  user: root
  password: ""

filter:
  whitelist:
    databases: []         # 数据库名
    tables: []            # 表名或 db.table
  blacklist:
    databases: []
    tables: []

state:
  enabled: false
  type: file              # file 或 mysql
  everyEvents: 100        # 每处理 N 个事件做一次 checkpoint
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

过滤规则说明：

- 如果 whitelist 和 blacklist 都为空，则同步所有数据库和表。
- blacklist 优先级始终高于 whitelist。
- 表规则支持 `db.table` 或 `table` 两种写法；只写 `table` 时表示匹配任意数据库中的同名表。
- DDL 过滤针对常见数据库和表级语句做尽力而为的处理；行事件始终严格遵守过滤规则。

## 使用方式

```bash
gbs -config config.yaml
```

## 监控

HTTP 服务同时会暴露 Prometheus 指标：

- `GET /metrics`
- 完整地址示例：`http://127.0.0.1:8081/metrics`，具体取决于 `server.host` 和 `server.port`

Prometheus 抓取配置示例：

```yaml
scrape_configs:
  - job_name: gbs
    static_configs:
      - targets: ["127.0.0.1:8081"]
```

核心指标：

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

建议优先配置以下告警：

- `gbs_worker_up == 0` 持续 1 分钟
- `gbs_binlog_read_errors_total` 增速异常升高
- `gbs_event_apply_errors_total` 增速异常升高
- `gbs_replication_lag_seconds` 持续高于你的 SLO 阈值
- `gbs_state_checkpoint_total{result="success"}` 在较长时间窗口内没有增长

### 源库 MySQL 配置

确保源 MySQL 实例满足以下设置：

```sql
-- 检查是否开启 binlog
SHOW VARIABLES LIKE 'log_bin';            -- 必须为 ON

-- 检查 binlog 格式
SHOW VARIABLES LIKE 'binlog_format';      -- 必须为 ROW

-- 检查 binlog row image
SHOW VARIABLES LIKE 'binlog_row_image';   -- 推荐 FULL

-- GTID 模式（MySQL）
SHOW VARIABLES LIKE 'gtid_mode';          -- 必须为 ON
SHOW VARIABLES LIKE 'enforce_gtid_consistency'; -- 必须为 ON

-- GTID 模式（MariaDB）
SHOW VARIABLES LIKE 'gtid_strict_mode';   -- 必须为 ON
```

源库上的复制账号至少需要具备以下权限：

```sql
GRANT REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'repl_user'@'%';
```

## 架构

```text
Source MySQL
    |
    v
BinlogReader  -- 通过复制协议拉取 binlog 事件
    |
    v
Event Channel
    |
    v
EventWriter   -- 将事件应用到目标库
    |           ├── RowsEvent  -> INSERT / UPDATE / DELETE
    |           ├── QueryEvent -> DDL
    |           └── Other      -> skip
    v
StateRecorder -- 持久化 checkpoint（文件或 MySQL）
```

## 本地测试

仓库提供了本地端到端测试说明，见 [localtests/README.md](localtests/README.md)。

## 许可证

[BSD 3-Clause](LICENSE)
