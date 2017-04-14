# Using O_DSYNC when opening file.

$ go test -bench=. -parallel 8 -benchtime 10s                       ~/go/src/github.com/dgraph-io/badger-bench
Dir: bench-tmp
Seeking at offset: 0
Replayed 0 KVs
BenchmarkWriteBatchRandom/db=badger_valuesize=100-4         	Storing offset: 269571452
Num level 0 tables increased from 0 to 1
LOG Compact 0->1: Del [0,1), Del [0,0), Add [0,2), took 3.226299114s
Storing offset: 538758528
Num level 0 tables increased from 0 to 1
    5000	   4340670 ns/op
BenchmarkWriteBatchRandom/db=badger_valuesize=1000-4        	LOG Compact 0->1: Del [0,1), Del [0,2), Add [0,3), took 8.601355666s
    1000	  10765561 ns/op
BenchmarkWriteBatchRandom/db=badger_valuesize=10000-4       	Storing offset: 2957263064
Num level 0 tables increased from 0 to 1
LOG Compact 0->1: Del [0,1), Del [0,3), Add [0,4), took 7.817752443s
     200	  76942350 ns/op
BenchmarkWriteBatchRandom/db=badger_valuesize=100000-4      	      20	 697220392 ns/op
BenchmarkWriteBatchRandom/db=rocksdb_valuesize=100-4        	    2000	   9892488 ns/op
BenchmarkWriteBatchRandom/db=rocksdb_valuesize=1000-4       	     500	  44004765 ns/op
BenchmarkWriteBatchRandom/db=rocksdb_valuesize=10000-4      	     100	 272424093 ns/op
BenchmarkWriteBatchRandom/db=rocksdb_valuesize=100000-4     	      10	2291584086 ns/op
PASS
ok  	github.com/dgraph-io/badger-bench	181.937s

# Iteration using Badger vs RocksDB

BenchmarkIterate/badger-writes=10000000-4         	       3	2869533549 ns/op
--- BENCH: BenchmarkIterate/badger-writes=10000000-4
	bench_test.go:80: [0] Counted 9516069 keys
	bench_test.go:80: [0] Counted 9516069 keys
	bench_test.go:80: [1] Counted 9516069 keys
	bench_test.go:80: [0] Counted 9516069 keys
	bench_test.go:80: [1] Counted 9516069 keys
	bench_test.go:80: [2] Counted 9516069 keys
BenchmarkIterate/rocksdb-writes=10000000-4        	       3	2776771610 ns/op
--- BENCH: BenchmarkIterate/rocksdb-writes=10000000-4
	bench_test.go:91: [0] Counted 9516068 keys
	bench_test.go:91: [0] Counted 9516068 keys
	bench_test.go:91: [1] Counted 9516068 keys
	bench_test.go:91: [0] Counted 9516068 keys
	bench_test.go:91: [1] Counted 9516068 keys
	bench_test.go:91: [2] Counted 9516068 keys

