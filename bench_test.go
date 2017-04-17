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

var ctx = context.Background()

func getStores() (*badger.KV, *store.Store) {
	opt := badger.DefaultOptions
	opt.MapTablesTo = table.Nothing
	opt.Verbose = true
	opt.Dir = "tmp/badger"
	opt.DoNotCompact = true
	rdir := "tmp/rocks"
	bdb := badger.NewKV(&opt)

	rdb, err := store.NewSyncStore(rdir)
	y.Check(err)
	return bdb, rdb
}

var numKeys = flag.Int("keys_mil", 10, "How many million keys to write.")

const mil int = 1000000

func newKey() (key []byte, pow int) {
	pow = 10 // 1KB
	if rand.Intn(2) == 1 {
		pow = 14 // 16KB
	}
	k := rand.Int() % (*numKeys * mil)
	key = []byte(fmt.Sprintf("v=%02d-k=%016d", pow, k))
	return key, pow
}

func BenchmarkReadRandom(b *testing.B) {
	ctx := context.Background()

	bdb, rdb := getStores()
	b.Run(fmt.Sprintf("badger-random-reads=%d", *numKeys), func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			var count int
			for pb.Next() {
				key, _ := newKey()
				if val := bdb.Get(ctx, key); val != nil {
					count++
				}
			}
			// b.Logf("%d keys had valid values.", count)
		})
	})

	b.Run(fmt.Sprintf("rocksdb-random-reads=%d", *numKeys), func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			var count int
			for pb.Next() {
				key, _ := newKey()
				if _, err := rdb.Get(key); err == nil {
					count++
				}
			}
			// b.Logf("%d keys had valid values.", count)
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

	b.Run("badger-iterate-onlykeys", func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			var count int
			itr := bdb.NewIterator(context.Background(), 100, 0)
			itr.SeekToFirst()
			for item := range itr.Ch() {
				if item.Key() == nil {
					break
				}
				count++
			}
			b.Logf("[%d] Counted %d keys\n", j, count)
		}
	})

	b.Run("badger-iterate-withvals", func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			var count int
			itr := bdb.NewIterator(context.Background(), 100, 100)
			itr.SeekToFirst()
			for item := range itr.Ch() {
				if item.Key() == nil {
					break
				}
				item.Value()
				count++
			}
			b.Logf("[%d] Counted %d keys\n", j, count)
		}
	})

	b.Run("rocksdb-iterate", func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			itr := rdb.NewIterator()
			var count int
			for itr.SeekToFirst(); itr.Valid(); itr.Next() {
				// To make it equivalent of what Badger iterator does,
				// we allocate memory for both key and value.
				key := make([]byte, itr.Key().Size())
				copy(key, itr.Key().Data())
				// val := make([]byte, itr.Value().Size())
				// copy(val, itr.Value().Data())
				count++
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
