# Benchmarks for BadgerDB
This repo contains the code for benchmarking [BadgerDB], along with detailed logs from previous benchmarking runs.

[BadgerDB]:https://github.com/dgraph-io/badger

- Install badger bench

```
$ go get github.com/dgraph-io/badger-bench/...
```

- Run  `go test -c` and make sure everything compiles. Refer to the benchmarking logs below for commands to run individual benchmarks.

## Benchmarking Logs and Blog Posts
We have performed comprehensive benchmarks against RocksDB, BoltDB and LMDB.
Detailed logs of all the steps are made available in this repo. Refer to the 
blog posts for graphs and other information.

* [Benchmarking log for RocksDB](https://github.com/dgraph-io/badger-bench/blob/master/BENCH-rocks.txt) (link to [blog post](https://blog.dgraph.io/post/badger/))
* [Benchmarking log for BoltDB and LMDB](https://github.com/dgraph-io/badger-bench/blob/master/BENCH-lmdb-bolt.md) (link to [blog post](https://blog.dgraph.io/post/badger-lmdb-boltdb/))

