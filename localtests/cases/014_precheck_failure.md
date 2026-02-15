# 014 precheck 失败

## 目标
验证 source precheck 失败时 gbs 直接退出并给出清晰日志。

## 前置条件
- 可修改 `localtests/docker-compose.yml` 并重启 source。

## 步骤（任选其一）

A. 关闭 binlog
1. 修改 `mysql-source` 启动参数，移除 `--log-bin`。
2. `docker compose -f localtests/docker-compose.yml up -d --force-recreate mysql-source`
3. 启动 gbs。

B. 使用非 ROW 格式
1. 将 `--binlog-format=ROW` 改为 `--binlog-format=STATEMENT`。
2. 重启 `mysql-source`。
3. 启动 gbs。

## 验证
- gbs 日志出现 `precheck failed`，包含具体变量名与期望值。
- gbs 进程退出或不进入同步循环。
