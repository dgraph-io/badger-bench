# Benchmarks for Badger

# Provisioning AWS Instance for Benchmarking
- EC2 Instance Type: i3-large (use SSD option)
- RAM Size: 16G (to generate 32G of data)

# Setting Up
- Clone badger-bench repo

```
$ git clone https://github.com/dgraph-io/badger-bench
```

- Install rocksdb using steps here: https://github.com/facebook/rocksdb/blob/master/INSTALL.md

```
    $ make shared_lib
    $ sudo make install-shared
    $ ldconfig # To update the ld.so.cache
```

- Run  `go test` and make sure everything compiles.

