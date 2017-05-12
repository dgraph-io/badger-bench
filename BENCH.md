# Amazon i3.large dedicated instance: 2 cores, 16G RAM, 450G local SSD.

As shown by fio, this instance gives 93K random iops at 4K block size.

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
Jobs: 1 (f=1): [r(1),_(15)] [100.0% done] [240.4MB/0KB/0KB /s] [61.6K/0/0 iops] [eta 00m:00s]        s]
randread: (groupid=0, jobs=16): err= 0: pid=13063: Sat Apr 29 12:37:49 2017
  read : io=32768MB, bw=371947KB/s, iops=92986, runt= 90213msec
    slat (usec): min=31, max=24800, avg=163.54, stdev=200.35
    clat (usec): min=1, max=69452, avg=5180.95, stdev=1919.00
     lat (usec): min=91, max=69546, avg=5345.05, stdev=1958.18
    clat percentiles (usec):
     |  1.00th=[ 3152],  5.00th=[ 3312], 10.00th=[ 3440], 20.00th=[ 3664],
     | 30.00th=[ 3856], 40.00th=[ 4128], 50.00th=[ 4512], 60.00th=[ 5024],
     | 70.00th=[ 5728], 80.00th=[ 6624], 90.00th=[ 7904], 95.00th=[ 9024],
     | 99.00th=[11456], 99.50th=[12352], 99.90th=[14528], 99.95th=[15680],
     | 99.99th=[20096]
    bw (KB  /s): min=18632, max=36608, per=6.43%, avg=23925.03, stdev=2987.85
    lat (usec) : 2=0.01%, 4=0.01%, 100=0.01%, 250=0.01%, 500=0.01%
    lat (usec) : 750=0.01%, 1000=0.01%
    lat (msec) : 2=0.01%, 4=36.40%, 10=60.93%, 20=2.66%, 50=0.01%
    lat (msec) : 100=0.01%
  cpu          : usr=2.31%, sys=6.14%, ctx=8463944, majf=0, minf=653
  IO depths    : 1=0.1%, 2=0.1%, 4=0.1%, 8=0.1%, 16=0.1%, 32=100.0%, >=64=0.0%
     submit    : 0=0.0%, 4=100.0%, 8=0.0%, 16=0.0%, 32=0.0%, 64=0.0%, >=64=0.0%
     complete  : 0=0.0%, 4=100.0%, 8=0.0%, 16=0.0%, 32=0.1%, 64=0.0%, >=64=0.0%
     issued    : total=r=8388608/w=0/d=0, short=r=0/w=0/d=0, drop=r=0/w=0/d=0
     latency   : target=0, window=0, percentile=100.00%, depth=32

Run status group 0 (all jobs):
   READ: io=32768MB, aggrb=371946KB/s, minb=371946KB/s, maxb=371946KB/s, mint=90213msec, maxt=90213msec

Disk stats (read/write):
  nvme0n1: ios=8386313/19379, merge=0/0, ticks=877396/60, in_queue=880100, util=100.00%

        Command being timed: "./populate --kv rocksdb --valsz 128 --keys_mil 250"
        User time (seconds): 2685.96
        System time (seconds): 532.66
        Percent of CPU this job got: 136%
        Elapsed (wall clock) time (h:mm:ss or m:ss): 39:10.78
        Average shared text size (kbytes): 0
        Average unshared data size (kbytes): 0
        Average stack size (kbytes): 0
        Average total size (kbytes): 0
        Maximum resident set size (kbytes): 611888
        Average resident set size (kbytes): 0
        Major (requiring I/O) page faults: 39
        Minor (reclaiming a frame) page faults: 2169690
        Voluntary context switches: 11455264
        Involuntary context switches: 4606594
        Swaps: 0
        File system inputs: 132138368
        File system outputs: 594809048
        Socket messages sent: 0
        Socket messages received: 0
        Signals delivered: 0
        Page size (bytes): 4096
        Exit status: 0

$ du -sh /mnt/data/rocks
24G     /mnt/data/rocks


