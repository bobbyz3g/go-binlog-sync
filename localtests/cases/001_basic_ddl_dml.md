# 001 基础 DDL/DML 同步

## 目标
验证基础 DML 与简单 DDL 能从 source 同步到 dest。

## 前置条件
- 已启动 docker compose。
- 使用 `localtests/configs/gtid-mysql.yaml` 并写入 `gtidSet`。

## 步骤
1. 初始化库表与数据（source）。

```bash
mysql -h 127.0.0.1 -P 3307 -uroot -proot < localtests/sql/00_create_dbs.sql
mysql -h 127.0.0.1 -P 3307 -uroot -proot < localtests/sql/01_tables.sql
mysql -h 127.0.0.1 -P 3307 -uroot -proot < localtests/sql/02_seed.sql
```

2. 启动 gbs。

```bash
cp localtests/configs/gtid-mysql.yaml config.yaml
./gbs -config config.yaml
```

3. 触发 DML + 简单 DDL（source）。

```bash
mysql -h 127.0.0.1 -P 3307 -uroot -proot < localtests/sql/03_workload_basic.sql
mysql -h 127.0.0.1 -P 3307 -uroot -proot < localtests/sql/06_workload_ddl.sql
```

## 验证
- dest 中 `gbs_test.users`、`gbs_test.orders` 与 source 行数一致。
- dest 中 `gbs_test.ddl_tbl` 不存在（已在 source drop）。

## 清理
无（复用数据即可）。
