# 004 批量写入性能与一致性

## 目标
验证较大批量写入的同步稳定性与一致性。

## 前置条件
- gbs 运行中。

## 步骤
1. 在 source 执行批量插入。

```bash
mysql -h 127.0.0.1 -P 3307 -uroot -proot < localtests/sql/05_workload_batch.sql
```

## 验证
- dest 中 `gbs_test.txn_tbl` 的行数增加 1000。
