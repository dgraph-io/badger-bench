package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/dgraph-io/badger/badger"
	"github.com/dgraph-io/badger/value"
	"github.com/dgraph-io/badger/y"
	"github.com/dgraph-io/dgraph/store"
)

const mil int = 1000000
const nw int = 1000 * mil

func newKey() (key []byte, pow int) {
	pow = 10 // 1KB
	if rand.Intn(2) == 1 {
		pow = 14 // 16KB
	}
	// pow = 4 + rand.Intn(21) // 2^4 = 16B -> 2^24 = 16 MB
	k := rand.Int() % nw
	key = []byte(fmt.Sprintf("v=%02d-k=%016d", pow, k))
	return key, pow
}

func newValue(pow int) []byte {
	sz := 2 ^ pow
	v := make([]byte, sz)
	rand.Read(v)
	return v
}

var ctx = context.Background()

func writeBatch(bdb *badger.KV, rdb *store.Store) int {
	wb := rdb.NewWriteBatch()
	entries := make([]value.Entry, 0, 10000)
	for i := 0; i < 10000; i++ {
		key, pow := newKey()
		v := newValue(pow)
		e := value.Entry{
			Key:   key,
			Value: v,
		}
		entries = append(entries, e)
		wb.Put(e.Key, e.Value)
	}
	y.Check(bdb.Write(ctx, entries))
	y.Check(rdb.WriteBatch(wb))
	return len(entries)
}

func main() {
	opt := badger.DefaultOptions
	opt.Verbose = true
	opt.Dir = "tmp/badger"
	rdir := "tmp/rocks"

	if _, err := os.Stat("tmp/generated"); os.IsExist(err) {
		fmt.Println("tmp/generated already exists. Not populating.")
	}

	os.MkdirAll("tmp/badger", 0777)
	os.MkdirAll("tmp/rocks", 0777)

	bdb := badger.NewKV(&opt)
	rdb, err := store.NewSyncStore(rdir)
	y.Check(err)

	N := 10
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(proc int) {
			for written := 0; written < nw/N; {
				written += writeBatch(bdb, rdb)
				if written%mil == 0 {
					fmt.Printf("[%d] Written %dM key-val pairs\n", proc, written/mil)
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	bdb.Close()
	rdb.Close()
	time.Sleep(10 * time.Second)
	_, err = os.Create("tmp/generated")
	y.Check(err)
}
