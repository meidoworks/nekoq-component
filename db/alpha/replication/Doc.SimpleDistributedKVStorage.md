# Simple Distributed Replication KV Storage

## 1. Concepts

* leader: source of data
* replica: acceptor of data

## 2. Design

### 2.1 Full replication flow

1. replica: request full db copy
2. leader: respond full db copy
3. replica: build whole db from full db copy
4. replica: request wal position
5. leader: respond WAL logs from requested position to latest WAL position including a flag to inform the replica of
   starting service
6. replica: replay WAL

### 2.2 Incremental replication flow

1. replica: tell current WAL position
2. #1 leader: respond WAL logs from requested position to latest WAL position
3. #2 leader: respond full replication required message if requested WAL position does not exist

### 2.3 Realtime replication flow

1. do [2.1] in the beginning
2. record WAL position of replica
3. do [2.2] continuously
4. update WAL position of replica

Note:

1. [2.2] can be triggered by leaders, then leaders can decide the numbers of sync and async replicas; or it can be
   triggered by replicas, then the replication is always in async mode.

### 2.4 Data integrity guarantee

1. replicas must not start servicing before [2.1]; the reason is: full db copy may have incomplete data, which requires
   WAL log to correct. Note that there are a lot of mechanisms to ensure the full db copy is complete. But here it is
   not required.
2. in [2.2] and [2.3] stages, replicas can serve requests and allow stale data; the reason is: in these stages, only WAL
   log is updated.
3. The order of WAL log must be the same as db write operations

### 2.5 Performance optimization

1. For the system that separate db and wal mechanisms, in order to guarantee the order between WAL log and db write
   operations, write thread count should be 1
2. For the system that combines db and wal mechanisms, multiple write threads can be achieved by atomic page management
   by the db, which includes wal logs are organized by txns, page pool for parallel writes, MVCC, etc.
3. For simple kv system which does not support multiple write operations in a single txn, multiple write thread can be
   achieved using key hashing.

### 2.6 Schedule tasks

1. Write current WAL position in db periodically

### 2.7 Startup initialization flow

1. read WAL position from db
2. Compare db WAL position to WAL log position
3. If db pos > wal pos, raise error, data may be corrupted
4. If db pos = wal pos, continue to start service
5. If db pos < wal pos, replay WAL log from db pos until the latest WAL log, after that continue to start service

