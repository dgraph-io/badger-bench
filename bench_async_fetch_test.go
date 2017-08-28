package main

import (
	"testing"

	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/y"
)

func BenchmarkIterateValueFetch(b *testing.B) {
	bdb, err := getBadger()
	y.Check(err)
	defer bdb.Close()
	k := make([]byte, 1024)
	v := make([]byte, Mi)
	b.ResetTimer()

	b.Run("iterate-async-fetch", func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			var count int
			opt := badger.IteratorOptions{}
			opt.PrefetchSize = 256
			opt.FetchValues = true
			itr := bdb.NewIterator(opt)
			for itr.Rewind(); itr.Valid(); itr.Next() {
				item := itr.Item()
				vsz := len(item.Value())
				y.AssertTruef(vsz == *flagValueSize, "Assertion failed. value size is %d, expected %d", vsz, *flagValueSize)
				{
					// do some processing.
					k = safecopy(k, item.Key())
					v = safecopy(v, item.Value())
				}
				count++
				if count >= 2*Mi {
					break
				}
			}
			b.Logf("[%d] Counted %d keys\n", j, count)
		}
	})

	b.Run("iterate-sync-fetch", func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			var count int
			opt := badger.IteratorOptions{}
			opt.FetchValues = false
			itr := bdb.NewIterator(opt)
			for itr.Rewind(); itr.Valid(); itr.Next() {
				item := itr.Item()
				bdb.FillValue(item)
				vsz := len(item.Value())
				y.AssertTruef(vsz == *flagValueSize, "Assertion failed. value size is %d, expected %d", vsz, *flagValueSize)
				{
					// do some processing.
					k = safecopy(k, item.Key())
					v = safecopy(v, item.Value())
				}
				count++
				if count >= 2*Mi {
					break
				}
			}
			b.Logf("[%d] Counted %d keys\n", j, count)
		}
	})
}
