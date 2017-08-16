# Benchmarking lmdb-go

- Badger benchmarking repo: https://github.com/dgraph-io/badger-bench
- lmdb-go: https://github.com/bmatsuo/lmdb-go
# Coding
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


- Install build-essentials: `sudo apt-get install build-essentials`


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
    $ go get github.com/dgraph-io/badger-bench


- Run `go test` and make sure rocksdb is linked up


- Put this in `~/.bashrc`
    export GOMAXPROCS=128
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
    [0000] Write key rate per minute:  10.00K. Total:  10.00K
    [0001] Write key rate per minute:  19.00K. Total:  19.00K
    [0002] Write key rate per minute:  28.00K. Total:  28.00K
    [0003] Write key rate per minute:  37.00K. Total:  37.00K
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

    $ go test -v --bench BenchmarkReadRandomLmdb --keys_mil 5 --valsz 16384 --dir "/mnt/data/16kb" --timeout 10m --benchtime 3m
    BenchmarkReadRandomLmdb/read-random-lmdb-128             2000000            125050 ns/op
    --- BENCH: BenchmarkReadRandomLmdb
            bench_test.go:178: lmdb 1265654 keys had valid values.
            bench_test.go:179: lmdb 734346 keys had no values
            bench_test.go:180: lmdb 0 keys had errors
            bench_test.go:181: lmdb 2000000 total keys looked at
            bench_test.go:182: lmdb hit rate : 0.63
    PASS
    ok      github.com/dgraph-io/badger-bench       389.863s
    

**Time random read for badger** `**--keys_mil 5**` ****`**--valsz 16384**`

    $ go test -v --bench BenchmarkReadRandomBadger --keys_mil 5 --valsz 16384 --dir "/mnt/data/16kb" --timeout 10m --benchtime 3m                                           /mnt/data/16kb/badger
    BenchmarkReadRandomBadger/read-randombadger-128                 10000000            22578 ns/op
    --- BENCH: BenchmarkReadRandomBadger
            bench_test.go:94: badger: 6324701 keys had valid values.
            bench_test.go:95: badger: 3675299 keys had no values
            bench_test.go:96: badger: 0 keys had errors
            bench_test.go:97: badger: 10000000 total keys looked at
            bench_test.go:98: badger: hit rate : 0.63
    PASS
    ok      github.com/dgraph-io/badger-bench       249.877s 

**Time iterate for lmdb** `**--keys_mil 5**` ****`**--valsz 16384**`

    $ go test -v --bench BenchmarkIterateLmdb --keys_mil 5 --valsz 16384 --dir "/mnt/data/16kb" --timeout 60m
    ....................
    BenchmarkIterateLmdb/lmdb-iterate-128                      1    488071445140 ns/op
    --- BENCH: BenchmarkIterateLmdb/lmdb-iterate-128
            bench_test.go:275: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       488.454s
    
    

**Time iterate for badger (with values)** `**--keys_mil 5**` ****`**--valsz 16384**`

    $ go test -v --bench BenchmarkIterateBadgerWithValues --keys_mil 5 --valsz 16384 --dir "/mnt/data/16kb" --timeout 60m
    /mnt/data/16kb/badger
    ....................BenchmarkIterateBadgerWithValues/badger-iterate-withvals-128                       1        89551210385 ns/op
    --- BENCH: BenchmarkIterateBadgerWithValues/badger-iterate-withvals-128
            bench_test.go:349: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       90.070s

**Time iterate for badger (keys only)** `**--keys_mil 5**` ****`**--valsz 16384**`

    $ go test -v --bench BenchmarkIterateBadgerOnlyKeys --keys_mil 5 --valsz 16384 --dir "/mnt/data/16kb" --timeout 60m
    /mnt/data/16kb/badger
    ....................BenchmarkIterateBadgerOnlyKeys/badger-iterate-onlykeys-128          ........................................       2         502355718 ns/op
    --- BENCH: BenchmarkIterateBadgerOnlyKeys/badger-iterate-onlykeys-128
            bench_test.go:314: [0] Counted 2000000 keys
            bench_test.go:314: [0] Counted 2000000 keys
            bench_test.go:314: [1] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       1.830s


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
    BenchmarkReadRandomLmdb/read-random-lmdb-128             1000000            304019 ns/op
    --- BENCH: BenchmarkReadRandomLmdb
            bench_test.go:178: lmdb 632140 keys had valid values.
            bench_test.go:179: lmdb 367860 keys had no values
            bench_test.go:180: lmdb 0 keys had errors
            bench_test.go:181: lmdb 1000000 total keys looked at
            bench_test.go:182: lmdb hit rate : 0.63
    PASS
    ok      github.com/dgraph-io/badger-bench       306.704s

**Time random read for badger** `**--keys_mil 75**` ****`**--valsz 1024**`

    $ go test -v --bench BenchmarkReadRandomBadger --keys_mil 75 --valsz 1024 --dir "/mnt/data/1kb" --timeout 10m --benchtime 3m
    /mnt/data/1kb/badger
    BenchmarkReadRandomBadger/read-randombadger-128                 20000000            10919 ns/op
    --- BENCH: BenchmarkReadRandomBadger
            bench_test.go:94: badger: 12642254 keys had valid values.
            bench_test.go:95: badger: 7357746 keys had no values
            bench_test.go:96: badger: 0 keys had errors
            bench_test.go:97: badger: 20000000 total keys looked at
            bench_test.go:98: badger: hit rate : 0.63
    PASS
    ok      github.com/dgraph-io/badger-bench       236.196s

