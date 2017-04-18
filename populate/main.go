package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/dgraph-io/badger/badger"
	"github.com/dgraph-io/badger/table"
	"github.com/dgraph-io/badger/value"
	"github.com/dgraph-io/badger/y"
	"github.com/dgraph-io/dgraph/store"
)

const mil float64 = 1000000

var (
	which   = flag.String("kv", "badger", "Which KV store to use.")
	numKeys = flag.Float64("keys_mil", 10.0, "How many million keys to write.")
)

func newKey() (key []byte, pow uint) {
	pow = 10 // 1KB
	if rand.Intn(2) == 1 {
		pow = 14 // 16KB
	}
	// pow = 4 + rand.Intn(21) // 2^4 = 16B -> 2^24 = 16 MB
	k := rand.Int() % int(*numKeys*mil)
	key = []byte(fmt.Sprintf("v=%02d-k=%016d", pow, k))
	return key, pow
}

func newValue(pow uint) []byte {
	sz := 1 << pow
	v := make([]byte, sz)
	rand.Read(v)
	return v
}

var ctx = context.Background()

func writeBatch(bdb *badger.KV, rdb *store.Store) int {
	wb := rdb.NewWriteBatch()
	entries := make([]value.Entry, 0, 10000)
	values := make(map[uint][]byte)
	for i := 0; i < 10000; i++ {
		key, pow := newKey()
		var v []byte
		if val, has := values[pow]; has {
			v = val
		} else {
			v = newValue(pow)
			values[pow] = v
		}
		e := value.Entry{
			Key:   key,
			Value: v,
		}
		entries = append(entries, e)
		wb.Put(e.Key, e.Value)
	}
	if bdb != nil {
		y.Check(bdb.Write(ctx, entries))
	}
	if rdb != nil {
		y.Check(rdb.WriteBatch(wb))
	}
	return len(entries)
}

func main() {
	flag.Parse()

	nw := *numKeys * mil
	opt := badger.DefaultOptions
	opt.MapTablesTo = table.Nothing
	opt.Verbose = true
	opt.Dir = "tmp/badger"
	rdir := "tmp/rocks"

	var err error
	var bdb *badger.KV
	var rdb *store.Store

	if *which == "badger" {
		os.RemoveAll("tmp/badger")
		os.MkdirAll("tmp/badger", 0777)
		bdb = badger.NewKV(&opt)
	}
	if *which == "rocksdb" {
		os.RemoveAll("tmp/rocks")
		os.MkdirAll("tmp/rocks", 0777)
		rdb, err = store.NewSyncStore(rdir)
		y.Check(err)
	}

	N := 10
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(proc int) {
			var written float64
			for written < nw/float64(N) {
				written += float64(writeBatch(bdb, rdb))
				if int(written)%int(mil) == 0 {
					fmt.Printf("[%d] Written %dM key-val pairs\n", proc, written/mil)
				}
			}
			fmt.Printf("[%d] Written %5.2fM key-val pairs\n", proc, written/mil)
			wg.Done()
		}(i)
	}
	wg.Add(1) // Block
	wg.Wait()
	if bdb != nil {
		bdb.Close()
	}
	if rdb != nil {
		rdb.Close()
	}
	f, err := os.Create("m.prof")
	y.Check(err)
	pprof.WriteHeapProfile(f)
	defer f.Close()

	time.Sleep(10 * time.Second)
}
