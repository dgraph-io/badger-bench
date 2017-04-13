package main

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
)

func BenchmarkWriteBatchRandom(b *testing.B) {
	ctx := context.Background()

	bd := new(BadgerAdapter)
	bd.Init("bench-tmp")
	defer bd.Close()

	rd := new(RocksDBAdapter)
	rd.Init("bench-tmp")
	defer rd.Close()

	batchSize := 1000
	valSizes := []int{100, 1000, 10000, 100000}

	for i := 0; i < 2; i++ {
		var db Database
		name := "badger"
		db = bd
		if i == 1 {
			db = rd
			name = "rocksdb"
		}
		for _, vsz := range valSizes {
			b.Run(fmt.Sprintf("db=%s valuesize=%d", name, vsz), func(b *testing.B) {
				b.RunParallel(func(pb *testing.PB) {
					keys := make([][]byte, batchSize)
					vals := make([][]byte, batchSize)
					for pb.Next() {
						for j := 0; j < batchSize; j++ {
							keys[j] = []byte(fmt.Sprintf("%016d", rand.Int()))
							vals[j] = make([]byte, vsz)
							rand.Read(vals[j])
						}
						db.BatchPut(ctx, keys, vals)
					}
				})
			})
		}
	}
}
