# 010 文件状态续传

## 目标
验证 state(file) 续传正确。

## 前置条件
- 修改 `config.yaml`：
  - `state.enabled=true`
  - `state.type=file`
  - `state.everyEvents=1`
- gbs 已重启生效。

## 步骤
1. 运行 `03_workload_basic.sql`。
2. 停止 gbs（Ctrl+C）。
3. 在 source 再插入一条记录（手工或再执行一次 `03_workload_basic.sql`）。
4. 启动 gbs。

## 验证
- dest 最终数据与 source 一致，无重复行。
- `gbs.state.json` 更新时间与内容有变化。
