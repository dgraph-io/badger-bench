package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sync/atomic"
	"testing"

	"github.com/bmatsuo/lmdb-go/lmdb"
	"github.com/boltdb/bolt"
	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger-bench/store"
	"github.com/dgraph-io/badger/options"
	"github.com/dgraph-io/badger/y"
)

var (
	ctx           = context.Background()
	numKeys       = flag.Float64("keys_mil", 10.0, "How many million keys to write.")
	flagDir       = flag.String("dir", "bench-tmp", "Where data is temporarily stored.")
	flagValueSize = flag.Int("valsz", 128, "Size of each value.")
)

const Mi int = 1000000
const Mf float64 = 1000000

func getBadger() (*badger.DB, error) {
	opt := badger.DefaultOptions
	opt.TableLoadingMode = options.LoadToRAM
	opt.Dir = *flagDir + "/badger"
	opt.ValueDir = opt.Dir
	opt.DoNotCompact = true
	return badger.Open(opt)
}

func getRocks() *store.Store {
	rdb, err := store.NewReadOnlyStore(*flagDir + "/rocks")
	y.Check(err)
	return rdb
}

func getLmdb() *lmdb.Env {
	lmdbEnv, err := lmdb.NewEnv()
	y.Check(err)
	err = lmdbEnv.SetMaxReaders(math.MaxInt64)
	y.Check(err)
	err = lmdbEnv.SetMaxDBs(1)
	y.Check(err)
	err = lmdbEnv.SetMapSize(1 << 38) // ~273Gb
	y.Check(err)

	err = lmdbEnv.Open(*flagDir+"/lmdb", lmdb.Readonly|lmdb.NoReadahead, 0777)
	y.Check(err)
	return lmdbEnv
}

func getBoltDB() *bolt.DB {
	opts := bolt.DefaultOptions
	opts.ReadOnly = true
	boltdb, err := bolt.Open(*flagDir+"/bolt/bolt.db", 0777, opts)
	y.Check(err)
	return boltdb
}

func newKey() []byte {
	k := rand.Int() % int(*numKeys*Mf)
	key := fmt.Sprintf("vsz=%05d-k=%010d", *flagValueSize, k) // 22 bytes.
	return []byte(key)
}

func print(count int) {
	if count%100000 == 0 {
		fmt.Printf(".")
	} else if count%Mi == 0 {
		fmt.Printf("-")
	}
}

type hitCounter struct {
	found    uint64
	notFound uint64
	errored  uint64
}

func (h *hitCounter) Reset() {
	h.found, h.notFound, h.errored = 0, 0, 0
}

func (h *hitCounter) Update(c *hitCounter) {
	atomic.AddUint64(&h.found, c.found)
	atomic.AddUint64(&h.notFound, c.notFound)
	atomic.AddUint64(&h.errored, c.errored)
}

func (h *hitCounter) Print(storeName string, b *testing.B) {
	b.Logf("%s: %d keys had valid values.", storeName, h.found)
	b.Logf("%s: %d keys had no values", storeName, h.notFound)
	b.Logf("%s: %d keys had errors", storeName, h.errored)
	b.Logf("%s: %d total keys looked at", storeName, h.found+h.notFound+h.errored)
	b.Logf("%s: hit rate : %.2f", storeName, float64(h.found)/float64(h.found+h.notFound+h.errored))
}

// A generic read benchmark that runs the doBench func for a specific key value store,
// aggregates the hit counts and prints them out.
func runRandomReadBenchmark(b *testing.B, storeName string, doBench func(*hitCounter, *testing.PB)) {
	counter := &hitCounter{}
	b.Run("read-random"+storeName, func(b *testing.B) {
		counter.Reset()
		b.RunParallel(func(pb *testing.PB) {
			c := &hitCounter{}
			doBench(c, pb)
			counter.Update(c)
		})
	})
	counter.Print(storeName, b)
}

