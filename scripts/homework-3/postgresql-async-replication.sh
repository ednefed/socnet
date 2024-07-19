#!/usr/bin/env bash

set -o errexit

compose_cmd="docker compose -f docker-compose.yaml"
postgresql_cmd="$compose_cmd exec postgresql psql -U postgres"
existing_user=$($postgresql_cmd -tAc "SELECT 1 FROM pg_roles WHERE rolname='replicator'")

# Создаём пользователя replicator, если его ещё нет
if [ "$existing_user" = "1" ]; then
	echo "User 'replicator' already exists"
else 
	$postgresql_cmd -c "CREATE USER replicator WITH REPLICATION PASSWORD 'replicator';"
fi

# Добавляем для репликации запись в pg_hba.conf, если ее ещё нет
if $compose_cmd exec postgresql grep replicator /var/lib/postgresql/data/pg_hba.conf &> /dev/null; then
	echo "pg_hba.conf entry already exists"
else
	$compose_cmd exec postgresql sh -c 'echo "host replication replicator 0.0.0.0/0 md5" | tee -a /var/lib/postgresql/data/pg_hba.conf'
fi

# Включаем потоковую репликацию на мастере
$postgresql_cmd -c "ALTER SYSTEM SET wal_level = 'replica';" &> /dev/null
$postgresql_cmd -c "ALTER SYSTEM SET synchronous_commit = off;" &> /dev/null
$postgresql_cmd -c "ALTER SYSTEM SET synchronous_standby_names = '';" &> /dev/null
$postgresql_cmd -c "SELECT pg_reload_conf();" &> /dev/null

echo "Doing pg_basebackup for postgresql2"

# Создаём полный бэкап для первой реплики, сохраняя результат в volume
# --write-recovery-conf ообеспечит standby.signal файл и параметры подключения к мастеру в postgresql.auto.conf 
if ! $compose_cmd --profile postgresql-replica ps | grep postgresql2 &> /dev/null; then
	docker run --rm -t \
		-e PGPASSWORD=replicator \
		--network social_network_default \
		-v social_network_postgresql2_data:/var/lib/postgresql/data \
		postgres:16.3-alpine \
			pg_basebackup \
				--host=postgresql \
				--username=replicator \
				--wal-method=stream \
				--write-recovery-conf \
				--pgdata=/var/lib/postgresql/data \
				--progress
fi

echo "Doing pg_basebackup for postgresql3"

# Создаём полный бэкап для второй реплики
if ! $compose_cmd --profile postgresql-replica ps | grep postgresql3 &> /dev/null; then
	docker run --rm -t \
		-e PGPASSWORD=replicator \
		--network social_network_default \
		-v social_network_postgresql3_data:/var/lib/postgresql/data \
		postgres:16.3-alpine \
			pg_basebackup \
				--host=postgresql \
				--username=replicator \
				--wal-method=stream \
				--write-recovery-conf \
				--pgdata=/var/lib/postgresql/data \
				--progress
fi

# Запускаем реплики
echo "Spinning up replicas"
$compose_cmd --profile postgresql-replica up --wait postgresql2 postgresql3
echo "Replication state:"
$postgresql_cmd -c "SELECT application_name, client_addr, state, sync_state, sent_lsn, replay_lsn FROM pg_stat_replication;"
