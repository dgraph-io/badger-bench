# Benchmarks for BadgerDB
This repo contains the code for benchmarking [BadgerDB], along with detailed logs from previous benchmarking runs.  We will update this repo to incorporate benchmarks for Pebble, BoltDB, and Pogreb.

[BadgerDB]:https://github.com/dgraph-io/badger

- Install badger bench

```
$ go get github.com/dgraph-io/badger-bench/...
```

- Run  `go test -c` and make sure everything compiles. Refer to the benchmarking logs below for commands to run individual benchmarks.

## Benchmarking Logs and Blog Posts
This repo is based on a series of benchmarks of Badger against RocksDB, BoltDB, and LMDB.
