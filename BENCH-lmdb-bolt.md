# lmdb-go vs BoltDB vs Badger

- Badger benchmarking repo: https://github.com/dgraph-io/badger-bench
- lmdb-go: https://github.com/bmatsuo/lmdb-go
- BoltDB: https://github.com/boltdb/bolt
# Coding
## lmdb-go
- lmdb-go docs: https://godoc.org/github.com/bmatsuo/lmdb-go/lmdb


- lmdb does not have a method to do a batched write explicitly. We benchmarked two different ways to do batched writes: one with an explicit txn and another without. The benchmark code for that is in `populate/lmdb_txn_bench_test.go`. 


- Benchmark results for batched writes on Lenovo Thinkpad T470:
    $ go test --bench BenchmarkLmdbBatch --dir /tmp --benchtime 30s
    BenchmarkLmdbBatch/SimpleBatched-4                 20000           2069261 ns/op
    BenchmarkLmdbBatch/TxnBatched-4                    20000           2114079 ns/op
    PASS
    ok      github.com/dgraph-io/badger-bench/populate      125.410s


    $ go test --bench BenchmarkLmdbBatch --dir /tmp --benchtime 45s
    BenchmarkLmdbBatch/SimpleBatched-4                 30000           2037230 ns/op
    BenchmarkLmdbBatch/TxnBatched-4                    30000           2147498 ns/op
    PASS
    ok      github.com/dgraph-io/badger-bench/populate      167.050s

Based on results above, we can conclude that batched update inside transactions is slightly more expensive.


- lmdb needs to lock the goroutine to a single OS thread at runtime:
  > Write transactions (those created without the Readonly flag) must be created in a goroutine that has been locked to its thread by calling the function `runtime.LockOSThread`. Futhermore, all methods on such transactions must be called from the goroutine which created them. This is a fundamental limitation of LMDB even when using the NoTLS flag (which the package always uses). The `Env.Update` method assists the programmer by calling runtime.LockOSThread automatically but it cannot sufficiently abstract write transactions to make them completely safe in Go.
## BoltDB
- Pull request to badger-bench repo: https://github.com/dgraph-io/badger-bench/pull/7
- Used `NoSync` option while writing to avoid fsync after every commit. Need to determine if it actually helps
- After a bit of testing, we switched to `bolt.DB.Batch` method to populate the data instead of `bolt.DB.Update`.
# Launching AWS Instance to benchmark
- AMI: **Ubuntu Server 16.04 LTS (HVM), SSD Volume Type** - *ami-10547475*