In this case, we set the value log GC threshold to 0.5. Turns out doing value log GC can be expensive.
So, we should only do it sometimes. It's only worth if saving significant amount of disk space.

        Command being timed: "./populate --kv badger --valsz 128 --keys_mil 250"
        User time (seconds): 4983.09
        System time (seconds): 166.96
        Percent of CPU this job got: 188%
        Elapsed (wall clock) time (h:mm:ss or m:ss): 45:26.56
        Average shared text size (kbytes): 0
        Average unshared data size (kbytes): 0
        Average stack size (kbytes): 0
        Average total size (kbytes): 0
        Maximum resident set size (kbytes): 14660624
        Average resident set size (kbytes): 0
        Major (requiring I/O) page faults: 10690
        Minor (reclaiming a frame) page faults: 6659331
        Voluntary context switches: 1141184
        Involuntary context switches: 1071168
        Swaps: 0
        File system inputs: 14994928
        File system outputs: 291238896
        Socket messages sent: 0
        Socket messages received: 0
        Signals delivered: 0
        Page size (bytes): 4096
        Exit status: 0


$ du -sh /mnt/data/badger
38G     /mnt/data/badger
5.8G *.sst  # LSM tree, can be kept in RAM.

Random Reads: Badger is 3.67x faster

$ go test --bench BenchmarkReadRandomRocks --keys_mil 250 --valsz 128 --dir "/mnt/data" --timeout 10m --benchtime 3m 
BenchmarkReadRandomRocks/read-random-rocks-2             2000000            118982 ns/op
--- BENCH: BenchmarkReadRandomRocks/read-random-rocks-2
        bench_test.go:92: rocks 149864 keys had valid values.
        bench_test.go:92: rocks 150136 keys had valid values.
        bench_test.go:92: rocks 1000693 keys had valid values.
        bench_test.go:92: rocks 999307 keys had valid values.
PASS

$ go test --bench BenchmarkReadRandomBadger --keys_mil 250 --valsz 128 --dir "/mnt/data" --timeout 10m --benchtime 3m
Called BenchmarkReadRandomBadger
Replaying compact log: /mnt/data/badger/clog
All compactions in compact log are done.
NOT running any compactions due to DB options.
NOT running any compactions due to DB options.
NOT running any compactions due to DB options.
Seeking at value pointer: {Fid:37 Len:163 Offset:1022845212}
l.opt.ValueGCThreshold = 0.0. Exiting runGCInLoop
key=vsz=00128-k=0025059055
BenchmarkReadRandomBadger/read-random-badger-2          10000000            32361 ns/op
--- BENCH: BenchmarkReadRandomBadger/read-random-badger-2
        bench_test.go:72: badger 325009 keys had valid values.
        bench_test.go:72: badger 324883 keys had valid values.
        bench_test.go:72: badger 3247736 keys had valid values.
        bench_test.go:72: badger 3243258 keys had valid values.
Sending signal to 0 registered with name "value-gc"
Sending signal to 1 registered with name "writes"
--->> Size of bloom filter: 116
=======> Deallocating skiplist
Sending signal to 0 registered with name "memtable"
Level "value-gc" already got signal
Level "writes" already got signal
PASS

### Iteration

$ go test --bench BenchmarkIterateRocks --keys_mil 250 --valsz 128 --dir "/mnt/data" --timeout 10m --cpuprofile cpu.out
BenchmarkIterateRocks/rocksdb-iterate-2                        1        5806763436 ns/op
--- BENCH: BenchmarkIterateRocks/rocksdb-iterate-2
        bench_test.go:128: [0] Counted 2000000 keys
PASS
ok      github.com/dgraph-io/badger-bench       6.987s

$ go test --bench BenchmarkIterateBadgerOnly --keys_mil 250 --valsz 128 --dir "/mnt/data" --timeout 10m --cpuprofile cpu.out
Replaying compact log: /mnt/data/badger/clog
All compactions in compact log are done.
NOT running any compactions due to DB options.
NOT running any compactions due to DB options.
NOT running any compactions due to DB options.
Seeking at value pointer: {Fid:39 Len:163 Offset:1012268142}
l.opt.ValueGCThreshold = 0.0. Exiting runGCInLoop
key=vsz=00128-k=0098569193
BenchmarkIterateBadgerOnlyKeys/badger-iterate-onlykeys-2                       2         713078716 ns/op
--- BENCH: BenchmarkIterateBadgerOnlyKeys/badger-iterate-onlykeys-2
        bench_test.go:156: [0] Counted 2000000 keys
        bench_test.go:156: [0] Counted 2000000 keys
        bench_test.go:156: [1] Counted 2000000 keys
