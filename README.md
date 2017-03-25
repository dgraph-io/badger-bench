# badger-bench
Benchmarks of Badger

```
function run() {
	DB=$1
	TEST=$2

	./badger-bench -bench writerandom \
	-num 1000000 \
	-value_size 10000 \
	-db  ${DB} \
	-cpu_profile ${DB}_${TEST}.pprof
	
	go tool pprof -svg badger-bench ${DB}_${TEST}.pprof > ${DB}_${TEST}.svg
}

run badger writerandom
run rocksdb writerandom
```