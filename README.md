# badger-bench
Benchmarks of Badger

```
export VERBOSE=false
export BATCHSIZE=100
export DIR=/tmp/badger_bench
export DIRLOW=/tmp/ramdisk

function run() {
	DB=$1
	TEST=$2
	WRITES=$3
	READS=$4
	VALUESIZE=$5
	rm -Rf $DIR && mkdir -p $DIR

	./badger-bench \
	-bench $TEST \
	-writes $WRITES \
	-reads $READS \
	-value_size $VALUESIZE \
	-rand_size 0 \
	-batch_size $BATCHSIZE \
	-db $DB \
	-cpu_profile data/${DB}_${TEST}_${WRITES}_${READS}_${VALUESIZE}.pprof \
	-verbose=$VERBOSE \
	-dir $DIR \
	-dir_low_levels $DIRLOW
	
	go tool pprof -svg \
	badger-bench \
	data/${DB}_${TEST}_${WRITES}_${READS}_${VALUESIZE}.pprof \
	> data/${DB}_${TEST}_${WRITES}_${READS}_${VALUESIZE}.svg
}

run badger writerandom 10000000 0 10
run rocksdb writerandom 10000000 0 10

run badger  writerandom 10000000 0 100
run rocksdb writerandom 10000000 0 100

run badger  writerandom 1000000 0 1000
run rocksdb writerandom 1000000 0 1000

run badger  batchwriterandom 100000 0 100
run rocksdb batchwriterandom 100000 0 100

run badger  readrandom 100000 10000000 100 
2017/04/03 20:13:36 bench.go:302: Overall: 19.10s, 57.91Mb/s

run rocksdb readrandom 100000 10000000 100
2017/04/03 20:15:17 bench.go:302: Overall: 15.69s, 70.51Mb/s  // Before value threshold.

run badger  readrandom 100000 10000000 10
2017/04/03 20:11:13 bench.go:302: Overall: 16.45s, 15.07Mb/s

run rocksdb readrandom 100000 10000000 10
2017/04/03 20:16:57 bench.go:302: Overall: 20.93s, 11.85Mb/s

```

