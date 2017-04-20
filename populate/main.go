package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sync"
	"time"

	"github.com/dgraph-io/badger/badger"
	"github.com/dgraph-io/badger/table"
	"github.com/dgraph-io/badger/value"
	"github.com/dgraph-io/badger/y"
	"github.com/pkg/profile"
)

const mil float64 = 1000000

var (
	which   = flag.String("kv", "badger", "Which KV store to use.")
	numKeys = flag.Float64("keys_mil", 10.0, "How many million keys to write.")
)

func fillEntry(e *value.Entry) {
	var pow uint = 10 // 1KB
	if rand.Intn(2) == 1 {
		pow = 14 // 16KB
	}
	k := rand.Int() % int(*numKeys*mil)
	key := fmt.Sprintf("v=%02d-k=%016d", pow, k)
	if cap(e.Key) < len(key) {
		e.Key = make([]byte, 2*len(key))
	}
	e.Key = e.Key[:len(key)]
	copy(e.Key, key)

	y.AssertTrue(cap(e.Value) == 1<<14)
	vsz := 1 << pow
	e.Value = e.Value[:vsz]
	rand.Read(e.Value)

	e.Meta = 0
	e.Offset = 0
}

var ctx = context.Background()

func writeBatch(entries []*value.Entry, bdb *badger.KV) int {
	for _, e := range entries {
		fillEntry(e)
	}
	if bdb != nil {
		y.Check(bdb.Write(ctx, entries))
	}
	return len(entries)
}

func main() {
	mode := flag.String("profile.mode", "", "enable profiling mode, one of [cpu, mem, mutex, block]")
	flag.Parse()
	switch *mode {
	case "cpu":
		defer profile.Start(profile.CPUProfile).Stop()
	case "mem":
		defer profile.Start(profile.MemProfile).Stop()
	case "mutex":
		defer profile.Start(profile.MutexProfile).Stop()
	case "block":
		defer profile.Start(profile.BlockProfile).Stop()
	default:
		// do nothing
	}

	nw := *numKeys * mil
	opt := badger.DefaultOptions
	opt.NumMemtables = 3
	opt.MapTablesTo = table.Nothing
	opt.Verbose = true
	opt.Dir = "tmp/badger"
	// rdir := "tmp/rocks"

	// var err error
	var bdb *badger.KV
	// var rdb *store.Store

	if *which == "badger" {
		y.Check(os.RemoveAll("tmp/badger"))
		os.MkdirAll("tmp/badger", 0777)
		bdb = badger.NewKV(&opt)
	}
	// if *which == "rocksdb" {
	// 	os.RemoveAll("tmp/rocks")
	// 	os.MkdirAll("tmp/rocks", 0777)
	// 	rdb, err = store.NewSyncStore(rdir)
	// 	y.Check(err)
	// }

	go http.ListenAndServe(":8080", nil)

	N := 10
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(proc int) {
			entries := make([]*value.Entry, 100)
			for i := 0; i < len(entries); i++ {
				e := new(value.Entry)
				e.Key = make([]byte, 10)
				e.Value = make([]byte, 1<<14)
				entries[i] = e
			}

			var written float64
			for written < nw/float64(N) {
				written += float64(writeBatch(entries, bdb))
				if int(written)%int(mil) == 0 {
					fmt.Printf("[%d] Written %dM key-val pairs\n", proc, written/mil)
				}
			}
			fmt.Printf("[%d] Written %5.2fM key-val pairs\n", proc, written/mil)
			wg.Done()
		}(i)
	}
	// 	wg.Add(1) // Block
	wg.Wait()
	if bdb != nil {
		bdb.Close()
	}
	// if rdb != nil {
	// 	rdb.Close()
	// }
	time.Sleep(10 * time.Second)
}
