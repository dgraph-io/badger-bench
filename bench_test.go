package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"runtime/pprof"
	"testing"
	"time"

	"github.com/dgraph-io/badger/badger"
	"github.com/dgraph-io/badger/value"
	"github.com/dgraph-io/badger/y"
	"github.com/dgraph-io/dgraph/store"
)

func writeBatch(bdb *badger.KV, rdb *store.Store, max int) int {
	wb := rdb.NewWriteBatch()
	entries := make([]value.Entry, 0, 10000)
	for i := 0; i < 10000; i++ {
		v := make([]byte, 10)
		rand.Read(v)
		e := value.Entry{
			Key:   []byte(fmt.Sprintf("%016d", rand.Int()%max)),
			Value: v,
		}
		entries = append(entries, e)
		wb.Put(e.Key, e.Value)
	}
	y.Check(bdb.Write(context.Background(), entries))
	y.Check(rdb.WriteBatch(wb))
	return len(entries)
}

func BenchmarkIterate(b *testing.B) {
	opt := badger.DefaultOptions
	opt.Verbose = true
	dir, err := ioutil.TempDir("tmp", "badger")
	Check(err)
	opt.Dir = dir
	bdb := badger.NewKV(&opt)

	dir, err = ioutil.TempDir("tmp", "rocks")
	Check(err)
	rdb, err := store.NewSyncStore(dir)
	Check(err)

	nw := 10000000
	for written := 0; written < nw; {
		written += writeBatch(bdb, rdb, nw*10)
	}
	bdb.Close()
	rdb.Close()
	b.Log("Sleeping for 10 seconds to allow compaction.")
	time.Sleep(time.Second)

	opt.DoNotCompact = true
	bdb = badger.NewKV(&opt)
	rdb, err = store.NewSyncStore(dir)
	Check(err)
	b.ResetTimer()

	f, err := os.Create("cpu.prof")
	if err != nil {
		b.Fatalf("Error: %v", err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	b.Run(fmt.Sprintf("badger-onlykeys-writes=%d", nw), func(b *testing.B) {
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

	b.Run(fmt.Sprintf("badger-withvals-writes=%d", nw), func(b *testing.B) {
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

	b.Run(fmt.Sprintf("rocksdb-writes=%d", nw), func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			itr := rdb.NewIterator()
			var count int
			for itr.SeekToFirst(); itr.Valid(); itr.Next() {
				// To make it equivalent of what Badger iterator does,
				// we allocate memory for both key and value.
				key := make([]byte, itr.Key().Size())
				copy(key, itr.Key().Data())
				val := make([]byte, itr.Value().Size())
				copy(val, itr.Value().Data())
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
	// call flag.Parse() here if TestMain uses flags
	go http.ListenAndServe(":8080", nil)
	os.Exit(m.Run())
}
