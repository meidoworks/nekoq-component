# Replication

## 1. Features

* [x] [WAL] Write
* [ ] [WAL] Write file rotation
    * With position rewind
* [ ] [WAL] Read from position roughly for scenarios including full replication
    * read the given range of wal logs
    * fault-tolerant beyond the range
* [ ] [WAL] Auto remove earliest wal files if total size exceeded
* [ ] [WAL] Page based write operation
* [x] [WAL] Read existing data file in startup process
* [ ] [WAL] handling data corruption in WAL log files when startup
    * [ ] Opt1. remove invalid record which applies for latest ones
    * [ ] Opt2. drop current WAL log files and retrieve new ones from other nodes
    * [ ] Opt3. manually handle data corruption
* [ ] [WAL] Safe shutdown and data integrity for unexpected shutdown

## 2. Usage(integration with a real database, a replication coordinator and a WAL logger)

### 2.1 Initialize WAL

TODO

### 2.2 Write operation

TODO

### 2.3 Full replication

TODO

### 2.4 Realtime replication

TODO