- Instance Type: **i3-large**
![2 vCPUs, 15.25Gb, 1x475 (SSD)](https://d2mxuefqeaa7sj.cloudfront.net/s_CE68978A348E19B7DD1520A31AD4737F14F8B2D2704BBCAA008EA13523642F20_1502103139556_Screenshot+from+2017-08-07+16-21-06.png)

- Additional instance details: Dedicated instance
![](https://d2mxuefqeaa7sj.cloudfront.net/s_CE68978A348E19B7DD1520A31AD4737F14F8B2D2704BBCAA008EA13523642F20_1502103250117_Screenshot+from+2017-08-07+16-23-30.png)

- Storage Details
![Why does it say 8GiB?](https://d2mxuefqeaa7sj.cloudfront.net/s_CE68978A348E19B7DD1520A31AD4737F14F8B2D2704BBCAA008EA13523642F20_1502103505922_Screenshot+from+2017-08-07+16-28-04.png)

# Setting Up Instance
- Make sure SSD instance is available: `lsblk`
    NAME    MAJ:MIN RM   SIZE RO TYPE MOUNTPOINT
    xvda    202:0    0     8G  0 disk
    └─xvda1 202:1    0     8G  0 part /
    nvme0n1 259:0    0 442.4G  0 disk <------------
    


- Format and mount SSD instance
    $ sudo mkdir /mnt/data
    $ sudo mkfs -t ext4 /dev/nvme0n1
    $ sudo mount -t ext4 /dev/nvme0n1  /mnt/data/
    $ sudo chmod 777 /mnt/data
    $ df -h
    Filesystem      Size  Used Avail Use% Mounted on
    udev            7.5G     0  7.5G   0% /dev
    tmpfs           1.5G  8.6M  1.5G   1% /run
    /dev/xvda1      7.7G  2.0G  5.8G  26% /
    tmpfs           7.5G     0  7.5G   0% /dev/shm
    tmpfs           5.0M     0  5.0M   0% /run/lock
    tmpfs           7.5G     0  7.5G   0% /sys/fs/cgroup
    tmpfs           1.5G     0  1.5G   0% /run/user/1000
    /dev/nvme0n1    436G   71M  414G   1% /mnt/data   <---------------


- Put this in `~/.bashrc`
    export GOMAXPROCS=128


- Create a file `touch ~/drop_caches.sh && chmod +x ~/drop_caches.sh`
    #!/bin/sh
    echo 3 | sudo tee /proc/sys/vm/drop_caches
    sudo blockdev --flushbufs  /dev/nvme0n1


- Install build-essentials: `sudo apt-get install build-essential`


- Launch a new screen session and do everything in that to avoid any problems due to disconnection: `screen -S bench`


- Install rocksdb using steps here: https://github.com/facebook/rocksdb/blob/master/INSTALL.md
    $ sudo apt-get update && sudo apt-get install libgflags-dev libsnappy-dev zlib1g-dev libbz2-dev liblz4-dev libzstd-dev
    $ wget https://github.com/facebook/rocksdb/archive/v5.1.4.tar.gz
    $ tar -xzvf v5.1.4.tar.gz
    $ cd rocksdb-5.1.4
    $ make shared_lib
    $ sudo make install-shared
    $ sudo ldconfig # to update ld.so.cache


- Install Go 1.8: https://github.com/golang/go/wiki/Ubuntu


- Install badger bench (This will also pull in the lmdb-go package, along with the lmdb C code and install it)
    $ go get -t -v github.com/dgraph-io/badger-bench/...


-  `cd ~/go/src/github.com/dgraph-io/badger-bench/`


- Run `go test` and make sure rocksdb is linked up4
# Benchmarking
- Install `sar` to monitor disk activity: `sudo apt install sysstat`
- Always remember to use the mounted SSD disk to do writes. The home directory is mounted on EBS and gives very little IOPS.
## Rerunning BenchmarkLmdbBatch

We rerun the lmdb batched write benchmarks above, on this instance

    ubuntu@ip-172-31-39-80:~/go/src/github.com/dgraph-io/badger-bench/populate$ go test --bench BenchmarkLmdbBatch --dir /mnt/data --benchtime 30s
    BenchmarkLmdbBatch/SimpleBatched-2                200000            321828 ns/op
    BenchmarkLmdbBatch/TxnBatched-2                   200000            324301 ns/op
    PASS
    ok      github.com/dgraph-io/badger-bench/populate      135.817s



    ubuntu@ip-172-31-39-80:~/go/src/github.com/dgraph-io/badger-bench/populate$ go test --bench BenchmarkLmdbBatch --dir /mnt/data --benchtime 45s
    BenchmarkLmdbBatch/SimpleBatched-2                200000            329147 ns/op
    BenchmarkLmdbBatch/TxnBatched-2                   200000            327282 ns/op
    PASS
    ok      github.com/dgraph-io/badger-bench/populate      138.006s


## Running `fio` to get baseline IOPS numbers

Make sure you change into a directory on the mounted SSD drive.

As you can see below, this instance gives about **94k iops** at **4k block size**.

    $ sudo apt-get install fio
    $ mkdir /mnt/data/fio
    $ cd /mnt/data/fio
    $ fio --name=randread --ioengine=libaio --iodepth=32 --rw=randread --bs=4k --direct=0 --size=2G --numjobs=16 --runtime=240 --group_reporting
    
    randread: (g=0): rw=randread, bs=4K-4K/4K-4K/4K-4K, ioengine=libaio, iodepth=32
    ...
    fio-2.2.10
    Starting 16 processes
    randread: Laying out IO file(s) (1 file(s) / 2048MB)
    randread: Laying out IO file(s) (1 file(s) / 2048MB)
    randread: Laying out IO file(s) (1 file(s) / 2048MB)
    randread: Laying out IO file(s) (1 file(s) / 2048MB)
    randread: Laying out IO file(s) (1 file(s) / 2048MB)
    randread: Laying out IO file(s) (1 file(s) / 2048MB)
    randread: Laying out IO file(s) (1 file(s) / 2048MB)
    randread: Laying out IO file(s) (1 file(s) / 2048MB)
    randread: Laying out IO file(s) (1 file(s) / 2048MB)
    randread: Laying out IO file(s) (1 file(s) / 2048MB)
    randread: Laying out IO file(s) (1 file(s) / 2048MB)
    randread: Laying out IO file(s) (1 file(s) / 2048MB)
    randread: Laying out IO file(s) (1 file(s) / 2048MB)
    randread: Laying out IO file(s) (1 file(s) / 2048MB)
    randread: Laying out IO file(s) (1 file(s) / 2048MB)
    randread: Laying out IO file(s) (1 file(s) / 2048MB)
    
    …<snip>…
    
    randread: (groupid=0, jobs=16): err= 0: pid=25740: Tue Aug  8 03:42:24 2017
      read : io=32768MB, bw=378842KB/s, iops=94710, runt= 88571msec
        slat (usec): min=40, max=37496, avg=159.80, stdev=172.96
        clat (usec): min=1, max=41205, avg=5064.23, stdev=1834.76
         lat (usec): min=93, max=41350, avg=5224.60, stdev=1873.67
        clat percentiles (usec):
         |  1.00th=[ 3152],  5.00th=[ 3280], 10.00th=[ 3440], 20.00th=[ 3632],
         | 30.00th=[ 3824], 40.00th=[ 4016], 50.00th=[ 4384], 60.00th=[ 4896],
         | 70.00th=[ 5536], 80.00th=[ 6432], 90.00th=[ 7648], 95.00th=[ 8768],
         | 99.00th=[11072], 99.50th=[11968], 99.90th=[14144], 99.95th=[15424],
         | 99.99th=[21120]
        bw (KB  /s): min=19224, max=38464, per=6.46%, avg=24472.97, stdev=3100.21
        lat (usec) : 2=0.01%, 4=0.01%, 10=0.01%, 20=0.01%, 100=0.01%
        lat (usec) : 250=0.01%, 500=0.01%, 750=0.01%, 1000=0.01%
        lat (msec) : 2=0.01%, 4=38.81%, 10=59.00%, 20=2.17%, 50=0.01%
      cpu          : usr=2.42%, sys=6.36%, ctx=8463022, majf=0, minf=652
      IO depths    : 1=0.1%, 2=0.1%, 4=0.1%, 8=0.1%, 16=0.1%, 32=100.0%, >=64=0.0%
         submit    : 0=0.0%, 4=100.0%, 8=0.0%, 16=0.0%, 32=0.0%, 64=0.0%, >=64=0.0%
         complete  : 0=0.0%, 4=100.0%, 8=0.0%, 16=0.0%, 32=0.1%, 64=0.0%, >=64=0.0%
         issued    : total=r=8388608/w=0/d=0, short=r=0/w=0/d=0, drop=r=0/w=0/d=0
         latency   : target=0, window=0, percentile=100.00%, depth=32
    
    Run status group 0 (all jobs):
       READ: io=32768MB, aggrb=378842KB/s, minb=378842KB/s, maxb=378842KB/s, mint=88571msec, maxt=88571msec
    
    Disk stats (read/write):
      nvme0n1: ios=8387535/30, merge=0/0, ticks=822124/0, in_queue=825008, util=100.00%
    
## Benchmarking `--keys_mil 5` and `--valsz 16384`

**Time population of lmdb** `**--valsz 16384**` **and** `**--keys_mil 5**`

    $ /usr/bin/time -v ./populate --kv lmdb --valsz 16384 --keys_mil 5 --dir /mnt/data/16kb
    TOTAL KEYS TO WRITE:   5.00M
    Init lmdb
    [0000] Write key rate per minute:  11.00K. Total:  11.00K
    [0001] Write key rate per minute:  30.00K. Total:  30.00K
    [0002] Write key rate per minute:  45.00K. Total:  45.00K
    [0003] Write key rate per minute:  62.00K. Total:  62.00K
    …<snip>…
    [0685] Write key rate per minute: 332.00K. Total:   4.98M
    [0686] Write key rate per minute: 338.00K. Total:   4.98M
    [0687] Write key rate per minute: 326.00K. Total:   4.99M
    [0688] Write key rate per minute: 332.00K. Total:   4.99M
    [0689] Write key rate per minute: 337.00K. Total:   5.00M
    closing lmdb
    
    WROTE 5004000 KEYS
            Command being timed: "./populate --kv lmdb --valsz 16384 --keys_mil 5 --dir /mnt/data/16kb"
            User time (seconds): 344.08
            System time (seconds): 172.82
            Percent of CPU this job got: 74%
            Elapsed (wall clock) time (h:mm:ss or m:ss): 11:31.16
            Average shared text size (kbytes): 0
            Average unshared data size (kbytes): 0
            Average stack size (kbytes): 0
            Average total size (kbytes): 0
            Maximum resident set size (kbytes): 10933704
            Average resident set size (kbytes): 0
            Major (requiring I/O) page faults: 1166660
            Minor (reclaiming a frame) page faults: 438898
            Voluntary context switches: 3321574
            Involuntary context switches: 249619
            Swaps: 0
            File system inputs: 257987728
            File system outputs: 250540080
            Socket messages sent: 0
            Socket messages received: 0
            Signals delivered: 0
            Page size (bytes): 4096
            Exit status: 0
    


    $ du -sh /mnt/data/16kb/lmdb/
    61G     /mnt/data/16kb/lmdb/


    $ sar -d 1 -p
    Average:          DEV       tps  rd_sec/s  wr_sec/s  avgrq-sz  avgqu-sz     await     svctm     %util
    Average:      nvme0n1   6223.99 386668.41 193847.54     93.27    483.51     78.20      0.14     86.43
    

**Time population of BoltDB** `**--valsz 16384**` **and** `**--keys_mil 5**`

    $ /usr/bin/time -v ./populate --kv bolt --valsz 16384 --keys_mil 5 --dir /mnt/data/16kb
    TOTAL KEYS TO WRITE:   5.00M
    Init BoldDB
    [0000] Write key rate per minute:   6.00K. Total:   6.00K
    [0001] Write key rate per minute:  10.00K. Total:  10.00K
    [0002] Write key rate per minute:  13.00K. Total:  13.00K
    …<snip>…
    [10651] Write key rate per minute:  19.00K. Total:   5.00M
    [10652] Write key rate per minute:  20.00K. Total:   5.00M
    [10653] Write key rate per minute:  19.00K. Total:   5.00M
    closing bolt
    
    WROTE 5004000 KEYS
            Command being timed: "./populate --kv bolt --valsz 16384 --keys_mil 5 --dir /mnt/data/16kb"
            User time (seconds): 5719.37
            System time (seconds): 307.54
            Percent of CPU this job got: 65%
            Elapsed (wall clock) time (h:mm:ss or m:ss): 2:34:14
            Average shared text size (kbytes): 0
            Average unshared data size (kbytes): 0
            Average stack size (kbytes): 0
            Average total size (kbytes): 0
            Maximum resident set size (kbytes): 8705716
            Average resident set size (kbytes): 0
            Major (requiring I/O) page faults: 28118500
            Minor (reclaiming a frame) page faults: 3582686
            Voluntary context switches: 31597388
            Involuntary context switches: 67596
            Swaps: 0
            File system inputs: 224969160
            File system outputs: 707677000
            Socket messages sent: 0
            Socket messages received: 0
            Signals delivered: 0
            Page size (bytes): 4096
            Exit status: 0


    $ du -sh /mnt/data/16kb/bolt/
    55G     /mnt/data/16kb/bolt/


    $ sar -d 1 -p
    Average:          DEV       tps  rd_sec/s  wr_sec/s  avgrq-sz  avgqu-sz     await     svctm     %util
    Average:      nvme0n1   3607.69      0.22 326872.54     90.60    221.42     61.35      0.08     29.58
    


    Average:          DEV       tps  rd_sec/s  wr_sec/s  avgrq-sz  avgqu-sz     await     svctm     %util
    Average:      nvme0n1   3854.80  21760.53  86716.40     28.14     50.86     13.19      0.14     52.94
    


    Average:          DEV       tps  rd_sec/s  wr_sec/s  avgrq-sz  avgqu-sz     await     svctm     %util
    Average:      nvme0n1   3466.31  22565.38  44105.02     19.23     29.25      8.44      0.14     46.89
    

**Time population of badger** `**--valsz 16384**` **and** `**--keys_mil 5**`

    $ /usr/bin/time -v ./populate --kv badger --valsz 16384 --keys_mil 5 --dir /mnt/data/16kb
    TOTAL KEYS TO WRITE:   5.00M
    Init Badger
    [0000] Write key rate per minute:   1.00K. Total:   1.00K
    [0001] Write key rate per minute:  24.00K. Total:  24.00K
    [0002] Write key rate per minute:  36.00K. Total:  36.00K
    …<snip>…
    [0403] Write key rate per minute: 704.00K. Total:   4.96M
    [0404] Write key rate per minute: 716.00K. Total:   4.98M
    [0405] Write key rate per minute: 684.00K. Total:   4.99M
    closing badger
    2
    WROTE 5004000 KEYS
            Command being timed: "./populate --kv badger --valsz 16384 --keys_mil 5 --dir /mnt/data/16kb"
            User time (seconds): 367.25
            System time (seconds): 71.29
            Percent of CPU this job got: 106%
            Elapsed (wall clock) time (h:mm:ss or m:ss): 6:52.06
            Average shared text size (kbytes): 0
            Average unshared data size (kbytes): 0
            Average stack size (kbytes): 0
            Average total size (kbytes): 0
            Maximum resident set size (kbytes): 1347180
            Average resident set size (kbytes): 0
            Major (requiring I/O) page faults: 103
            Minor (reclaiming a frame) page faults: 205372
            Voluntary context switches: 2670070
            Involuntary context switches: 804367
            Swaps: 0
            File system inputs: 25112
            File system outputs: 160891728
            Socket messages sent: 0
            Socket messages received: 0
            Signals delivered: 0
            Page size (bytes): 4096
            Exit status: 0
    



    $ du -sh /mnt/data/16kb/badger/
    77G     /mnt/data/16kb/badger/


**Time random read for lmdb** `**--keys_mil 5**` ****`**--valsz 16384**`
Performed by opening the DB with `lmdb.NoReadahead` option.

    go test -v --bench BenchmarkReadRandomLmdb --keys_mil 5 --valsz 16384 --dir /mnt/data/16kb -benchtime 3m --timeout 10m                                                       
    BenchmarkReadRandomLmdb/read-randomlmdb-128             100000000             2015 ns/op
    --- BENCH: BenchmarkReadRandomLmdb
            bench_test.go:104: lmdb: 63257263 keys had valid values.
            bench_test.go:105: lmdb: 36742737 keys had no values
            bench_test.go:106: lmdb: 0 keys had errors
            bench_test.go:107: lmdb: 100000000 total keys looked at
            bench_test.go:108: lmdb: hit rate : 0.63
    PASS
    ok      github.com/dgraph-io/badger-bench       311.542s

**Time random read for boltdb** `**--keys_mil 5**` ****`**--valsz 16384**`

    $ go test -v --bench BenchmarkReadRandomBolt --keys_mil 5 --valsz 16384 --dir "/mnt/data/16kb" --timeout 10m --benchtime 3m
    BenchmarkReadRandomBolt/read-randombolt-128             100000000             2126 ns/op
    --- BENCH: BenchmarkReadRandomBolt
            bench_test.go:104: bolt: 63234082 keys had valid values.
            bench_test.go:105: bolt: 36765918 keys had no values
            bench_test.go:106: bolt: 0 keys had errors
            bench_test.go:107: bolt: 100000000 total keys looked at
            bench_test.go:108: bolt: hit rate : 0.63
    PASS
    ok      github.com/dgraph-io/badger-bench       216.550s

**Time random read for badger** `**--keys_mil 5**` ****`**--valsz 16384**`

    $ go test -v --bench BenchmarkReadRandomBadger --keys_mil 5 --valsz 16384 --dir "/mnt/data/16kb" --timeout 10m --benchtime 3m
    BenchmarkReadRandomBadger/read-randombadger-128                 50000000             3845 ns/op
    --- BENCH: BenchmarkReadRandomBadger
            bench_test.go:104: badger: 31619437 keys had valid values.
            bench_test.go:105: badger: 18380563 keys had no values
            bench_test.go:106: badger: 0 keys had errors
            bench_test.go:107: badger: 50000000 total keys looked at
            bench_test.go:108: badger: hit rate : 0.63
    PASS
    ok      github.com/dgraph-io/badger-bench       339.063s

**Time iterate for lmdb** `**--keys_mil 5**` ****`**--valsz 16384**`

    $ go test -v --bench BenchmarkIterateLmdb --keys_mil 5 --valsz 16384 --dir "/mnt/data/16kb" --timeout 60m
    ....................
    BenchmarkIterateLmdb/lmdb-iterate-128                      1    488071445140 ns/op
    --- BENCH: BenchmarkIterateLmdb/lmdb-iterate-128
            bench_test.go:275: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       488.454s
    
    

w/ `lmdb.NoReadahead`


    $ go test -v --bench BenchmarkIterateLmdb --keys_mil 5 --valsz 16384 --dir "/mnt/data/16kb" --timeout 60m
    ....................BenchmarkIterateLmdb/lmdb-iterate-128                      1
            2745882237254 ns/op
    --- BENCH: BenchmarkIterateLmdb/lmdb-iterate-128
            bench_test.go:317: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       2746.396s

**Time iterate for boltdb** `**--keys_mil 5**` ****`**--valsz 16384**`
There is a lot of variability in this. Recorded runs of `1122s` and `1329s` as well

    $ go test -v --bench BenchmarkIterateBolt --keys_mil 5 --valsz 16384 --dir "/mnt/data/16kb" --timeout 60m
    ....................BenchmarkIterateBolt/boltdb-iterate-128                    1
            930340179734 ns/op
    --- BENCH: BenchmarkIterateBolt/boltdb-iterate-128
            bench_test.go:353: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       930.774s

**Time iterate for badger (with values)** `**--keys_mil 5**` ****`**--valsz 16384**`

    $ go test -v --bench BenchmarkIterateBadgerWithValues --keys_mil 5 --valsz 16384 --dir "/mnt/data/16kb" --timeout 60m
    ....................BenchmarkIterateBadgerWithValues/badger-iterate-withvals-128                       1        93944919433 ns/op
    --- BENCH: BenchmarkIterateBadgerWithValues/badger-iterate-withvals-128
            bench_test.go:433: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       94.902s
    
    $ go test -v --bench BenchmarkIterateBadgerWithValues --keys_mil 5 --valsz 16384 --dir "/mnt/data/16kb" --timeout 60m
    ....................BenchmarkIterateBadgerWithValues/badger-iterate-withvals-128                       1        84715326781 ns/op
    --- BENCH: BenchmarkIterateBadgerWithValues/badger-iterate-withvals-128
            bench_test.go:433: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       85.647s

**Time iterate for badger (keys only)** `**--keys_mil 5**` ****`**--valsz 16384**`

    $ go test -v --bench BenchmarkIterateBadgerOnlyKeys --keys_mil 5 --valsz 16384 --dir "/mnt/data/16kb" --timeout 60m
    
    ....................BenchmarkIterateBadgerOnlyKeys/badger-iterate-onlykeys-128          ........................................       2         789184586 ns/op
    --- BENCH: BenchmarkIterateBadgerOnlyKeys/badger-iterate-onlykeys-128
            bench_test.go:394: [0] Counted 2000000 keys
            bench_test.go:394: [0] Counted 2000000 keys
            bench_test.go:394: [1] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       2.730s


## Benchmarking `--keys_mil 75` and `--valsz 1024`

**Time population of lmdb** `**--valsz 1024**` **and** `**--keys_mil 75**`
First run, we accidentally ran into an issue because we exceeded the MapSize setting for the database (we had set it to 64gb). The populate run terminated prematurely

    $ /usr/bin/time -v ./populate --kv lmdb --valsz 1024 --keys_mil 75 --dir /mnt/data/1kb
    TOTAL KEYS TO WRITE:  75.00M
    Init lmdb
    [0000] Write key rate per minute:  43.00K. Total:  43.00K
    [0001] Write key rate per minute:  70.00K. Total:  70.00K
    [0002] Write key rate per minute:  96.00K. Total:  96.00K
    …<snip>…
    [7521] Write key rate per minute: 215.00K. Total:  43.77M
    [7522] Write key rate per minute: 219.00K. Total:  43.78M
    [7523] Write key rate per minute: 223.00K. Total:  43.78M
    2017/08/08 08:57:17 mdb_put: MDB_MAP_FULL: Environment mapsize limit reached
    
    github.com/dgraph-io/badger/y.Wrap
            /home/ubuntu/go/src/github.com/dgraph-io/badger/y/error.go:71
    github.com/dgraph-io/badger/y.Check
            /home/ubuntu/go/src/github.com/dgraph-io/badger/y/error.go:43
    main.writeBatch
            /home/ubuntu/go/src/github.com/dgraph-io/badger-bench/populate/main.go:86
    main.main.func5
            /home/ubuntu/go/src/github.com/dgraph-io/badger-bench/populate/main.go:221
    runtime.goexit
            /usr/lib/go-1.8/src/runtime/asm_amd64.s:2197
    Command exited with non-zero status 1
            Command being timed: "./populate --kv lmdb --valsz 1024 --keys_mil 75 --dir /mnt/data/1kb"
            User time (seconds): 558.00
            System time (seconds): 1575.94
            Percent of CPU this job got: 28%
            Elapsed (wall clock) time (h:mm:ss or m:ss): 2:05:25
            Average shared text size (kbytes): 0
            Average unshared data size (kbytes): 0
            Average stack size (kbytes): 0
            Average total size (kbytes): 0
            Maximum resident set size (kbytes): 14542052
            Average resident set size (kbytes): 0
            Major (requiring I/O) page faults: 23234347
            Minor (reclaiming a frame) page faults: 4794003
            Voluntary context switches: 37399045
            Involuntary context switches: 302805
            Swaps: 0
            File system inputs: 4911778792
            File system outputs: 1046433264
            Socket messages sent: 0
            Socket messages received: 0
            Signals delivered: 0
            Page size (bytes): 4096
            Exit status: 1
    


    $ sar -d 1 -p
    …<snip>…
    
    Average:          DEV       tps  rd_sec/s  wr_sec/s  avgrq-sz  avgqu-sz     await     svctm     %util
    Average:      nvme0n1  24231.12 665881.00 146593.00     33.53     54.90      2.27      0.04     90.30

Ran the benchmark again after setting the map size to about 270Gb

Noticed that the Write key rate steadily dropped over time, after starting at close to a  1M keys a minute to a steady 200K keys a minute.

    /usr/bin/time -v ./populate --kv lmdb --valsz 1024 --keys_mil 75 --dir /mnt/data/1kb
    TOTAL KEYS TO WRITE:  75.00M
    Init lmdb
    [0000] Write key rate per minute:  52.00K. Total:  52.00K
    [0001] Write key rate per minute:  79.00K. Total:  79.00K
    [0002] Write key rate per minute: 104.00K. Total: 104.00K
    …<snip>…
    [16672] Write key rate per minute: 189.00K. Total:  74.99M
    [16673] Write key rate per minute: 192.00K. Total:  75.00M
    [16674] Write key rate per minute: 185.00K. Total:  75.00M
    closing lmdb
    
    WROTE 75000000 KEYS
            Command being timed: "./populate --kv lmdb --valsz 1024 --keys_mil 75 --dir /mnt/data/1kb"
            User time (seconds): 1010.87
            System time (seconds): 3288.08
            Percent of CPU this job got: 25%
            Elapsed (wall clock) time (h:mm:ss or m:ss): 4:37:55
            Average shared text size (kbytes): 0
            Average unshared data size (kbytes): 0
            Average stack size (kbytes): 0
            Average total size (kbytes): 0
            Maximum resident set size (kbytes): 14641980
            Average resident set size (kbytes): 0
            Major (requiring I/O) page faults: 52980386
            Minor (reclaiming a frame) page faults: 7319073
            Voluntary context switches: 78258475
            Involuntary context switches: 391508
            Swaps: 0
            File system inputs: 11860339200
            File system outputs: 1815854784
            Socket messages sent: 0
            Socket messages received: 0
            Signals delivered: 0
            Page size (bytes): 4096
            Exit status: 0


    $ sar -d 1 -p
    …<snip>…
    
    Average:          DEV       tps  rd_sec/s  wr_sec/s  avgrq-sz  avgqu-sz     await     svctm     %util
    Average:      nvme0n1  18721.60 722433.96 108669.16     44.39     41.09      2.19      0.05     91.24
    


    $ du -sh /mnt/data/1kb/lmdb/
    92G     /mnt/data/1kb/lmdb/

**Time population of boltdb** `**--valsz 1024**` **and** `**--keys_mil 75**`

    $ /usr/bin/time -v ./populate --kv bolt --valsz 1024 --keys_mil 75 --dir /mnt/data/1kb
    TOTAL KEYS TO WRITE:  75.00M
    Init BoldDB
    [0000] Write key rate per minute:  47.00K. Total:  47.00K
    [0001] Write key rate per minute:  79.00K. Total:  79.00K
    [0002] Write key rate per minute: 109.00K. Total: 109.00K
    …<snip>…
    [12968] Write key rate per minute: 279.00K. Total:  74.98M
    [12969] Write key rate per minute: 272.00K. Total:  74.99M
    [12970] Write key rate per minute: 278.00K. Total:  75.00M
    closing bolt
    
    WROTE 75000000 KEYS
            Command being timed: "./populate --kv bolt --valsz 1024 --keys_mil 75 --dir /mnt/data/1kb"
            User time (seconds): 4161.84
            System time (seconds): 1518.50
            Percent of CPU this job got: 43%
            Elapsed (wall clock) time (h:mm:ss or m:ss): 3:36:12
            Average shared text size (kbytes): 0
            Average unshared data size (kbytes): 0
            Average stack size (kbytes): 0
            Average total size (kbytes): 0
            Maximum resident set size (kbytes): 14268780
            Average resident set size (kbytes): 0
            Major (requiring I/O) page faults: 57249993
            Minor (reclaiming a frame) page faults: 16601016
            Voluntary context switches: 74487720
            Involuntary context switches: 397177
            Swaps: 0
            File system inputs: 458025016
            File system outputs: 1605560432
            Socket messages sent: 0
            Socket messages received: 0
            Signals delivered: 0
            Page size (bytes): 4096
            Exit status: 0


    $ du -sh /mnt/data/1kb/
    79G     /mnt/data/1kb/


    $ sar -d 1 -p
    Average:          DEV       tps  rd_sec/s  wr_sec/s  avgrq-sz  avgqu-sz     await     svctm     %util
    Average:      nvme0n1  23556.75      0.13 278911.16     11.84    533.38     22.65      0.02     48.21
    

**Time population of badger** `**--valsz 1024**` **and** `**--keys_mil 75**`

    $ /usr/bin/time -v ./populate --kv badger --valsz 1024 --keys_mil 75 --dir /mnt/data/1kb
    TOTAL KEYS TO WRITE:  75.00M
    Init Badger
    [0000] Write key rate per minute: 157.00K. Total: 157.00K
    [0001] Write key rate per minute: 311.00K. Total: 311.00K
    [0002] Write key rate per minute: 437.00K. Total: 437.00K
    …<snip>…
    [0926] Write key rate per minute:   4.08M. Total:  74.79M
    [0927] Write key rate per minute:   3.88M. Total:  74.81M
    [0928] Write key rate per minute:   4.00M. Total:  74.93M
    closing badger
    
    WROTE 75000000 KEYS
            Command being timed: "./populate --kv badger --valsz 1024 --keys_mil 75 --dir /mnt/data/1kb"
            User time (seconds): 1171.72
            System time (seconds): 115.08
            Percent of CPU this job got: 138%
            Elapsed (wall clock) time (h:mm:ss or m:ss): 15:31.28
            Average shared text size (kbytes): 0
            Average unshared data size (kbytes): 0
            Average stack size (kbytes): 0
            Average total size (kbytes): 0
            Maximum resident set size (kbytes): 10120620
            Average resident set size (kbytes): 0
            Major (requiring I/O) page faults: 2
            Minor (reclaiming a frame) page faults: 2307465
            Voluntary context switches: 2455400
            Involuntary context switches: 842949
            Swaps: 0
            File system inputs: 686120
            File system outputs: 181301424
            Socket messages sent: 0
            Socket messages received: 0
            Signals delivered: 0
            Page size (bytes): 4096
            Exit status: 0
    



    $ sar -d 1 -p
    …<snip>…
    
    Average:          DEV       tps  rd_sec/s  wr_sec/s  avgrq-sz  avgqu-sz     await     svctm     %util
    Average:      nvme0n1    743.36      6.40 188613.76    253.74    136.94    184.85      0.37     27.79
    


    $ du -sh /mnt/data/1kb/badger/
    77G     /mnt/data/1kb/badger/cd ..


**Time random read for lmdb** `**--keys_mil 75**` ****`**--valsz 1024**`

    $ go test -v --bench BenchmarkReadRandomLmdb --keys_mil 75 --valsz 1024 --dir "/mnt/data/1kb" --timeout 10m --benchtime 3m
    BenchmarkReadRandomLmdb/read-randomlmdb-128             20000000             10710 ns/op
    --- BENCH: BenchmarkReadRandomLmdb
            bench_test.go:108: lmdb: 12643872 keys had valid values.
            bench_test.go:109: lmdb: 7356128 keys had no values
            bench_test.go:110: lmdb: 0 keys had errors
            bench_test.go:111: lmdb: 20000000 total keys looked at
            bench_test.go:112: lmdb: hit rate : 0.63
    PASS
    ok      github.com/dgraph-io/badger-bench       226.968s

**Time random read for boltdb** `**--keys_mil 75**` ****`**--valsz 1024**`

    $ go test -v --bench BenchmarkReadRandomBolt --keys_mil 75 --valsz 1024 --dir "/mnt/data/1kb" --timeout 10m --benchtime 3m
    BenchmarkReadRandomBolt/read-randombolt-128             20000000             12205 ns/op
    --- BENCH: BenchmarkReadRandomBolt
            bench_test.go:104: bolt: 12643848 keys had valid values.
            bench_test.go:105: bolt: 7356152 keys had no values
            bench_test.go:106: bolt: 0 keys had errors
            bench_test.go:107: bolt: 20000000 total keys looked at
            bench_test.go:108: bolt: hit rate : 0.63
    PASS
    ok      github.com/dgraph-io/badger-bench       260.661s

**Time random read for badger** `**--keys_mil 75**` ****`**--valsz 1024**`

    $ go test -v --bench BenchmarkReadRandomBadger --keys_mil 75 --valsz 1024 --dir "/mnt/data/1kb" --timeout 10m --benchtime 3m
    BenchmarkReadRandomBadger/read-randombadger-128                 30000000            11185 ns/op
    --- BENCH: BenchmarkReadRandomBadger
            bench_test.go:104: badger: 18962075 keys had valid values.
            bench_test.go:105: badger: 11037925 keys had no values
            bench_test.go:106: badger: 0 keys had errors
            bench_test.go:107: badger: 30000000 total keys looked at
            bench_test.go:108: badger: hit rate : 0.63
    PASS
    ok      github.com/dgraph-io/badger-bench       351.715s

**Time iterate for lmdb** `**--keys_mil 75**` ****`**--valsz 1024**`
time w/ `lmdb.Readahead`

    $ go test -v --bench BenchmarkIterateLmdb --keys_mil 75 --valsz 1024 --dir "/mnt/data/1kb" --timeout 60m
    ....................BenchmarkIterateLmdb/lmdb-iterate-128                      1
            105376471805 ns/op
    --- BENCH: BenchmarkIterateLmdb/lmdb-iterate-128
            bench_test.go:317: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       105.602s
    ubuntu@ip-172-31-39-80:~/go/src/github.com/dgraph-io/badger-bench$ go test -v --bench BenchmarkIterateLmdb --keys_mil 75 --valsz 1024 --dir "/mnt/data/1kb" --timeout 60m
    ....................BenchmarkIterateLmdb/lmdb-iterate-128                      1
            2248460380 ns/op
    --- BENCH: BenchmarkIterateLmdb/lmdb-iterate-128
            bench_test.go:317: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       2.556s
    ubuntu@ip-172-31-39-80:~/go/src/github.com/dgraph-io/badger-bench$ go test -v --bench BenchmarkIterateLmdb --keys_mil 75 --valsz 1024 --dir "/mnt/data/1kb" --timeout 60m
    ....................BenchmarkIterateLmdb/lmdb-iterate-128                      1
            2247389381 ns/op
    --- BENCH: BenchmarkIterateLmdb/lmdb-iterate-128
            bench_test.go:317: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       2.448s

time w/o `lmdb.NoReadahead`

    go test -v --bench BenchmarkIterateLmdb --keys_mil 75 --valsz 1024 --dir "/mnt/data/1kb" --timeout 60m
    ....................
    BenchmarkIterateLmdb/lmdb-iterate-128                      1    259171636211 ns/op
    --- BENCH: BenchmarkIterateLmdb/lmdb-iterate-128
            bench_test.go:275: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       259.509s

**Time iterate for boltdb** `**--keys_mil 75**` ****`**--valsz 1024**`

    $ go test -v --bench BenchmarkIterateBolt --keys_mil 75  --valsz 1024 --dir "/mnt/data/1kb" --timeout 60m
    ....................BenchmarkIterateBolt/boltdb-iterate-128                    1
            89161784704 ns/op
    --- BENCH: BenchmarkIterateBolt/boltdb-iterate-128
            bench_test.go:363: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       89.266s
    
    $ go test -v --bench BenchmarkIterateBolt --keys_mil 75  --valsz 1024 --dir "/mnt/data/1kb" --timeout 60m
    ....................BenchmarkIterateBolt/boltdb-iterate-128             ........................................       2         569115231 ns/op
    --- BENCH: BenchmarkIterateBolt/boltdb-iterate-128
            bench_test.go:363: [0] Counted 2000000 keys
            bench_test.go:363: [0] Counted 2000000 keys
            bench_test.go:363: [1] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       2.151s
    
    $ go test -v --bench BenchmarkIterateBolt --keys_mil 75  --valsz 1024 --dir "/mnt/data/1kb" --timeout 60m
    ....................BenchmarkIterateBolt/boltdb-iterate-128             ........................................       2         573134064 ns/op
    --- BENCH: BenchmarkIterateBolt/boltdb-iterate-128
            bench_test.go:363: [0] Counted 2000000 keys
            bench_test.go:363: [0] Counted 2000000 keys
            bench_test.go:363: [1] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       2.066s
    ubuntu@ip-172-31-36-37:~/go/src/github.com/dgraph-io/badger-bench$
    

**Time iterate for badger (with values)** `**--keys_mil 75**` ****`**--valsz 1024**`

    $ go test -v --bench BenchmarkIterateBadgerWithValues --keys_mil 75 --valsz 1024 --dir "/mnt/data/1kb" --timeout 60m
    ....................BenchmarkIterateBadgerWithValues/badger-iterate-withvals-128                       1        23179999529 ns/op
    --- BENCH: BenchmarkIterateBadgerWithValues/badger-iterate-withvals-128
            bench_test.go:433: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       28.341s
    
    $ go test -v --bench BenchmarkIterateBadgerWithValues --keys_mil 75 --valsz 1024 --dir "/mnt/data/1kb" --timeout 60m
    ....................BenchmarkIterateBadgerWithValues/badger-iterate-withvals-128                       1        4313581872 ns/op
    --- BENCH: BenchmarkIterateBadgerWithValues/badger-iterate-withvals-128
            bench_test.go:433: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       8.745s
    
    $ go test -v --bench BenchmarkIterateBadgerWithValues --keys_mil 75 --valsz 1024 --dir "/mnt/data/1kb" --timeout 60m
    ....................BenchmarkIterateBadgerWithValues/badger-iterate-withvals-128                       1        4225754494 ns/op
    --- BENCH: BenchmarkIterateBadgerWithValues/badger-iterate-withvals-128
            bench_test.go:433: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       8.239s
    
    

**Time iterate for badger (keys only)** `**--keys_mil 75**` ****`**--valsz 1024**`

    $ go test -v --bench BenchmarkIterateBadgerOnlyKeys --keys_mil 75 --valsz 1024 --dir "/mnt/data/1kb" --timeout 60m
    ....................BenchmarkIterateBadgerOnlyKeys/badger-iterate-onlykeys-128          ........................................       2         673648670 ns/op
    --- BENCH: BenchmarkIterateBadgerOnlyKeys/badger-iterate-onlykeys-128
            bench_test.go:394: [0] Counted 2000000 keys
            bench_test.go:394: [0] Counted 2000000 keys
            bench_test.go:394: [1] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       6.852s
    
## Benchmarking `--keys_mil 250` and `--valsz 128`

At this point, our local SSD was pretty full, with only 100Gb of free space. This would not have been enough for this benchmark. So we did the following


    $ rm -rf /mnt/data/16kb /mnt/data/1kb
    $ fstrim -v
    mnt/data: 315.4 GiB (338627407872 bytes) trimmed
    $ df -h
    Filesystem      Size  Used Avail Use% Mounted on
    udev            7.5G     0  7.5G   0% /dev
    tmpfs           1.5G   25M  1.5G   2% /run
    /dev/xvda1      7.7G  2.3G  5.5G  29% /
    tmpfs           7.5G     0  7.5G   0% /dev/shm
    tmpfs           5.0M     0  5.0M   0% /run/lock
    tmpfs           7.5G     0  7.5G   0% /sys/fs/cgroup
    tmpfs           1.5G     0  1.5G   0% /run/user/1000
    /dev/nvme0n1    436G   71M  414G   1% /mnt/data

**Time population of lmdb** `**--valsz 128**` **and** `**--keys_mil 250**`

    $ /usr/bin/time -v ./populate --kv lmdb --valsz 128 --keys_mil 250 --dir /mnt/data/128b
    TOTAL KEYS TO WRITE: 250.00M
    Init lmdb
    [0000] Write key rate per minute:  85.00K. Total:  85.00K
    [0001] Write key rate per minute: 125.00K. Total: 125.00K
    [0002] Write key rate per minute: 162.00K. Total: 162.00K
    …<snip>…
    [27442] Write key rate per minute: 313.00K. Total: 249.99M
    [27443] Write key rate per minute: 319.00K. Total: 250.00M
    [27444] Write key rate per minute: 309.00K. Total: 250.00M
    closing lmdb
    
    WROTE 250008000 KEYS
            Command being timed: "./populate --kv lmdb --valsz 128 --keys_mil 250 --dir /mnt/data/128b"
            User time (seconds): 2052.60
            System time (seconds): 3901.18
            Percent of CPU this job got: 22%
            Elapsed (wall clock) time (h:mm:ss or m:ss): 7:24:45
            Average shared text size (kbytes): 0
            Average unshared data size (kbytes): 0
            Average stack size (kbytes): 0
            Average total size (kbytes): 0
            Maximum resident set size (kbytes): 14846000
            Average resident set size (kbytes): 0
            Major (requiring I/O) page faults: 76802692
            Minor (reclaiming a frame) page faults: 23967298
            Voluntary context switches: 91777445
            Involuntary context switches: 539592
            Swaps: 0
            File system inputs: 13858641520
            File system outputs: 2239372136
            Socket messages sent: 0
            Socket messages received: 0
            Signals delivered: 0
            Page size (bytes): 4096
            Exit status: 0
    


    $ sar -d 1 -p
    Average:          DEV       tps  rd_sec/s  wr_sec/s  avgrq-sz  avgqu-sz     await     svctm     %util
    Average:      nvme0n1  23664.15 685001.55 124322.31     34.20     52.01      2.20      0.04     91.01


    $ du -sh /mnt/data/128b/lmdb/
    36G     /mnt/data/128b/lmdb/

**Time population of boltdb** `**--valsz 128**` **and** `**--keys_mil 250**`

    $ /usr/bin/time -v ./populate --kv bolt --valsz 128 --keys_mil 250 --dir /mnt/data/128b
    TOTAL KEYS TO WRITE: 250.00M
    Init BoldDB
    [0000] Write key rate per minute:  89.00K. Total:  89.00K
    [0001] Write key rate per minute: 156.00K. Total: 156.00K
    [0002] Write key rate per minute: 216.00K. Total: 216.00K
    [0003] Write key rate per minute: 274.00K. Total: 274.00K
    …<snip>…
    [29959] Write key rate per minute: 333.00K. Total: 249.99M
    [29960] Write key rate per minute: 342.00K. Total: 249.99M
    [29961] Write key rate per minute: 351.00K. Total: 250.00M
    closing bolt
    
    WROTE 250008000 KEYS
            Command being timed: "./populate --kv bolt --valsz 128 --keys_mil 250 --dir /mnt/data/128b"
            User time (seconds): 13542.44
            System time (seconds): 2165.84
            Percent of CPU this job got: 52%
            Elapsed (wall clock) time (h:mm:ss or m:ss): 8:19:23
            Average shared text size (kbytes): 0
            Average unshared data size (kbytes): 0
            Average stack size (kbytes): 0
            Average total size (kbytes): 0
            Maximum resident set size (kbytes): 14840728
            Average resident set size (kbytes): 0
            Major (requiring I/O) page faults: 76816997
            Minor (reclaiming a frame) page faults: 7717740
            Voluntary context switches: 110316798
            Involuntary context switches: 1970763
            Swaps: 0
            File system inputs: 614570800
            File system outputs: 3064287104
            Socket messages sent: 0
            Socket messages received: 0
            Signals delivered: 0
            Page size (bytes): 4096
            Exit status: 0


    $ du -sh /mnt/data/128b/
    37G     /mnt/data/128b/

**Time population of badger** `**--valsz 128**` **and** `**--keys_mil 250**`
At this point, Badger was running out of memory when we ran the command below.

We made [some tweaks](https://github.com/dgraph-io/badger-bench/commit/166b53ce1c6d2b5de918bddf5e6d55ca3dbabd3e) to the way Badger is accessing the LSM tree data structure. At this point we also [made a change](https://github.com/dgraph-io/badger-bench/commit/417af08ccad705cb878a685c59ead21b25076576) to set the `ValueGCRunInterval` option to a very high value. This was an oversight on our part, and these values should have been in place from the very beginning. (FIXME *How do these changes impact the benchmarks done above?*)

After making the changes above, the populate run completed successfully.
 

    /usr/bin/time -v ./populate --kv badger --valsz 128 --keys_mil 250 --dir /mnt/data/128b
    TOTAL KEYS TO WRITE: 250.00M
    Init Badger
    [0000] Write key rate per minute: 295.00K. Total: 295.00K
    [0001] Write key rate per minute: 553.00K. Total: 553.00K
    [0002] Write key rate per minute: 757.00K. Total: 757.00K
    …<snip>…
    [2527] Write key rate per minute:   6.12M. Total: 249.73M
    [2528] Write key rate per minute:   6.25M. Total: 249.86M
    [2529] Write key rate per minute:   6.15M. Total: 249.91M
    closing badger
    
    WROTE 250008000 KEYS
        Command being timed: "./populate --kv badger --valsz 128 --keys_mil 250 --dir /mnt/data/128b"
        User time (seconds): 4586.68
        System time (seconds): 142.57
        Percent of CPU this job got: 186%
        Elapsed (wall clock) time (h:mm:ss or m:ss): 42:17.69
        Average shared text size (kbytes): 0
        Average unshared data size (kbytes): 0
        Average stack size (kbytes): 0
        Average total size (kbytes): 0
        Maximum resident set size (kbytes): 11385564
        Average resident set size (kbytes): 0
        Major (requiring I/O) page faults: 179
        Minor (reclaiming a frame) page faults: 1701827
        Voluntary context switches: 2315182
        Involuntary context switches: 1955656
        Swaps: 0
        File system inputs: 37560
        File system outputs: 236382672
        Socket messages sent: 0
        Socket messages received: 0
        Signals delivered: 0
        Page size (bytes): 4096
        Exit status: 0


    $ sar -d 1 -p
    Average:          DEV       tps  rd_sec/s  wr_sec/s  avgrq-sz  avgqu-sz     await     svctm     %util
    Average:      nvme0n1    384.31      3.29  96121.63    250.12     30.66     79.78      0.18      6.87


    $ du -sh /mnt/data/128b/badger/
    49G     /mnt/data/128b/badger/

**Time random read for lmdb** `**--keys_mil 250**` ****`**--valsz 128**`

    $ go test -v --bench BenchmarkReadRandomLmdb --keys_mil 250 --valsz 128 --dir "/mnt/data/128b" --timeout 10m --benchtime 3m
    BenchmarkReadRandomLmdb/read-randomlmdb-128             30000000              7858 ns/op
    --- BENCH: BenchmarkReadRandomLmdb
            bench_test.go:104: lmdb: 18963166 keys had valid values.
            bench_test.go:105: lmdb: 11036834 keys had no values
            bench_test.go:106: lmdb: 0 keys had errors
            bench_test.go:107: lmdb: 30000000 total keys looked at
            bench_test.go:108: lmdb: hit rate : 0.63
    PASS
    ok      github.com/dgraph-io/badger-bench       246.661s

**Time random read for boltdb** `**--keys_mil 250**` ****`**--valsz 128**`

    $ go test -v --bench BenchmarkReadRandomBolt --keys_mil 250 --valsz 128 --dir "/mnt/data/128b" --timeout 10m --benchtime 3m
    BenchmarkReadRandomBolt/read-randombolt-128             20000000             10899 ns/op
    --- BENCH: BenchmarkReadRandomBolt
            bench_test.go:104: bolt: 12638883 keys had valid values.
            bench_test.go:105: bolt: 7361117 keys had no values
            bench_test.go:106: bolt: 0 keys had errors
            bench_test.go:107: bolt: 20000000 total keys looked at
            bench_test.go:108: bolt: hit rate : 0.63
    PASS
    ok      github.com/dgraph-io/badger-bench       230.213s
    

**Time random read for badger** `**--keys_mil 250**` ****`**--valsz 128**`

    $ go test -v --bench BenchmarkReadRandomBadger --keys_mil 250 --valsz 128 --dir "/mnt/data/128b" --timeout 10m --benchtime 3m
    BenchmarkReadRandomBadger/read-randombadger-128                 20000000            12517 ns/op
    --- BENCH: BenchmarkReadRandomBadger
            bench_test.go:104: badger: 12646587 keys had valid values.
            bench_test.go:105: badger: 7353413 keys had no values
            bench_test.go:106: badger: 0 keys had errors
            bench_test.go:107: badger: 20000000 total keys looked at
            bench_test.go:108: badger: hit rate : 0.63
    PASS
    ok      github.com/dgraph-io/badger-bench       277.471s

**Time iterate for lmdb** `**--keys_mil 250**` ****`**--valsz 128**`
Times for 3 consecutive runs w/ `lmdb.NoReadAhead`

    $ go test -v --bench BenchmarkIterateLmdb --keys_mil 250 --valsz 128 --dir "/mnt/data/128b" --timeout 60m
    ....................BenchmarkIterateLmdb/lmdb-iterate-128                      1
            13819169930 ns/op
    --- BENCH: BenchmarkIterateLmdb/lmdb-iterate-128
            bench_test.go:317: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       13.881s
    ubuntu@ip-172-31-47-171:~/go/src/github.com/dgraph-io/badger-bench$ go test -v --bench BenchmarkIterateLmdb --keys_mil 250 --valsz 128 --dir "/mnt/data/128b" --timeout 60m
    ....................BenchmarkIterateLmdb/lmdb-iterate-128               ........................................       2         720681417 ns/op
    --- BENCH: BenchmarkIterateLmdb/lmdb-iterate-128
            bench_test.go:317: [0] Counted 2000000 keys
            bench_test.go:317: [0] Counted 2000000 keys
            bench_test.go:317: [1] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       2.393s
    ubuntu@ip-172-31-47-171:~/go/src/github.com/dgraph-io/badger-bench$ go test -v --bench BenchmarkIterateLmdb --keys_mil 250 --valsz 128 --dir "/mnt/data/128b" --timeout 60m
    ....................BenchmarkIterateLmdb/lmdb-iterate-128               ........................................       2         721250591 ns/op
    --- BENCH: BenchmarkIterateLmdb/lmdb-iterate-128
            bench_test.go:317: [0] Counted 2000000 keys
            bench_test.go:317: [0] Counted 2000000 keys
            bench_test.go:317: [1] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       2.392s

Times for 3 consecutive runs w/o `lmdb.NoReadahead`

    ubuntu@ip-172-31-39-80:~/go/src/github.com/dgraph-io/badger-bench$ go test -v --bench BenchmarkIterateLmdb --keys_mil 250 --valsz 128 --dir "/mnt/data/128b" --timeout 60m
    ....................BenchmarkIterateLmdb/lmdb-iterate-128                      1
            19946953386 ns/op
    --- BENCH: BenchmarkIterateLmdb/lmdb-iterate-128
            bench_test.go:317: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       20.219s
    ubuntu@ip-172-31-39-80:~/go/src/github.com/dgraph-io/badger-bench$ go test -v --bench BenchmarkIterateLmdb --keys_mil 250 --valsz 128 --dir "/mnt/data/128b" --timeout 60m
    ....................BenchmarkIterateLmdb/lmdb-iterate-128                      1
            1319455375 ns/op
    --- BENCH: BenchmarkIterateLmdb/lmdb-iterate-128
            bench_test.go:317: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       1.632s
    ubuntu@ip-172-31-39-80:~/go/src/github.com/dgraph-io/badger-bench$ go test -v --bench BenchmarkIterateLmdb --keys_mil 250 --valsz 128 --dir "/mnt/data/128b" --timeout 60m
    ....................BenchmarkIterateLmdb/lmdb-iterate-128                      1
            1129621638 ns/op
    --- BENCH: BenchmarkIterateLmdb/lmdb-iterate-128
            bench_test.go:317: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       1.398s
    
    

**Time iterate for boltdb** `**--keys_mil 250**` ****`**--valsz 128**`
3 consecutive runs

    $ go test -v --bench BenchmarkIterateBolt --keys_mil 250 --valsz 128 --dir "/mnt/data/128b" --timeout 60m
    ....................BenchmarkIterateBolt/boltdb-iterate-128                    1
            12878891767 ns/op
    --- BENCH: BenchmarkIterateBolt/boltdb-iterate-128
            bench_test.go:363: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       12.908s
    ubuntu@ip-172-31-36-37:~/go/src/github.com/dgraph-io/badger-bench$ go test -v --bench BenchmarkIterateBolt --keys_mil 250 --valsz 128 --dir "/mnt/data/128b" --timeout 60m
    ....................BenchmarkIterateBolt/boltdb-iterate-128             ................................................................................................................................................................       5
             255047587 ns/op
    --- BENCH: BenchmarkIterateBolt/boltdb-iterate-128
            bench_test.go:363: [0] Counted 2000000 keys
            bench_test.go:363: [0] Counted 2000000 keys
            bench_test.go:363: [1] Counted 2000000 keys
            bench_test.go:363: [2] Counted 2000000 keys
            bench_test.go:363: [0] Counted 2000000 keys
            bench_test.go:363: [1] Counted 2000000 keys
            bench_test.go:363: [2] Counted 2000000 keys
            bench_test.go:363: [3] Counted 2000000 keys
            bench_test.go:363: [4] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       2.389s
    ubuntu@ip-172-31-36-37:~/go/src/github.com/dgraph-io/badger-bench$ go test -v --bench BenchmarkIterateBolt --keys_mil 250 --valsz 128 --dir "/mnt/data/128b" --timeout 60m
    ....................BenchmarkIterateBolt/boltdb-iterate-128             ................................................................................................................................................................       5
             259654821 ns/op
    --- BENCH: BenchmarkIterateBolt/boltdb-iterate-128
            bench_test.go:363: [0] Counted 2000000 keys
            bench_test.go:363: [0] Counted 2000000 keys
            bench_test.go:363: [1] Counted 2000000 keys
            bench_test.go:363: [2] Counted 2000000 keys
            bench_test.go:363: [0] Counted 2000000 keys
            bench_test.go:363: [1] Counted 2000000 keys
            bench_test.go:363: [2] Counted 2000000 keys
            bench_test.go:363: [3] Counted 2000000 keys
            bench_test.go:363: [4] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       2.399s
    

**Time iterate for badger (with values)** `**--keys_mil 250**` ****`**--valsz 128**`
It is worth noting here that there is a large startup delay for Badger, possibly due to loading and setting up the tables in memory. The actual time to iterate over keys is 19.3s. The same thing applies to the next benchmark as well.

    $ go test -v --bench BenchmarkIterateBadgerWithValues --keys_mil 250 --valsz 128 --dir "/mnt/data/128b" --timeout 60m
    ....................BenchmarkIterateBadgerWithValues/badger-iterate-withvals-128                       1        20606028744 ns/op
    --- BENCH: BenchmarkIterateBadgerWithValues/badger-iterate-withvals-128
            bench_test.go:433: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       36.844s
    
    $ go test -v --bench BenchmarkIterateBadgerWithValues --keys_mil 250 --valsz 128 --dir "/mnt/data/128b" --timeout 60m
    ....................BenchmarkIterateBadgerWithValues/badger-iterate-withvals-128                       1        21468654790 ns/op
    --- BENCH: BenchmarkIterateBadgerWithValues/badger-iterate-withvals-128
            bench_test.go:433: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       39.128s
    

**Time iterate for badger (keys only)** `**--keys_mil 250**` ****`**--valsz 128**`

    $ go test -v --bench BenchmarkIterateBadgerOnlyKeys --keys_mil 250 --valsz 128 --dir "/mnt/data/128b" --timeout 60m
    ....................BenchmarkIterateBadgerOnlyKeys/badger-iterate-onlykeys-128          ........................................       2         693002872 ns/op
    --- BENCH: BenchmarkIterateBadgerOnlyKeys/badger-iterate-onlykeys-128
            bench_test.go:394: [0] Counted 2000000 keys
            bench_test.go:394: [0] Counted 2000000 keys
            bench_test.go:394: [1] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       17.944s
    



