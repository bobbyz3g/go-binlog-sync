# 008 过滤下的 DDL

## 目标
验证被过滤库/表的 DDL 不会传播到 dest。

## 前置条件
- 修改 `config.yaml`：
  - `filter.whitelist.databases=["gbs_test"]`
- gbs 已重启生效。

## 步骤
1. 在 source 对被过滤库执行 DDL。

```bash
mysql -h 127.0.0.1 -P 3307 -uroot -proot < localtests/sql/11_workload_filter_ddl.sql
```

## 验证
- dest 中不存在 `gbs_filter.ddl_filtered`。
