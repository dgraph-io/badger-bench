# badger-bench
Benchmarks of Badger

```
export VERBOSE=true

function run() {
	DB=$1
	TEST=$2
	NUM=$3
	VALUESIZE=$4

	./badger-bench \
	-bench $TEST \
	-num $NUM \
	-value_size $VALUESIZE \
	-rand_size 0 \
	-db $DB \
	-cpu_profile data/${DB}_${TEST}_${NUM}_${VALUESIZE}.pprof \
	-verbose=$VERBOSE
	
	go tool pprof -svg \
	badger-bench \
	data/${DB}_${TEST}_${NUM}_${VALUESIZE}.pprof \
	> data/${DB}_${TEST}_${NUM}_${VALUESIZE}.svg
}

run badger writerandom 10000000 10
run rocksdb writerandom 10000000 10

run badger  writerandom 10000000 100
run rocksdb writerandom 10000000 100

run badger writerandom 1000000 1000
run rocksdb writerandom 1000000 1000
```

