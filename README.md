# badger-bench
Benchmarks of Badger

```
export VERBOSE=false
export BATCHSIZE=100
export DIR=/tmp/badger_bench

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
	-dir $DIR
	
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
run rocksdb readrandom 100000 10000000 100
```

