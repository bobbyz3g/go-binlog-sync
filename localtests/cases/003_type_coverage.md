# 003 多类型字段同步

## 目标
验证常见字段类型在同步中保持一致。

## 前置条件
- 已执行 `02_seed.sql`。

## 步骤
1. 在 source 插入一条多类型记录（已由 `02_seed.sql` 完成）。

## 验证
- dest 中 `gbs_test.types` 至少 1 行。
- 抽样比对字段：`c_json`、`c_decimal`、`c_blob`、`c_date`、`c_time`、`c_ts`。