func BenchmarkReadRandomBadger(b *testing.B) {
	bdb, err := getBadger()
	y.Check(err)
	defer bdb.Close()

	runRandomReadBenchmark(b, "badger", func(c *hitCounter, pb *testing.PB) {
		var item *badger.Item
		var err error
		for pb.Next() {
			key := newKey()
			txn := bdb.NewTransaction(false)
			item, err = txn.Get(key)
			if err != nil {
				c.errored++
			} else {
				val, err := item.Value()

				if err != nil {
					c.errored++
					continue
				}

				if val == nil {
					c.notFound++
					continue
				}
				y.AssertTruef(len(val) == *flagValueSize,
					"Assertion failed. value size is %d, expected %d", len(val), *flagValueSize)
				c.found++
			}
		}
	})
}

func BenchmarkReadRandomRocks(b *testing.B) {
	rdb := getRocks()
	defer rdb.Close()

	runRandomReadBenchmark(b, "rocksdb", func(c *hitCounter, pb *testing.PB) {
		for pb.Next() {
			key := newKey()
			// FIXME lookup API and increment accordingly
			if _, err := rdb.Get(key); err == nil {
				c.found++
			}
		}
	})
}

func BenchmarkReadRandomLmdb(b *testing.B) {
	lmdbEnv := getLmdb()
	defer lmdbEnv.Close()

	var lmdbDBI lmdb.DBI
	// Acquire handle
	err := lmdbEnv.View(func(txn *lmdb.Txn) error {
		var err error
		lmdbDBI, err = txn.OpenDBI("bench", 0)
		return err
	})
	y.Check(err)
	defer lmdbEnv.CloseDBI(lmdbDBI)

	runRandomReadBenchmark(b, "lmdb", func(c *hitCounter, pb *testing.PB) {
		err := lmdbEnv.View(func(txn *lmdb.Txn) error {
			txn.RawRead = true
			for pb.Next() {
				key := newKey()
				v, err := txn.Get(lmdbDBI, key)
				if lmdb.IsNotFound(err) {
					c.notFound++
					continue
				} else if err != nil {
					c.errored++
					continue
				}
				y.AssertTruef(len(v) == *flagValueSize, "Assertion failed. value size is %d, expected %d", len(v), *flagValueSize)
				c.found++
			}
			return nil
		})
		if err != nil {
			y.Check(err)
		}
	})
}

func BenchmarkReadRandomBolt(b *testing.B) {
	boltdb := getBoltDB()
	defer boltdb.Close()

	runRandomReadBenchmark(b, "bolt", func(c *hitCounter, pb *testing.PB) {
		err := boltdb.View(func(txn *bolt.Tx) error {
			boltBkt := txn.Bucket([]byte("bench"))
			y.AssertTrue(boltBkt != nil)
			for pb.Next() {
				key := newKey()
				v := boltBkt.Get(key)
				if v == nil {
					c.notFound++
					continue
				}
				y.AssertTruef(len(v) == *flagValueSize, "Assertion failed. value size is %d, expected %d", len(v), *flagValueSize)
				c.found++
			}
			return nil
		})
		y.Check(err)
	})
}

func safecopy(dst []byte, src []byte) []byte {
	if cap(dst) < len(src) {
		dst = make([]byte, len(src))
	}
	dst = dst[0:len(src)]
	copy(dst, src)
	return dst
}

func BenchmarkIterateRocks(b *testing.B) {
	rdb := getRocks()
	defer rdb.Close()
	k := make([]byte, 1024)
	v := make([]byte, Mi)
	b.ResetTimer()

	b.Run("rocksdb-iterate", func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			itr := rdb.NewIterator()
			var count int
			for itr.SeekToFirst(); itr.Valid(); itr.Next() {
				{
					// do some processing.
					k = safecopy(k, itr.Key().Data())
					v = safecopy(v, itr.Value().Data())
				}
				count++
				print(count)
				if count >= 2*Mi {
					break
				}
			}
			b.Logf("[%d] Counted %d keys\n", j, count)
		}
	})
}

