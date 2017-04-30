package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"runtime/pprof"
	"testing"

	"github.com/dgraph-io/badger/badger"
	"github.com/dgraph-io/badger/table"
	"github.com/dgraph-io/badger/y"
	"github.com/dgraph-io/dgraph/store"
)

var (
	ctx     = context.Background()
	numKeys = flag.Float64("keys_mil", 10.0, "How many million keys to write.")
)

const Mi int = 1000000
const Mf float64 = 1000000

func getStores() (*badger.KV, *store.Store) {
	opt := badger.DefaultOptions
	opt.MapTablesTo = table.LoadToRAM
	opt.Verbose = false
	opt.Dir = *flagDir + "/badger"
	opt.DoNotCompact = true
	opt.ValueGCThreshold = 0.0
	rdir := *flagDir + "/rocks"
	bdb := badger.NewKV(&opt)

	rdb, err := store.NewReadOnlyStore(rdir)
	y.Check(err)
	return bdb, rdb
}

func newKey() []byte {
	k := rand.Int() % int(*numKeys*Mf)
	key := fmt.Sprintf("vsz=%05d-k=%010d", *flagValueSize, k) // 22 bytes.
	return []byte(key)
}

func BenchmarkReadRandom(b *testing.B) {
	ctx := context.Background()

	bdb, rdb := getStores()
	b.Run(fmt.Sprintf("badger-random-reads=%f", *numKeys), func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			var count int
			for pb.Next() {
				key := newKey()
				if val, _ := bdb.Get(ctx, key); val != nil {
					count++
				}
			}
			if count > 100000 {
				b.Logf("badger %d keys had valid values.", count)
			}
		})
	})

	b.Run(fmt.Sprintf("rocksdb-random-reads=%f", *numKeys), func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			var count int
			for pb.Next() {
				key := newKey()
				if _, err := rdb.Get(key); err == nil {
					count++
				}
			}
			if count > 100000 {
				b.Logf("rocks %d keys had valid values.", count)
			}
		})
	})
}

func BenchmarkIterate(b *testing.B) {
	bdb, rdb := getStores()
	b.ResetTimer()

	f, err := os.Create("cpu.prof")
	if err != nil {
		b.Fatalf("Error: %v", err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	b.Run("rocksdb-iterate", func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			itr := rdb.NewIterator()
			var count int
			for itr.SeekToFirst(); itr.Valid(); itr.Next() {
				itr.Key().Data()
				itr.Value().Data()
				count++
			}
			b.Logf("[%d] Counted %d keys\n", j, count)
		}
	})

	b.Run("badger-iterate-onlykeys", func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			var count int
			// 100 = size, 0 = num workers, false = fwd direction.
			itr := bdb.NewIterator(context.Background(), 10000, 0, false)
			itr.Rewind()
			for item := range itr.Ch() {
				if item.Key() == nil {
					break
				}
				count++
				itr.Recycle(item)
			}
			b.Logf("[%d] Counted %d keys\n", j, count)
		}
	})

	b.Run("badger-iterate-withvals", func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			var count int
			itr := bdb.NewIterator(context.Background(), 10000, 100, false)
			itr.Rewind()
			for item := range itr.Ch() {
				if item.Key() == nil {
					break
				}
				item.Value()
				count++
				itr.Recycle(item)
			}
			b.Logf("[%d] Counted %d keys\n", j, count)
		}
	})
}

func BenchmarkWriteBatchRandom(b *testing.B) {
	ctx := context.Background()

	bd := new(BadgerAdapter)
	bd.Init("/tmp/bench-tmp")
	defer bd.Close()

	rd := new(RocksDBAdapter)
	rd.Init("/tmp/bench-tmp")
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

func TestMain(m *testing.M) {
	flag.Parse()
	// call flag.Parse() here if TestMain uses flags
	go http.ListenAndServe(":8080", nil)
	os.Exit(m.Run())
}