PASS
ok      github.com/dgraph-io/badger-bench       10.198s

$ go test --bench BenchmarkIterateBadgerWithValues --keys_mil 250 --valsz 128 --dir "/mnt/data" --timeout 10m
Replaying compact log: /mnt/data/badger/clog
All compactions in compact log are done.
NOT running any compactions due to DB options.
NOT running any compactions due to DB options.
NOT running any compactions due to DB options.
Seeking at value pointer: {Fid:39 Len:163 Offset:1012268142}
l.opt.ValueGCThreshold = 0.0. Exiting runGCInLoop
key=vsz=00128-k=0098569193
....................BenchmarkIterateBadgerWithValues/badger-iterate-withvals-2                 1        75781455080 ns/op
--- BENCH: BenchmarkIterateBadgerWithValues/badger-iterate-withvals-2
        bench_test.go:187: [0] Counted 2000000 keys
PASS
ok      github.com/dgraph-io/badger-bench       81.401s

WROTE 75000000 KEYS
        Command being timed: "./populate --kv rocksdb --valsz 1024 --keys_mil 75 --dir /mnt/data/1kb"
        User time (seconds): 2529.19
        System time (seconds): 1498.27
        Percent of CPU this job got: 85%
        Elapsed (wall clock) time (h:mm:ss or m:ss): 1:18:08
        Average shared text size (kbytes): 0
        Average unshared data size (kbytes): 0
        Average stack size (kbytes): 0
        Average total size (kbytes): 0
        Maximum resident set size (kbytes): 1040732
        Average resident set size (kbytes): 0
        Major (requiring I/O) page faults: 298
        Minor (reclaiming a frame) page faults: 11338619
        Voluntary context switches: 6822622
        Involuntary context switches: 1738511
        Swaps: 0
        File system inputs: 1046110728
        File system outputs: 1814480952
        Socket messages sent: 0
        Socket messages received: 0
        Signals delivered: 0
        Page size (bytes): 4096
        Exit status: 0

$ du -sh /mnt/data/1kb/rocks
49G


WROTE 75000000 KEYS
        Command being timed: "./populate --kv badger --valsz 1024 --keys_mil 75 --dir /mnt/data/1kb"
        User time (seconds): 1445.97
        System time (seconds): 109.23
        Percent of CPU this job got: 151%
        Elapsed (wall clock) time (h:mm:ss or m:ss): 17:09.66
        Average shared text size (kbytes): 0
        Average unshared data size (kbytes): 0
        Average stack size (kbytes): 0
        Average total size (kbytes): 0
        Maximum resident set size (kbytes): 11857204
        Average resident set size (kbytes): 0
        Major (requiring I/O) page faults: 1952
        Minor (reclaiming a frame) page faults: 10187929
        Voluntary context switches: 1804454
        Involuntary context switches: 282003
        Swaps: 0
        File system inputs: 205176
        File system outputs: 197457568
        Socket messages sent: 0
        Socket messages received: 0
        Signals delivered: 0
        Page size (bytes): 4096
        Exit status: 0

