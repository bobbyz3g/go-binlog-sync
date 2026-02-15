# 007 DDL 传播

## 目标
验证 DDL 事件能传播到 dest。

## 前置条件
- gbs 运行中，filter 为空或允许 `gbs_test`。

## 步骤
1. 在 source 执行 DDL 负载。

```bash
mysql -h 127.0.0.1 -P 3307 -uroot -proot < localtests/sql/06_workload_ddl.sql
```

## 验证
- dest 上 `SHOW TABLES LIKE ddl_tbl` 结果为空（已 drop）。
- 若查看 binlog 应看到对应 CREATE/ALTER/DROP 被应用。
