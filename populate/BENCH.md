## Badger: 128 byte value

```
Command being timed: "./populate --kv badger --keys_mil 150 --valsz 128"
User time (seconds): 7213.14
System time (seconds): 402.77
Percent of CPU this job got: 358%
Elapsed (wall clock) time (h:mm:ss or m:ss): 35:24.51
Average shared text size (kbytes): 0
Average unshared data size (kbytes): 0
Average stack size (kbytes): 0
Average total size (kbytes): 0
Maximum resident set size (kbytes): 6332612
Average resident set size (kbytes): 0
Major (requiring I/O) page faults: 3
Minor (reclaiming a frame) page faults: 122765948
Voluntary context switches: 2653765
Involuntary context switches: 969572
Swaps: 0
File system inputs: 1770920
File system outputs: 112093832
Socket messages sent: 0
Socket messages received: 0
Signals delivered: 0
Page size (bytes): 4096
Exit status: 0
```

Disk usage:
3.0G for LSM tree (.sst files), and 22G for value log.

```
$ du -shc * 
1.1G    000000.vlog
1.1G    000001.vlog
1.1G    000003.vlog
1.1G    000004.vlog
1.1G    000005.vlog
1.1G    000006.vlog
1.1G    000007.vlog
1.1G    000008.vlog
1.1G    000009.vlog
1.1G    000010.vlog
1.1G    000011.vlog
1.1G    000012.vlog
1.1G    000013.vlog
1.1G    000014.vlog
1.1G    000015.vlog
1.1G    000016.vlog
1.1G    000017.vlog
1.1G    000018.vlog
1.1G    000019.vlog
1.1G    000020.vlog
1.1G    000021.vlog
756M    000022.vlog
65M     000513.sst
29M     000514.sst
65M     000518.sst
65M     000519.sst
24M     000520.sst
65M     000539.sst
65M     000540.sst
31M     000541.sst
65M     000559.sst
65M     000560.sst
65M     000561.sst
65M     000601.sst
65M     000602.sst
65M     000603.sst
65M     000604.sst
252K    000605.sst
65M     000618.sst
65M     000619.sst
65M     000620.sst
65M     000621.sst
32M     000622.sst
65M     000626.sst
65M     000627.sst
65M     000629.sst
33M     000630.sst
65M     000632.sst
65M     000633.sst
65M     000634.sst
65M     000635.sst
65M     000636.sst
65M     000637.sst
65M     000638.sst
65M     000639.sst
65M     000640.sst
65M     000641.sst
17M     000642.sst
65M     000644.sst
65M     000645.sst
65M     000646.sst
65M     000647.sst
65M     000648.sst
65M     000649.sst
65M     000650.sst
65M     000651.sst
65M     000652.sst
65M     000653.sst
65M     000654.sst
65M     000655.sst
41M     000656.sst
52M     000659.sst
53M     000660.sst
50M     000661.sst
12K     clog
25G     total
```
