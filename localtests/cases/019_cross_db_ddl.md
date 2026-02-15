# 019 跨库 DDL

## 目标
验证跨库 DDL/写入可同步。

## 前置条件
- gbs 运行中。

## 步骤
1. 在 source 创建新库并写入数据。

```bash
mysql -h 127.0.0.1 -P 3307 -uroot -proot < localtests/sql/10_workload_cross_db.sql
```

## 验证
- dest 中存在 `gbs_cross.cross_tbl`。
- 行数与 source 一致。
