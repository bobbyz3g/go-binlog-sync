# 006 黑名单优先生效

## 目标
验证 blacklist 优先于 whitelist。

## 前置条件
- 修改 `config.yaml`：
  - `filter.whitelist.databases=["gbs_filter"]`
  - `filter.blacklist.tables=["gbs_filter.deny_table"]`
- gbs 已重启生效。

## 步骤
1. 在 source 写入过滤相关数据。

```bash
mysql -h 127.0.0.1 -P 3307 -uroot -proot < localtests/sql/07_workload_filter.sql
```

## 验证
- dest 中 `gbs_filter.allow_table` 有新增。
- dest 中 `gbs_filter.deny_table` 无新增（blacklist 生效）。
