# Benchmarks for Badger

# Provisioning AWS Instance for Benchmarking
- EC2 Instance Type: i3-large (use SSD option)
- RAM Size: 16G (to generate 32G of data)

# Setting Up
- Install rocksdb using steps here: https://github.com/facebook/rocksdb/blob/master/INSTALL.md

```
$ sudo apt-get update && sudo apt-get install libgflags-dev libsnappy-dev zlib1g-dev libbz2-dev liblz4-dev libzstd-dev
$ wget https://github.com/facebook/rocksdb/archive/v5.1.4.tar.gz
$ tar -xzvf v5.1.4.tar.gz
$ cd rocksdb-5.1.4
$ make shared_lib
$ sudo make install
$ ldconfig # to update ld.so.cache
```

- Install badger bench

```
$ go get github.com/dgraph-io/badger-bench
```

- Run  `go test` and make sure everything compiles.

