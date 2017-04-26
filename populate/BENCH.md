## Badger: 128 byte value

```
Command being timed: "./populate --kv badger --keys_mil 150 --valsz 128"
User time (seconds): 3576.44
System time (seconds): 431.65
Percent of CPU this job got: 249%
Elapsed (wall clock) time (h:mm:ss or m:ss): 26:47.03
Average shared text size (kbytes): 0
Average unshared data size (kbytes): 0
Average stack size (kbytes): 0
Average total size (kbytes): 0
Maximum resident set size (kbytes): 7543144
Average resident set size (kbytes): 0
Major (requiring I/O) page faults: 6
Minor (reclaiming a frame) page faults: 110987357
Voluntary context switches: 5470444
Involuntary context switches: 964837
Swaps: 0
File system inputs: 11862768
File system outputs: 174310312
Socket messages sent: 0
Socket messages received: 0
Signals delivered: 0
Page size (bytes): 4096
Exit status: 0
```

```
$ du -shc *
24G     total
# 2.9G for LSM tree (.sst files), and 21G for value log.
```

## RocksDB: 128 byte value

```
Command being timed: "./populate --kv rocksdb --keys_mil 150 --valsz 128"
User time (seconds): 1305.44
System time (seconds): 257.49
Percent of CPU this job got: 102%
Elapsed (wall clock) time (h:mm:ss or m:ss): 25:29.99
Average shared text size (kbytes): 0
Average unshared data size (kbytes): 0
Average stack size (kbytes): 0
Average total size (kbytes): 0
Maximum resident set size (kbytes): 479432
Average resident set size (kbytes): 0
Major (requiring I/O) page faults: 31
Minor (reclaiming a frame) page faults: 3728581
Voluntary context switches: 10066415
Involuntary context switches: 1577901
Swaps: 0
File system inputs: 6081600
File system outputs: 278045328
Socket messages sent: 0
Socket messages received: 0
Signals delivered: 0
Page size (bytes): 4096
Exit status: 0
```

```
du -shc *
15G
```

