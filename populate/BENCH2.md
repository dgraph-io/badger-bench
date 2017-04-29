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


        Command being timed: "./populate --kv badger --valsz 128 --keys_mil 250"
        User time (seconds): 8678.95
        System time (seconds): 655.35
        Percent of CPU this job got: 179%
        Elapsed (wall clock) time (h:mm:ss or m:ss): 1:26:27
        Average shared text size (kbytes): 0
        Average unshared data size (kbytes): 0
        Average stack size (kbytes): 0
        Average total size (kbytes): 0
        Maximum resident set size (kbytes): 8050912
        Average resident set size (kbytes): 0
        Major (requiring I/O) page faults: 104
        Minor (reclaiming a frame) page faults: 3528460
        Voluntary context switches: 4176458
        Involuntary context switches: 2220623
        Swaps: 0
        File system inputs: 27441928
        File system outputs: 337732832
        Socket messages sent: 0
        Socket messages received: 0
        Signals delivered: 0
        Page size (bytes): 4096
        Exit status: 0

$ du -sh /mnt/data/badger
40G     /mnt/data/badger
5.5G *.sst  # LSM tree, can be kept in RAM.

Random Reads: Badger is 3.8x faster (no bloom filters in Badger yet)

$ go test --bench BenchmarkReadRandom --keys_mil 250 --valsz 128 --dir "/mnt/data" --timeout 30m --benchtime 3m
Replaying compact log: /mnt/data/badger/clog
All compactions in compact log are done.
NOT running any compactions due to DB options.
NOT running any compactions due to DB options.
NOT running any compactions due to DB options.
Seeking at value pointer: {Fid:40 Len:159 Offset:482792052}
l.opt.ValueGCThreshold = 0.0. Exiting runGCInLoop
key=vsz=00128-k=0097023306
BenchmarkReadRandom/badger-random-reads=250.000000-2            10000000             40108 ns/op
--- BENCH: BenchmarkReadRandom/badger-random-reads=250.000000-2
        bench_test.go:62: badger 324642 keys had valid values.
        bench_test.go:62: badger 324473 keys had valid values.
        bench_test.go:62: badger 3244587 keys had valid values.
        bench_test.go:62: badger 3245332 keys had valid values.
BenchmarkReadRandom/rocksdb-random-reads=250.000000-2            2000000            152355 ns/op
--- BENCH: BenchmarkReadRandom/rocksdb-random-reads=250.000000-2
        bench_test.go:77: rocks 149828 keys had valid values.
        bench_test.go:77: rocks 150172 keys had valid values.
        bench_test.go:77: rocks 999335 keys had valid values.
        bench_test.go:77: rocks 1000665 keys had valid values.
PASS
ok      github.com/dgraph-io/badger-bench       820.758s

