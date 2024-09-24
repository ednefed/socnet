SELECT citus_set_coordinator_host('citus_1');
SELECT * from citus_add_node('citus_worker_1', 5432);
SELECT * FROM citus_add_node('citus_worker_2', 5432);
SELECT * FROM pg_dist_node;
SELECT * FROM citus_get_active_worker_nodes();
ALTER SYSTEM SET wal_level = logical;
SELECT run_command_on_workers('ALTER SYSTEM SET wal_level = logical');

DROP TABLE IF EXISTS public.dialogs;
CREATE TABLE public.dialogs (
    from_user integer NOT NULL,
    to_user integer NOT NULL,
    message text NOT NULL,
    created_at timestamptz NOT NULL default now(),
    CONSTRAINT dialogs_uniq UNIQUE (from_user, to_user, message, created_at)
);
SELECT create_distributed_table('dialogs', 'to_user');
CREATE INDEX dialogs_from_user ON public.dialogs (from_user, to_user, created_at DESC);
ALTER TABLE dialogs REPLICA IDENTITY
  USING INDEX dialogs_uniq;
