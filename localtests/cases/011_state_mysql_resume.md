# 011 MySQL 状态续传

## 目标
验证 state(mysql) 续传正确。

## 前置条件
- 使用 `localtests/configs/gtid-mysql-state-mysql.yaml`。
- gbs 已重启生效。

## 步骤
1. 运行 `03_workload_basic.sql`。
2. 停止 gbs（Ctrl+C）。
3. 在 source 再插入新记录。
4. 启动 gbs。

## 验证
- dest 最终数据与 source 一致。
- state 库中的 `gbs_state.gbs_sync_state` 有对应 source_id 的记录，且 `updated_at` 与 `version` 递增。
