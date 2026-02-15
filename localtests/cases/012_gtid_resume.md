# 012 GTID 续传

## 目标
验证 GTID 模式下断点续传。

## 前置条件
- 使用 `localtests/configs/gtid-mysql.yaml` 并设置 `gtidSet`。

## 步骤
1. 启动 gbs。
2. 在 source 执行 `03_workload_basic.sql`。
3. 停止 gbs。
4. 在 source 再插入 1 条数据（如 `INSERT INTO gbs_test.users ...`）。
5. 启动 gbs。

## 验证
- dest 最终数据与 source 一致。
- 重启后只补齐新增事务，无重复。
