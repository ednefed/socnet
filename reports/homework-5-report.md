# Диалоги
## Шардирование
В качестве горизонтально масштабируемого хранилища используется Citus.
Ключом шардирования, с условием учёта "эффекта Леди Гаги" будет ID пользователя, которому адресуется сообщение (см. [](../infrastructure/citus/schema.sql)). В этом случае, если один пользователь пишет сильно больше среднего, то его сообщения будут распределяться по разным шардам (автор один, а получатели разные), более-менее равномерно распределяя данные по кластеру.

Итоговый диалог между пользователями получается выборкой:
```sql
SELECT from_user, to_user, message, created_at FROM public.dialogs
WHERE (from_user = $1 AND to_user = $2) OR (from_user = $2 AND to_user = $1)
ORDER BY created_at DESC
```
Для ускорения этой выборки созданы два индекса:
```sql
CREATE INDEX dialogs_from_user ON public.dialogs (from_user, to_user, created_at DESC);
CREATE INDEX dialogs_to_user on public.dialogs (to_user, from_user, created_at DESC);
```
Для проверки использовался ситнететически сгенерированный диалог между двумя пользователями (ID 1 и 1001110):
```sql
INSERT INTO public.dialogs(from_user, to_user, message)
SELECT
	(array[1001110, 1])[floor(random() * 2 + 1)],
	(array[1001110, 1])[floor(random() * 2 + 1)],
	md5(random()::text)
FROM generate_series(1, 1000000);
```
План распределённого запроса показывает, что оба индекса используются:
```
Sort  (cost=11381.32..11631.32 rows=100000 width=48) (actual time=2834.222..2964.096 rows=499594 loops=1)
  Sort Key: remote_scan.created_at DESC
  Sort Method: external merge  Disk: 32344kB
  ->  Custom Scan (Citus Adaptive)  (cost=0.00..0.00 rows=100000 width=48) (actual time=2235.879..2448.313 rows=499594 loops=1)
        Task Count: 2
        Tuple data received from nodes: 23 MB
        Tasks Shown: One of 2
        ->  Task
              Tuple data received from node: 11 MB
              Node: host=citus_worker_2 port=5432 dbname=postgres
              ->  Bitmap Heap Scan on dialogs_102047 dialogs  (cost=3485.67..13671.07 rows=249220 width=49) (actual time=4.554..193.127 rows=249724 loops=1)
                    Recheck Cond: (((to_user = 1) AND (from_user = 1001110)) OR ((to_user = 1001110) AND (from_user = 1)))
                    Heap Blocks: exact=5157
                    ->  BitmapOr  (cost=3485.67..3485.67 rows=249220 width=0) (actual time=3.939..3.941 rows=0 loops=1)
                          ->  Bitmap Index Scan on dialogs_to_user_102047  (cost=0.00..4.43 rows=1 width=0) (actual time=0.014..0.015 rows=0 loops=1)
                                Index Cond: ((to_user = 1) AND (from_user = 1001110))
                          ->  Bitmap Index Scan on dialogs_to_user_102047  (cost=0.00..3356.62 rows=249220 width=0) (actual time=3.924..3.924 rows=249724 loops=1)
                                Index Cond: ((to_user = 1001110) AND (from_user = 1))
                  Planning Time: 0.550 ms
                  Execution Time: 520.976 ms
Planning Time: 0.991 ms
Execution Time: 3128.972 ms
```

## Решардинг
Citus "из коробки" умеет в решардинг, например, при добавлении в кластер нового хоста. Для этого, у распределённой таблицы должен быть задан первичный ключ или ограничение уникальности.
Допустим, у нас был только один шард "citus_worker_1":
```sql
SELECT nodename, count(*) FROM citus_shards GROUP BY nodename;
nodename      |count|
--------------+-----+
citus_worker_1|   32|
```
Действия следующие:
- Добавляем новый шард
```sql
SELECT * from citus_add_node('citus_worker_2', 5432);
```
- Меняем wal_level на кластере:
```sql
ALTER SYSTEM SET wal_level = logical;
SELECT run_command_on_workers('ALTER SYSTEM SET wal_level = logical');
SELECT run_command_on_workers('SHOW wal_level');
```
- Перезапускаем кластер
```bash
docker compose -f docker-compose.yaml restart citus citus_worker_1 citus_worker_2
```
- Запускаем решардинг
```sql
SELECT citus_rebalance_start();
SELECT * FROM citus_rebalance_status();
job_id|state  |job_type |description                    |started_at                   |finished_at
------+-------+---------+-------------------------------+-----------------------------+-----------
     1|running|rebalance|Rebalance all colocation groups|2024-09-19 01:55:25.285 +0300|           
...
job_id|state   |job_type |description                    |started_at                   |finished_at                  
------+--------+---------+-------------------------------+-----------------------------+-----------------------------
     1|finished|rebalance|Rebalance all colocation groups|2024-09-19 01:55:25.285 +0300|2024-09-19 01:56:06.533 +0300
```
Проверяем распределением шардов по нодам кластера:
```sql
SELECT nodename, count(*) FROM citus_shards GROUP BY nodename;
nodename      |count|
--------------+-----+
citus_worker_1|   18|
citus_worker_2|   14|
```
В итоге получаем перераспределённые данные между шардами без даунтайма в рамках обновлённого кластера.
