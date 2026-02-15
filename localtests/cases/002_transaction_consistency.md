# 002 事务一致性

## 目标
验证事务内的多条语句能一致同步到 dest。

## 前置条件
- gbs 已使用 GTID 模式运行。
- `gbs_test.txn_tbl` 已存在（见 01_tables.sql）。

## 步骤
1. 在 source 执行事务负载。

```bash
mysql -h 127.0.0.1 -P 3307 -uroot -proot < localtests/sql/04_workload_txn.sql
```

## 验证
- dest 中 `gbs_test.txn_tbl` 新增 2 行（note=txn-1/txn-2）。
- dest 中 `gbs_test.orders` id=1 的 status 为 `paid`。
