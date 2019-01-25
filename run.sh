#!/usr/bin/env bash
#set -x

export LD_LIBRARY_PATH=/usr/local/lib
#keysMil=(250 75 5 1000)
keysMil=(250)
#valueSizes=(128 1024 16384 16)
valueSizes=(128)

for i in "${!keysMil[@]}"; do 
    keyMil=${keysMil[$i]}
    valueSz=${valueSizes[$i]}
    echo "keyMil:$keyMil, valueSz:$valueSz"

    DATADIR=bench-data-$valueSz
    if [ ! -d "$DATADIR" ]; then
        mkdir $DATADIR
    fi

    populate --kv rocksdb --valsz $valueSz --keys_mil $keyMil --dir=$DATADIR | tee logs/populate-rocksdb-$valueSz.log
    populate --kv badger --valsz $valueSz --keys_mil $keyMil --dir=$DATADIR | tee logs/populate-badger-$valueSz.log

    echo "cleaning caches"
    echo 3 |sudo tee /proc/sys/vm/drop_caches
    sudo blockdev --flushbufs  /dev/nvme0n1
    echo "benchmark random read"
    
    go test --bench BenchmarkReadRandomRocks --keys_mil $keyMil --valsz $valueSz --dir $DATADIR --timeout 10m --benchtime 3m -v|tee logs/randomread-rocksdb-$valueSz.log
    go test --bench BenchmarkReadRandomBadger --keys_mil $keyMil --valsz $valueSz --dir $DATADIR --timeout 10m --benchtime 3m -v|tee  logs/randomread-badger-$valueSz.log
    go test --bench BenchmarkIterateRocks --keys_mil $keyMil --valsz $valueSz --dir $DATADIR --timeout 10m --cpuprofile logs/iterate-rocks-cpu-$valueSz.out -v|tee logs/iterate-rocks-$valueSz.log
    go test --bench BenchmarkIterateBadgerOnly --keys_mil $keyMil --valsz $valueSz --dir $DATADIR --timeout 10m --cpuprofile logs/iterate-badger-cpu-$valueSz.out -v|tee logs/iterate-badger-$valueSz.log
    go test --bench BenchmarkIterateBadgerWithValues --keys_mil $keyMil --valsz $valueSz --dir $DATADIR --timeout 10m  --cpuprofile logs/iterate-badger-with-values-cpu-$valueSz.out -v|tee logs/iterate-badger-with-values-$valueSz.log
done

