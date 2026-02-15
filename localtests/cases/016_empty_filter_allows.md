# 016 空过滤全量同步

## 目标
验证 whitelist/blacklist 都为空时全量同步。

## 前置条件
- `filter.whitelist` 与 `filter.blacklist` 均为空。

## 步骤
1. 在多个库写入数据（如执行 `02_seed.sql` 与 `07_workload_filter.sql`）。

## 验证
- dest 中 `gbs_test`、`gbs_filter`、`gbs_reserved` 数据与 source 一致。
