#!/usr/bin/env bash

set -o errexit

postgresql_cmd="docker compose -f docker-compose.yaml exec postgresql psql -U postgres"

# Включаем кворумный (ANY) синхронный коммит (synchronous_commit) на мастере с подтверждением от любой (ANY 1) из двух реплик
$postgresql_cmd -c "ALTER SYSTEM SET synchronous_commit = on;" &> /dev/null
$postgresql_cmd -c "ALTER SYSTEM SET synchronous_standby_names = 'ANY 1 (postgresql2, postgresql3)';" &> /dev/null
$postgresql_cmd -c "SELECT pg_reload_conf();" &> /dev/null
$postgresql_cmd -c "SELECT application_name, client_addr, state, sync_state, sent_lsn, replay_lsn FROM pg_stat_replication;"
