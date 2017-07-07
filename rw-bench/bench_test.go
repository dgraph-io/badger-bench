package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger-bench/rdb"
	"github.com/dgraph-io/badger-bench/store"
	"github.com/dgraph-io/badger/y"
)

var (
	numKeys   = flag.Int("keys_mil", 3, "How many million keys to write.")
	valueSize = flag.Int("valsz", 1024, "Value size in bytes.")
	mil       = 1000000
)

func fillEntry(e *badger.Entry) {
	k := rand.Intn(*numKeys * mil * 10)
	key := fmt.Sprintf("vsz=%05d-k=%010d", *valueSize, k) // 22 bytes.
	if cap(e.Key) < len(key) {
		e.Key = make([]byte, 2*len(key))
	}
	e.Key = e.Key[:len(key)]
	copy(e.Key, key)

	rand.Read(e.Value)
	e.Meta = 0
}

var bdg *badger.KV
var rocks *store.Store

func createEntries(entries []*badger.Entry) *rdb.WriteBatch {
	rb := rocks.NewWriteBatch()
	for _, e := range entries {
		fillEntry(e)
		rb.Put(e.Key, e.Value)
	}
	return rb
}

func TestPutAndIterate(t *testing.T) {
	rand.Seed(time.Now().Unix())
	opt := badger.DefaultOptions
	// opt.MapTablesTo = table.Nothing
	opt.Dir = "tmp/badger"
	opt.ValueDir = opt.Dir
	opt.SyncWrites = false
	opt.ValueCompressionMinSize = 10000

	var err error
	y.Check(os.RemoveAll("tmp/badger"))
	os.MkdirAll("tmp/badger", 0777)
	bdg, err = badger.NewKV(&opt)
	y.Check(err)

	y.Check(os.RemoveAll("tmp/rocks"))
	os.MkdirAll("tmp/rocks", 0777)
	rocks, err = store.NewStore("tmp/rocks")
	y.Check(err)

	entries := make([]*badger.Entry, *numKeys*1000000)
	for i := 0; i < len(entries); i++ {
		e := new(badger.Entry)
		e.Key = make([]byte, 22)
		e.Value = make([]byte, *valueSize)
		entries[i] = e
	}
	rb := createEntries(entries)

	fmt.Println("Value size:", *valueSize)
	fmt.Println("RocksDB:")
	rstart := time.Now()
	y.Check(rocks.WriteBatch(rb))
	var count int
	ritr := rocks.NewIterator()
	ristart := time.Now()
	for ritr.SeekToFirst(); ritr.Valid(); ritr.Next() {
		_ = ritr.Key()
		count++
	}
	fmt.Println("Num unique keys:", count)
	fmt.Println("Iteration time: ", time.Since(ristart))
	fmt.Println("Total time: ", time.Since(rstart))
	rb.Destroy()
	rocks.Close()

	fmt.Println("Badger:")
	bstart := time.Now()
	y.Check(bdg.BatchSet(entries))
	iopt := badger.IteratorOptions{}
	bistart := time.Now()
	iopt.FetchValues = false
	iopt.PrefetchSize = 1000
	bitr := bdg.NewIterator(iopt)
	count = 0
	for bitr.Rewind(); bitr.Valid(); bitr.Next() {
		_ = bitr.Item().Key()
		count++
	}
	fmt.Println("Num unique keys:", count)
	fmt.Println("Iteration time: ", time.Since(bistart))
	fmt.Println("Total time: ", time.Since(bstart))
	bdg.Close()
}