$ du -shc /mnt/data/1kb/badger/*.sst
1.7G

$ du -shc /mnt/data/1kb/badger/*.vlog
74G

$ go test --bench BenchmarkReadRandomRocks --keys_mil 75 --valsz 1024 --dir "/mnt/data/1kb" --timeout 10m --benchtime 3m                                                
BenchmarkReadRandomRocks/read-random-rocks-2             2000000            156694 ns/op
--- BENCH: BenchmarkReadRandomRocks/read-random-rocks-2
        bench_test.go:92: rocks 149796 keys had valid values.
        bench_test.go:92: rocks 150204 keys had valid values.
        bench_test.go:92: rocks 996349 keys had valid values.
        bench_test.go:92: rocks 1003651 keys had valid values.
PASS
ok      github.com/dgraph-io/badger-bench       385.121s

$ go test --bench BenchmarkReadRandomBadger --keys_mil 75 --valsz 1024 --dir "/mnt/data/1kb" --timeout 10m --benchtime 3m
Called BenchmarkReadRandomBadger
Replaying compact log: /mnt/data/1kb/badger/clog
All compactions in compact log are done.
NOT running any compactions due to DB options.
NOT running any compactions due to DB options.
NOT running any compactions due to DB options.
Seeking at value pointer: {Fid:73 Len:1059 Offset:1041730887}
l.opt.ValueGCThreshold = 0.0. Exiting runGCInLoop
key=vsz=01024-k=0015263159
BenchmarkReadRandomBadger/read-random-badger-2          10000000             37053 ns/op
--- BENCH: BenchmarkReadRandomBadger/read-random-badger-2
        bench_test.go:72: badger 317463 keys had valid values.
        bench_test.go:72: badger 317460 keys had valid values.
        bench_test.go:72: badger 3175198 keys had valid values.
        bench_test.go:72: badger 3169988 keys had valid values.
Sending signal to 0 registered with name "value-gc"
Sending signal to 1 registered with name "writes"
--->> Size of bloom filter: 116
=======> Deallocating skiplist
Sending signal to 0 registered with name "memtable"
Level "value-gc" already got signal
Level "writes" already got signal
PASS
ok      github.com/dgraph-io/badger-bench       415.068s

$ go test --bench BenchmarkIterateRocks --keys_mil 75 --valsz 1024 --dir "/mnt/data/1kb" --timeout 60m                         
BenchmarkIterateRocks/rocksdb-iterate-2                        1        24936162613 ns/op
--- BENCH: BenchmarkIterateRocks/rocksdb-iterate-2
        bench_test.go:128: [0] Counted 2000001 keys
PASS
ok      github.com/dgraph-io/badger-bench       26.416s

$ go test --bench BenchmarkIterateBadger --keys_mil 75 --valsz 1024 --dir "/mnt/data/1kb" --timeout 60m
Replaying compact log: /mnt/data/1kb/badger/clog
All compactions in compact log are done.
NOT running any compactions due to DB options.
NOT running any compactions due to DB options.
NOT running any compactions due to DB options.
Seeking at value pointer: {Fid:73 Len:1059 Offset:1041730887}
l.opt.ValueGCThreshold = 0.0. Exiting runGCInLoop
key=vsz=01024-k=0015263159
BenchmarkIterateBadgerOnlyKeys/badger-iterate-onlykeys-2                       2         536687829 ns/op
--- BENCH: BenchmarkIterateBadgerOnlyKeys/badger-iterate-onlykeys-2
        bench_test.go:156: [0] Counted 2000001 keys
        bench_test.go:156: [0] Counted 2000001 keys
        bench_test.go:156: [1] Counted 2000001 keys
Replaying compact log: /mnt/data/1kb/badger/clog
All compactions in compact log are done.
NOT running any compactions due to DB options.
NOT running any compactions due to DB options.
NOT running any compactions due to DB options.
Seeking at value pointer: {Fid:73 Len:1059 Offset:1041730887}
key=vsz=01024-k=0015263159
l.opt.ValueGCThreshold = 0.0. Exiting runGCInLoop
....................BenchmarkIterateBadgerWithValues/badger-iterate-withvals-2                 1        101801301675 ns/op
--- BENCH: BenchmarkIterateBadgerWithValues/badger-iterate-withvals-2
        bench_test.go:187: [0] Counted 2000000 keys
PASS
ok      github.com/dgraph-io/badger-bench       114.170s

WROTE 5004000 KEYS
        Command being timed: "./populate --kv rocksdb --valsz 16384 --keys_mil 5 --dir /mnt/data/16kb"
        User time (seconds): 1424.16
        System time (seconds): 1397.96
        Percent of CPU this job got: 57%
        Elapsed (wall clock) time (h:mm:ss or m:ss): 1:22:21
        Average shared text size (kbytes): 0
        Average unshared data size (kbytes): 0
        Average stack size (kbytes): 0
        Average total size (kbytes): 0
        Maximum resident set size (kbytes): 1224612
        Average resident set size (kbytes): 0
        Major (requiring I/O) page faults: 87
        Minor (reclaiming a frame) page faults: 6938621
        Voluntary context switches: 4541903
        Involuntary context switches: 1035841
        Swaps: 0
        File system inputs: 1141303472
        File system outputs: 1925444544
        Socket messages sent: 0
        Socket messages received: 0
        Signals delivered: 0
        Page size (bytes): 4096
        Exit status: 0

$ du -sh rocks
52G

WROTE 5004000 KEYS
        Command being timed: "./populate --kv badger --valsz 16384 --keys_mil 5 --dir /mnt/data/16kb"
        User time (seconds): 368.05
        System time (seconds): 113.57
        Percent of CPU this job got: 116%
        Elapsed (wall clock) time (h:mm:ss or m:ss): 6:55.01
        Average shared text size (kbytes): 0
        Average unshared data size (kbytes): 0
        Average stack size (kbytes): 0
        Average total size (kbytes): 0
        Maximum resident set size (kbytes): 2313908
        Average resident set size (kbytes): 0
        Major (requiring I/O) page faults: 55
        Minor (reclaiming a frame) page faults: 6128182
        Voluntary context switches: 2327323
        Involuntary context switches: 206230
        Swaps: 0
        File system inputs: 16424
        File system outputs: 161245240
        Socket messages sent: 0
        Socket messages received: 0
        Signals delivered: 0
        Page size (bytes): 4096
        Exit status: 0

$ du -shc badger/*.sst
105M
$ du -shc badger/*.vlog
77G

$ go test -v --bench BenchmarkReadRandomRocks --keys_mil 5 --valsz 16384 --dir "/mnt/data/16kb" --timeout 10m --benchtime 3m
BenchmarkReadRandomRocks/read-random-rocks-2            SIGQUIT: quit
PC=0x460ed9 m=0 sigcode=0

goroutine 0 [idle]:
runtime.epollwait(0x4, 0x7ffdbcaf5ad8, 0xffffffff00000080, 0x0, 0xffffffff00000000, 0x0, 0x0, 0x0, 0x0, 0x0, ...)
        /usr/local/go/src/runtime/sys_linux_amd64.s:560 +0x19
runtime.netpoll(0xc420029301, 0xc420028001)
        /usr/local/go/src/runtime/netpoll_epoll.go:67 +0x91
runtime.findrunnable(0xc420029300, 0x0)
        /usr/local/go/src/runtime/proc.go:2084 +0x31f
runtime.schedule()
        /usr/local/go/src/runtime/proc.go:2222 +0x14c
runtime.park_m(0xc420001040)
        /usr/local/go/src/runtime/proc.go:2285 +0xab
runtime.mcall(0x7ffdbcaf6240)
        /usr/local/go/src/runtime/asm_amd64.s:269 +0x5b
*** Test killed: ran too long (11m0s).
FAIL    github.com/dgraph-io/badger-bench       674.142s

NOTE: RocksDB took too much memory when doing random lookups. So, this crash happened multiple times.

$ go test -v --bench BenchmarkReadRandomRocks --keys_mil 5 --valsz 16384 --dir "/mnt/data/16kb" --timeout 10m --benchtime 1m
BenchmarkReadRandomRocks/read-random-rocks-2              300000            215171 ns/op
--- BENCH: BenchmarkReadRandomRocks/read-random-rocks-2
        bench_test.go:93: rocks 100391 keys had valid values.
        bench_test.go:93: rocks 150850 keys had valid values.
        bench_test.go:93: rocks 149150 keys had valid values.
PASS
ok      github.com/dgraph-io/badger-bench       121.391s

$ go test -v --bench BenchmarkReadRandomBadger --keys_mil 5 --valsz 16384 --dir "/mnt/data/16kb" --timeout 10m --benchtime 1m
Called BenchmarkReadRandomBadger
Replaying compact log: /mnt/data/16kb/badger/clog
All compactions in compact log are done.
NOT running any compactions due to DB options.
NOT running any compactions due to DB options.
NOT running any compactions due to DB options.
Seeking at value pointer: {Fid:76 Len:16419 Offset:554157669}
l.opt.ValueGCThreshold = 0.0. Exiting runGCInLoop
key=vsz=16384-k=0002454321
BenchmarkReadRandomBadger/read-random-badger-2           2000000             40178 ns/op
--- BENCH: BenchmarkReadRandomBadger/read-random-badger-2
        bench_test.go:73: badger 315956 keys had valid values.
        bench_test.go:73: badger 316468 keys had valid values.
        bench_test.go:73: badger 632648 keys had valid values.
        bench_test.go:73: badger 631790 keys had valid values.
Sending signal to 0 registered with name "value-gc"
Sending signal to 1 registered with name "writes"
--->> Size of bloom filter: 116
=======> Deallocating skiplist
Level "writes" already got signal
Sending signal to 0 registered with name "memtable"
Level "value-gc" already got signal
PASS
ok      github.com/dgraph-io/badger-bench       123.227s

$ go test -v --bench BenchmarkIterate --keys_mil 5 --valsz 16384 --dir "/mnt/data/16kb" --timeout 60m        
BenchmarkIterateRocks/rocksdb-iterate-2                        1        133313688657 ns/op
--- BENCH: BenchmarkIterateRocks/rocksdb-iterate-2
        bench_test.go:129: [0] Counted 2000001 keys
Replaying compact log: /mnt/data/16kb/badger/clog
All compactions in compact log are done.
NOT running any compactions due to DB options.
NOT running any compactions due to DB options.
NOT running any compactions due to DB options.
Seeking at value pointer: {Fid:76 Len:16419 Offset:554157669}
l.opt.ValueGCThreshold = 0.0. Exiting runGCInLoop
key=vsz=16384-k=0002454321
BenchmarkIterateBadgerOnlyKeys/badger-iterate-onlykeys-2                       3         475018676 ns/op
--- BENCH: BenchmarkIterateBadgerOnlyKeys/badger-iterate-onlykeys-2
        bench_test.go:157: [0] Counted 2000001 keys
        bench_test.go:157: [0] Counted 2000001 keys
        bench_test.go:157: [1] Counted 2000001 keys
        bench_test.go:157: [0] Counted 2000001 keys
        bench_test.go:157: [1] Counted 2000001 keys
        bench_test.go:157: [2] Counted 2000001 keys
Replaying compact log: /mnt/data/16kb/badger/clog
All compactions in compact log are done.
NOT running any compactions due to DB options.
NOT running any compactions due to DB options.
NOT running any compactions due to DB options.
Seeking at value pointer: {Fid:76 Len:16419 Offset:554157669}
l.opt.ValueGCThreshold = 0.0. Exiting runGCInLoop
key=vsz=16384-k=0002454321
....................BenchmarkIterateBadgerWithValues/badger-iterate-withvals-2                 1        125095134637 ns/op
--- BENCH: BenchmarkIterateBadgerWithValues/badger-iterate-withvals-2
        bench_test.go:188: [0] Counted 2000000 keys
PASS
ok      github.com/dgraph-io/badger-bench       264.244s


16 Byte values

WROTE 1000008000 KEYS
        Command being timed: "./populate --kv rocksdb --valsz 16 --keys_mil 1000 --dir /mnt/data/16"
        User time (seconds): 8515.35
        System time (seconds): 468.95
        Percent of CPU this job got: 151%
        Elapsed (wall clock) time (h:mm:ss or m:ss): 1:38:40
        Average shared text size (kbytes): 0
        Average unshared data size (kbytes): 0
        Average stack size (kbytes): 0
        Average total size (kbytes): 0
        Maximum resident set size (kbytes): 607276
        Average resident set size (kbytes): 0
        Major (requiring I/O) page faults: 41
        Minor (reclaiming a frame) page faults: 2068722
        Voluntary context switches: 17535643
        Involuntary context switches: 10364349
        Swaps: 0
        File system inputs: 22217320
        File system outputs: 491495256
        Socket messages sent: 0
        Socket messages received: 0
        Signals delivered: 0
        Page size (bytes): 4096
        Exit status: 0

