# E2E 用例清单

每个用例都有独立的步骤与验证点，按需选择执行。默认使用：
- source: 127.0.0.1:3307
- dest: 127.0.0.1:3308
- state-mysql: 127.0.0.1:3309

| ID | 用例 | 依赖 |
| --- | --- | --- |
| 001 | 基础 DDL/DML 同步 | gtid-mysql.yaml |
| 002 | 事务一致性 | gtid-mysql.yaml |
| 003 | 多类型字段同步 | gtid-mysql.yaml |
| 004 | 批量写入性能与一致性 | gtid-mysql.yaml |
| 005 | 白名单过滤 | gtid-mysql.yaml + filter |
| 006 | 黑名单优先生效 | gtid-mysql.yaml + filter |
| 007 | DDL 传播 | gtid-mysql.yaml |
| 008 | 过滤下的 DDL | gtid-mysql.yaml + filter |
| 009 | 保留字/特殊表名 | gtid-mysql.yaml |
| 010 | 文件状态续传 | gtid-mysql.yaml + state(file) |
| 011 | MySQL 状态续传 | gtid-mysql-state-mysql.yaml |
| 012 | GTID 续传 | gtid-mysql.yaml |
| 013 | 位置续传 | pos-mysql.yaml |
| 014 | precheck 失败 | gtid-mysql.yaml + 修改 source 配置 |
| 015 | 目标库不可达 | gtid-mysql.yaml + 修改 dest 配置 |
| 016 | 空过滤全量同步 | gtid-mysql.yaml |
| 017 | Schema 变更后写入 | gtid-mysql.yaml |
| 018 | MariaDB GTID 同步 | gtid-mariadb.yaml |
| 019 | 跨库 DDL | gtid-mysql.yaml |
