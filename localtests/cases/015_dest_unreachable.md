# 015 目标库不可达

## 目标
验证 dest 不可达时 gbs 能明确报错。

## 前置条件
- gbs 已配置为连接 dest。

## 步骤
1. 停止 dest 容器：

```bash
docker compose -f localtests/docker-compose.yml stop mysql-dest
```

2. 启动 gbs。

## 验证
- gbs 日志包含 `connect destination` 相关错误。
- gbs 退出或不再继续写入。
