# localtests

本目录用于手工/半自动的本地端到端测试（E2E）。结构参考 gh-ost 的 localtests 思路：
- 使用 docker-compose 启动多实例数据库
- 使用分离的用例文档描述步骤与验证点
- 用最小脚本与 SQL 文件支撑用例执行

## 目录结构

- `docker-compose.yml`：启动 source/dest/state 数据库实例
- `configs/`：gbs 配置样例（GTID/位置模式）
- `sql/`：通用初始化与工作负载 SQL
- `cases/`：用例清单与详细步骤

## 快速开始（手工）

1. 启动数据库

```bash
docker compose -f localtests/docker-compose.yml up -d
```

或使用轻量脚本：

```bash
localtests/run.zsh up
```

2. 初始化 source/dest

```bash
mysql -h 127.0.0.1 -P 3307 -uroot -proot < localtests/sql/00_create_dbs.sql
mysql -h 127.0.0.1 -P 3307 -uroot -proot < localtests/sql/01_tables.sql
mysql -h 127.0.0.1 -P 3307 -uroot -proot < localtests/sql/02_seed.sql
```

或使用轻量脚本：

```bash
localtests/run.zsh seed
```

3. 生成配置并启动 gbs

```bash
cp localtests/configs/gtid-mysql.yaml config.yaml
./gbs -config config.yaml
```

4. 逐个执行 `cases/` 中的测试用例，按步骤验证。

## 轻量运行脚本

`localtests/run.zsh` 提供常用操作封装：

```bash
localtests/run.zsh up
localtests/run.zsh seed
localtests/run.zsh basic
localtests/run.zsh check-basic
```

查看完整用法：

```bash
localtests/run.zsh --help
```

## 注意事项

- 端口约定：source=3307，dest=3308，state-mysql=3309。
- GTID 模式需要先执行 `SELECT @@GLOBAL.gtid_executed`，写入配置的 `gtidSet`。
- 位置模式需要执行 `SHOW MASTER STATUS`，写入 `binlog` 和 `position`。