**Time iterate for lmdb** `**--keys_mil 75**` ****`**--valsz 1024**`


    go test -v --bench BenchmarkIterateLmdb --keys_mil 75 --valsz 1024 --dir "/mnt/data/1kb" --timeout 60m
    ....................
    BenchmarkIterateLmdb/lmdb-iterate-128                      1    259171636211 ns/op
    --- BENCH: BenchmarkIterateLmdb/lmdb-iterate-128
            bench_test.go:275: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       259.509s

**Time iterate for badger (with values)** `**--keys_mil 75**` ****`**--valsz 1024**`

    $ go test -v --bench BenchmarkIterateBadgerWithValues --keys_mil 75 --valsz 1024 --dir "/mnt/data/1kb" --timeout 60m
    /mnt/data/1kb/badger
    ....................BenchmarkIterateBadgerWithValues/badger-iterate-withvals-128                       1        19843042846 ns/op
    --- BENCH: BenchmarkIterateBadgerWithValues/badger-iterate-withvals-128
            bench_test.go:349: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       25.924s

**Time iterate for badger (keys only)** `**--keys_mil 75**` ****`**--valsz 1024**`

    $ go test -v --bench BenchmarkIterateBadgerOnlyKeys --keys_mil 75 --valsz 1024 --dir "/mnt/data/1kb" --timeout 60m
    /mnt/data/1kb/badger
    ....................BenchmarkIterateBadgerOnlyKeys/badger-iterate-onlykeys-128          ........................................       2         573743486 ns/op
    --- BENCH: BenchmarkIterateBadgerOnlyKeys/badger-iterate-onlykeys-128
            bench_test.go:314: [0] Counted 2000000 keys
            bench_test.go:314: [0] Counted 2000000 keys
            bench_test.go:314: [1] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       7.839s
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
            User time (seconds): 1560.75
            System time (seconds): 6763.16
            Percent of CPU this job got: 30%
            Elapsed (wall clock) time (h:mm:ss or m:ss): 7:37:26
            Average shared text size (kbytes): 0
            Average unshared data size (kbytes): 0
            Average stack size (kbytes): 0
            Average total size (kbytes): 0
            Maximum resident set size (kbytes): 14684608
            Average resident set size (kbytes): 0
            Major (requiring I/O) page faults: 77358150
            Minor (reclaiming a frame) page faults: 23734451
            Voluntary context switches: 152401312
            Involuntary context switches: 1262078
            Swaps: 0
            File system inputs: 13931923344
            File system outputs: 5024014272
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
    BenchmarkReadRandomLmdb/read-randomlmdb-128              2000000            150972 ns/op
    --- BENCH: BenchmarkReadRandomLmdb
            bench_test.go:96: lmdb: 1263874 keys had valid values.
            bench_test.go:97: lmdb: 736126 keys had no values
            bench_test.go:98: lmdb: 0 keys had errors
            bench_test.go:99: lmdb: 2000000 total keys looked at
            bench_test.go:100: lmdb: hit rate : 0.63
    PASS
    ok      github.com/dgraph-io/badger-bench       572.472s

**Time random read for badger** `**--keys_mil 250**` ****`**--valsz 128**`

    $ go test -v --bench BenchmarkReadRandomBadger --keys_mil 250 --valsz 128 --dir "/mnt/data/128b" --timeout 10m --benchtime 3m
    /mnt/data/128b/badger
    BenchmarkReadRandomBadger/read-randombadger-128                 20000000             11759 ns/op
    --- BENCH: BenchmarkReadRandomBadger
            bench_test.go:96: badger: 12640014 keys had valid values.
            bench_test.go:97: badger: 7359986 keys had no values
            bench_test.go:98: badger: 0 keys had errors
            bench_test.go:99: badger: 20000000 total keys looked at
            bench_test.go:100: badger: hit rate : 0.63
    PASS
    ok      github.com/dgraph-io/badger-bench       269.888s

**Time iterate for lmdb** `**--keys_mil 250**` ****`**--valsz 128**`

    go test -v --bench BenchmarkIterateLmdb --keys_mil 250 --valsz 128 --dir "/mnt/data/128b" --timeout 60m
    ....................
    BenchmarkIterateLmdb/lmdb-iterate-128                      1        30361801869 ns/op
    --- BENCH: BenchmarkIterateLmdb/lmdb-iterate-128
            bench_test.go:285: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       30.591s
    

**Time iterate for badger (with values)** `**--keys_mil 250**` ****`**--valsz 128**`
It is worth noting here that there is a large startup delay for Badger, possibly due to loading and setting up the tables in memory. The actual time to iterate over keys is 19.3s which compares favorably with Badger’s 30.3s. The same thing applies to the next benchmark as well.

    $ go test -v --bench BenchmarkIterateBadgerWithValues --keys_mil 250 --valsz 128 --dir "/mnt/data/128b" --timeout 60m
    /mnt/data/128b/badger
    ....................BenchmarkIterateBadgerWithValues/badger-iterate-withvals-128                       1        19303328164 ns/op
    --- BENCH: BenchmarkIterateBadgerWithValues/badger-iterate-withvals-128
            bench_test.go:351: [0] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       40.655s

**Time iterate for badger (keys only)** `**--keys_mil 250**` ****`**--valsz 128**`

    $ go test -v --bench BenchmarkIterateBadgerOnlyKeys --keys_mil 250 --valsz 128 --dir "/mnt/data/128b" --timeout 60m
    /mnt/data/128b/badger
    ....................BenchmarkIterateBadgerOnlyKeys/badger-iterate-onlykeys-128          ........................................       2         819631115 ns/op
    --- BENCH: BenchmarkIterateBadgerOnlyKeys/badger-iterate-onlykeys-128
            bench_test.go:316: [0] Counted 2000000 keys
            bench_test.go:316: [0] Counted 2000000 keys
            bench_test.go:316: [1] Counted 2000000 keys
    PASS
    ok      github.com/dgraph-io/badger-bench       22.239s

