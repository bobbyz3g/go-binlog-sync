# 018 MariaDB GTID 同步

## 目标
验证 MariaDB GTID 模式同步。

## 前置条件
- 启动 MariaDB：

```bash
docker compose -f localtests/docker-compose.yml --profile mariadb up -d mariadb-source
```

- 使用 `localtests/configs/gtid-mariadb.yaml` 并写入 `gtidSet`。

## 步骤
1. 在 MariaDB source 初始化并写入数据（与 MySQL 相同 SQL）。
2. 启动 gbs。
3. 执行 `03_workload_basic.sql`。

## 验证
- dest 数据与 MariaDB source 一致。
