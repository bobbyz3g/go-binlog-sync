#!/usr/bin/env zsh
set -euo pipefail

ROOT=${0:A:h}
PROJECT_ROOT=${ROOT:h}

MYSQL_SOURCE_HOST=${MYSQL_SOURCE_HOST:-127.0.0.1}
MYSQL_SOURCE_PORT=${MYSQL_SOURCE_PORT:-3307}
MYSQL_DEST_HOST=${MYSQL_DEST_HOST:-127.0.0.1}
MYSQL_DEST_PORT=${MYSQL_DEST_PORT:-3308}
MYSQL_USER=${MYSQL_USER:-root}
MYSQL_PASSWORD=${MYSQL_PASSWORD:-root}

compose() {
  docker compose -f "${ROOT}/docker-compose.yml" "$@"
}

mysql_source() {
  mysql -h "${MYSQL_SOURCE_HOST}" -P "${MYSQL_SOURCE_PORT}" -u"${MYSQL_USER}" -p"${MYSQL_PASSWORD}" "$@"
}

mysql_dest() {
  mysql -h "${MYSQL_DEST_HOST}" -P "${MYSQL_DEST_PORT}" -u"${MYSQL_USER}" -p"${MYSQL_PASSWORD}" "$@"
}

run_sql_source() {
  local file="$1"
  mysql_source < "${file}"
}

count_table() {
  local db="$1"
  local table="$2"
  mysql_source -N -e "SELECT COUNT(*) FROM ${db}.${table};"
}

count_table_dest() {
  local db="$1"
  local table="$2"
  mysql_dest -N -e "SELECT COUNT(*) FROM ${db}.${table};"
}

compare_count() {
  local db="$1"
  local table="$2"
  local s
  local d
  s=$(count_table "${db}" "${table}")
  d=$(count_table_dest "${db}" "${table}")
  if [[ "${s}" == "${d}" ]]; then
    echo "OK count ${db}.${table}=${s}"
  else
    echo "MISMATCH count ${db}.${table}: source=${s} dest=${d}"
    return 1
  fi
}

table_exists_dest() {
  local db="$1"
  local table="$2"
  mysql_dest -N -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema= AND table_name=;"
}

usage() {
  cat <<USAGE
Usage: ${0:t} <command>

Commands:
  up                 Start local databases (docker compose)
  down               Stop local databases
  seed               Create databases, tables, and seed data on source
  basic              Run basic DML workload on source
  txn                Run transaction workload on source
  batch              Run batch workload on source
  ddl                Run DDL workload on source
  filter             Run filter workload on source
  filter-ddl          Run filtered DDL workload on source
  reserved           Run reserved-name workload on source
  schema             Run schema change workload on source
  cross              Run cross-db workload on source
  gtid-set            Print GTID executed from source
  pos                Print SHOW MASTER STATUS from source
  check-basic         Compare counts for users/orders
  check-reserved      Compare counts for reserved-name tables
  check-schema        Verify users.nickname exists on dest
  check-cross         Verify cross-db table exists and counts match
  check-ddl-dropped   Verify ddl_tbl does not exist on dest
  check-filter        Print allow/deny counts (manual verify)

Environment overrides:
  MYSQL_SOURCE_HOST, MYSQL_SOURCE_PORT, MYSQL_DEST_HOST, MYSQL_DEST_PORT,
  MYSQL_USER, MYSQL_PASSWORD
USAGE
}

cmd=${1:-}
case "${cmd}" in
  up)
    compose up -d
    ;;
  down)
    compose down
    ;;
  seed)
    run_sql_source "${ROOT}/sql/00_create_dbs.sql"
    run_sql_source "${ROOT}/sql/01_tables.sql"
    run_sql_source "${ROOT}/sql/02_seed.sql"
    ;;
  basic)
    run_sql_source "${ROOT}/sql/03_workload_basic.sql"
    ;;
  txn)
    run_sql_source "${ROOT}/sql/04_workload_txn.sql"
    ;;
  batch)
    run_sql_source "${ROOT}/sql/05_workload_batch.sql"
    ;;
  ddl)
    run_sql_source "${ROOT}/sql/06_workload_ddl.sql"
    ;;
  filter)
    run_sql_source "${ROOT}/sql/07_workload_filter.sql"
    ;;
  filter-ddl)
    run_sql_source "${ROOT}/sql/11_workload_filter_ddl.sql"
    ;;
  reserved)
    run_sql_source "${ROOT}/sql/08_workload_reserved.sql"
    ;;
  schema)
    run_sql_source "${ROOT}/sql/09_workload_schema_change.sql"
    ;;
  cross)
    run_sql_source "${ROOT}/sql/10_workload_cross_db.sql"
    ;;
  gtid-set)
    mysql_source -N -e "SELECT @@GLOBAL.gtid_executed;"
    ;;
  pos)
    mysql_source -e "SHOW MASTER STATUS;"
    ;;
  check-basic)
    compare_count gbs_test users
    compare_count gbs_test orders
    ;;
  check-reserved)
    compare_count gbs_reserved "order"
    compare_count gbs_reserved "user-log"
    ;;
  check-schema)
    mysql_dest -N -e "SHOW COLUMNS FROM gbs_test.users LIKE nickname;"
    ;;
  check-cross)
    compare_count gbs_cross cross_tbl
    ;;
  check-ddl-dropped)
    if [[ "$(table_exists_dest gbs_test ddl_tbl)" == "0" ]]; then
      echo "OK ddl_tbl is dropped on dest"
    else
      echo "MISMATCH ddl_tbl exists on dest"
      return 1
    fi
    ;;
  check-filter)
    echo "allow_table source=$(count_table gbs_filter allow_table) dest=$(count_table_dest gbs_filter allow_table)"
    echo "deny_table  source=$(count_table gbs_filter deny_table) dest=$(count_table_dest gbs_filter deny_table)"
    ;;
  ""|-h|--help)
    usage
    ;;
  *)
    echo "Unknown command: ${cmd}"
    usage
    exit 1
    ;;
esac
