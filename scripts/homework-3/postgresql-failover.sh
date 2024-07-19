#!/usr/bin/env bash

set -o errexit

compose_cmd="docker compose -f docker-compose.yaml --profile postgresql-replica"
postgresql2_cmd="$compose_cmd exec postgresql2 psql -U postgres"
postgresql3_cmd="$compose_cmd exec postgresql3 psql -U postgres"

# Kill master, get replicas lsn
echo "Killing master"
$compose_cmd stop postgresql
echo "Getting replicas lsn"
postgresql2_lsn=$($postgresql2_cmd -tAc "SELECT pg_last_wal_replay_lsn()")
postgresql3_lsn=$($postgresql3_cmd -tAc "SELECT pg_last_wal_replay_lsn()")
echo "postgresql2: $postgresql2_lsn"
echo "postgresql3: $postgresql3_lsn"

# if $postgresql2_lsn > $postgresql3_lsn then postgresql2 becomes master
if [[ "$postgresql2_lsn" > "$postgresql3_lsn" ]]; then
    # Promote postgresql2, switch postgresql3 from old master to postgresql2
    echo "postgresql2 will become master"
    $postgresql2_cmd -c "SELECT pg_promote()" &> /dev/null
    $postgresql2_cmd -c "ALTER SYSTEM SET synchronous_commit = on;" &> /dev/null
    $postgresql2_cmd -c "ALTER SYSTEM SET synchronous_standby_names = 'ANY 1 (postgresql, postgresql3)';" &> /dev/null
    $postgresql2_cmd -c "SELECT pg_reload_conf();" &> /dev/null
    $postgresql3_cmd -c "ALTER SYSTEM SET primary_conninfo = 'user=replicator password=replicator channel_binding=prefer dbname=replication host=postgresql2 port=5432 fallback_application_name=walreceiver sslmode=prefer sslcompression=0 sslcertmode=allow sslsni=1 ssl_min_protocol_version=TLSv1.2 gssencmode=prefer krbsrvname=postgres gssdelegation=0 target_session_attrs=any load_balance_hosts=disable'" &> /dev/null
    $postgresql3_cmd -c "SELECT pg_reload_conf();" &> /dev/null
    $postgresql2_cmd -c "SELECT application_name, client_addr, state, sync_state, sent_lsn, replay_lsn FROM pg_stat_replication;"
    $postgresql2_cmd -c "SELECT max(id) from public.users;"
else
    # Promote postgresql3, switch postgresql2 from old master to postgresql3
    echo "postgresql3 will become master"
    $postgresql3_cmd -c "SELECT pg_promote()" &> /dev/null
    $postgresql3_cmd -c "ALTER SYSTEM SET synchronous_commit = on;" &> /dev/null
    $postgresql3_cmd -c "ALTER SYSTEM SET synchronous_standby_names = 'ANY 1 (postgresql, postgresql2)';" &> /dev/null
    $postgresql3_cmd -c "SELECT pg_reload_conf();" &> /dev/null
    $postgresql2_cmd -c "ALTER SYSTEM SET primary_conninfo = 'user=replicator password=replicator channel_binding=prefer dbname=replication host=postgresql3 port=5432 fallback_application_name=walreceiver sslmode=prefer sslcompression=0 sslcertmode=allow sslsni=1 ssl_min_protocol_version=TLSv1.2 gssencmode=prefer krbsrvname=postgres gssdelegation=0 target_session_attrs=any load_balance_hosts=disable'" &> /dev/null
    $postgresql2_cmd -c "SELECT pg_reload_conf();" &> /dev/null
    $postgresql3_cmd -c "SELECT application_name, client_addr, state, sync_state, sent_lsn, replay_lsn FROM pg_stat_replication;"
    $postgresql3_cmd -c "SELECT max(id) from public.users;"
fi
