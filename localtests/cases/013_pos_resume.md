# 013 位置续传

## 目标
验证 position 模式下断点续传。

## 前置条件
- 使用 `localtests/configs/pos-mysql.yaml`。
- 在 source 执行 `SHOW MASTER STATUS` 获取 `binlog` 与 `position`，写入 config。

## 步骤
1. 启动 gbs。
2. 在 source 执行 `03_workload_basic.sql`。
3. 停止 gbs。
4. 在 source 再插入 1 条数据。
5. 启动 gbs。

## 验证
- dest 最终数据与 source 一致。
- 重启后只补齐新增事件，无重复。