func BenchmarkIterateLmdb(b *testing.B) {
	lmdbEnv := getLmdb()
	defer lmdbEnv.Close()

	var lmdbDBI lmdb.DBI
	// Acquire handle
	err := lmdbEnv.View(func(txn *lmdb.Txn) error {
		var err error
		lmdbDBI, err = txn.OpenDBI("bench", 0)
		return err
	})
	y.Check(err)
	defer lmdbEnv.CloseDBI(lmdbDBI)

	k := make([]byte, 1024)
	v := make([]byte, Mi)
	b.ResetTimer()

	b.Run("lmdb-iterate", func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			var count int
			err = lmdbEnv.View(func(txn *lmdb.Txn) error {
				txn.RawRead = true
				cur, err := txn.OpenCursor(lmdbDBI)
				if err != nil {
					return err
				}
				defer cur.Close()

				for {
					k1, v1, err := cur.Get(nil, nil, lmdb.Next)
					if lmdb.IsNotFound(err) {
						return nil
					}
					if err != nil {
						return err
					}

					//fmt.Printf("%s %s\n", k, v)

					y.AssertTruef(len(v1) == *flagValueSize, "Assertion failed. value size is %d, expected %d", len(v1), *flagValueSize)

					// do some processing.
					k = safecopy(k, k1)
					v = safecopy(v, v1)

					count++
					print(count)
					if count >= 2*Mi {
						break
					}
				}
				return nil
			})
			y.Check(err)
			b.Logf("[%d] Counted %d keys\n", j, count)
		}
	})
}

func BenchmarkIterateBolt(b *testing.B) {
	boltdb := getBoltDB()
	defer boltdb.Close()

	k := make([]byte, 1024)
	v := make([]byte, Mi)
	b.ResetTimer()

	b.Run("boltdb-iterate", func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			var count int
			err := boltdb.View(func(txn *bolt.Tx) error {
				boltBkt := txn.Bucket([]byte("bench"))
				y.AssertTrue(boltBkt != nil)
				cur := boltBkt.Cursor()
				for k1, v1 := cur.First(); k1 != nil; k1, v1 = cur.Next() {
					y.AssertTruef(len(v1) == *flagValueSize, "Assertion failed. value size is %d, expected %d", len(v1), *flagValueSize)

					// do some processing.
					k = safecopy(k, k1)
					v = safecopy(v, v1)

					count++
					print(count)
					if count >= 2*Mi {
						break
					}
				}
				return nil
			})
			y.Check(err)
			b.Logf("[%d] Counted %d keys\n", j, count)
		}
	})
}

func BenchmarkIterateBadgerOnlyKeys(b *testing.B) {
	bdb, err := getBadger()
	y.Check(err)
	defer bdb.Close()
	k := make([]byte, 1024)
	b.ResetTimer()

	b.Run("badger-iterate-onlykeys", func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			var count int
			// 100 = size, 0 = num workers, false = fwd direction.
			opt := badger.IteratorOptions{}
			opt.PrefetchSize = 256
			txn := bdb.NewTransaction(false)
			itr := txn.NewIterator(opt)
			for itr.Rewind(); itr.Valid(); itr.Next() {
				item := itr.Item()
				{
					// do some processing.
					k = safecopy(k, item.Key())
				}
				count++
				print(count)
				if count >= 2*Mi {
					break
				}
			}
			b.Logf("[%d] Counted %d keys\n", j, count)
		}
	})
}

func BenchmarkIterateBadgerWithValues(b *testing.B) {
	bdb, err := getBadger()
	y.Check(err)
	defer bdb.Close()
	k := make([]byte, 1024)
	v := make([]byte, Mi)
	b.ResetTimer()

	b.Run("badger-iterate-withvals", func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			var count int
			opt := badger.IteratorOptions{}
			opt.PrefetchSize = 256
			opt.PrefetchValues = true
			txn := bdb.NewTransaction(false)
			itr := txn.NewIterator(opt)
			for itr.Rewind(); itr.Valid(); itr.Next() {
				item := itr.Item()
				val, err := item.Value()
				y.Check(err)

				vsz := len(val)
				y.AssertTruef(vsz == *flagValueSize,
					"Assertion failed. value size is %d, expected %d", vsz, *flagValueSize)
				// do some processing.
				k = safecopy(k, item.Key())
				v = safecopy(v, val)
				count++
				print(count)
				if count >= 2*Mi {
					break
				}
			}
			b.Logf("[%d] Counted %d keys\n", j, count)
		}
	})
}

func TestMain(m *testing.M) {
	flag.Parse()
	// call flag.Parse() here if TestMain uses flags
	go http.ListenAndServe(":8080", nil)
	os.Exit(m.Run())
}
