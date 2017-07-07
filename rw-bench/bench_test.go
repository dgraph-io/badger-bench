package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger-bench/rdb"
	"github.com/dgraph-io/badger-bench/store"
	"github.com/dgraph-io/badger/y"
)

var (
	numKeys   = flag.Int("keys_mil", 1, "How many million keys to write.")
	valueSize = flag.Int("valsz", 128, "Value size in bytes.")
	mil       = 10 ^ 6
)

func fillEntry(e *badger.Entry) {
	k := rand.Int() % int(*numKeys*mil)
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

func BenchmarkPutAndIterate(b *testing.B) {
	opt := badger.DefaultOptions
	// opt.MapTablesTo = table.Nothing
	opt.Dir = "tmp/badger"
	opt.ValueDir = opt.Dir
	opt.SyncWrites = false

	var err error
	fmt.Println("Init Badger")
	y.Check(os.RemoveAll("tmp/badger"))
	os.MkdirAll("tmp/badger", 0777)
	bdg, err = badger.NewKV(&opt)
	y.Check(err)

	fmt.Println("Init Rocks")
	os.RemoveAll("tmp/rocks")
	os.MkdirAll("tmp/rocks", 0777)
	rocks, err = store.NewStore("tmp/rocks")
	y.Check(err)

	entries := make([]*badger.Entry, 100000)
	for i := 0; i < len(entries); i++ {
		e := new(badger.Entry)
		e.Key = make([]byte, 22)
		e.Value = make([]byte, *valueSize)
		entries[i] = e
	}

	fmt.Println(len(entries))
	rb := createEntries(entries)

	b.Run("rocksdb-iterate", func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			rocks.WriteBatch(rb)
			itr := rocks.NewIterator()
			for itr.SeekToFirst(); itr.Valid(); itr.Next() {
				_, _ = itr.Key(), itr.Value()
			}
		}
	})

	b.Run("badger-iterate-onlykeys", func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			bdg.BatchSet(entries)
			opt := badger.IteratorOptions{}
			opt.FetchValues = false
			opt.PrefetchSize = 10000
			itr := bdg.NewIterator(opt)
			for itr.Rewind(); itr.Valid(); itr.Next() {
				_ = itr.Item()
			}
		}
	})

	b.Run("badger-iterate-keyVal", func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			bdg.BatchSet(entries)
			opt := badger.IteratorOptions{}
			opt.PrefetchSize = 10000
			itr := bdg.NewIterator(opt)
			for itr.Rewind(); itr.Valid(); itr.Next() {
				_ = itr.Item()
			}
		}
	})
}
