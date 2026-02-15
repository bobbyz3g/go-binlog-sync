# 017 Schema 变更后写入

## 目标
验证列变更后继续写入仍可同步。

## 前置条件
- gbs 运行中。

## 步骤
1. 在 source 执行 schema 变更并写入新列。

```bash
mysql -h 127.0.0.1 -P 3307 -uroot -proot < localtests/sql/09_workload_schema_change.sql
```

## 验证
- dest `gbs_test.users` 存在 `nickname` 列。
- 新插入行在 dest 可见且 `nickname` 正确。
