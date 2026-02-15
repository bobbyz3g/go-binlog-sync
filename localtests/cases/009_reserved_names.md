# 009 保留字/特殊表名

## 目标
验证带保留字或特殊字符的表名可正确同步。

## 前置条件
- `gbs_reserved` 已创建（见 00/01_tables.sql）。

## 步骤
1. 在 source 写入特殊表名。

```bash
mysql -h 127.0.0.1 -P 3307 -uroot -proot < localtests/sql/08_workload_reserved.sql
```

## 验证
- dest 中 `gbs_reserved`.`order` 与 `gbs_reserved`.`user-log` 行数与 source 一致。
