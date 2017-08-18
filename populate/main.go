package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/trace"

	"github.com/bmatsuo/lmdb-go/lmdb"
	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger-bench/store"
	"github.com/dgraph-io/badger/y"
	"github.com/paulbellamy/ratecounter"
	"github.com/pkg/profile"
)

const mil float64 = 1000000

var (
	which     = flag.String("kv", "badger", "Which KV store to use. Options: badger, rocksdb, lmdb")
	numKeys   = flag.Float64("keys_mil", 10.0, "How many million keys to write.")
	valueSize = flag.Int("valsz", 128, "Value size in bytes.")
	dir       = flag.String("dir", "", "Base dir for writes.")
	mode      = flag.String("profile.mode", "", "enable profiling mode, one of [cpu, mem, mutex, block]")
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

var bdb *badger.KV
var rdb *store.Store
var lmdbEnv *lmdb.Env
var lmdbDBI lmdb.DBI

func writeBatch(entries []*badger.Entry) int {
	for _, e := range entries {
		fillEntry(e)
	}

	if bdb != nil {
		bdb.BatchSet(entries)
		for _, e := range entries {
			y.Check(e.Error)
		}
	}

	if rdb != nil {
		rb := rdb.NewWriteBatch()
		defer rb.Destroy()

		for _, e := range entries {
			rb.Put(e.Key, e.Value)
		}
		y.Check(rdb.WriteBatch(rb))
	}

	if lmdbEnv != nil {
		err := lmdbEnv.Update(func(txn *lmdb.Txn) error {
			for _, e := range entries {
				err := txn.Put(lmdbDBI, e.Key, e.Value, 0)
				if err != nil {
					return err
				}
			}
			return nil
		})
		y.Check(err)

	}

	return len(entries)
}

func humanize(n int64) string {
	if n >= 1000000 {
		return fmt.Sprintf("%6.2fM", float64(n)/1000000.0)
	}
	if n >= 1000 {
		return fmt.Sprintf("%6.2fK", float64(n)/1000.0)
	}
	return fmt.Sprintf("%5.2f", float64(n))
}

func main() {
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

	trace.AuthRequest = func(req *http.Request) (any, sensitive bool) {
		return true, true
	}

	nw := *numKeys * mil
	fmt.Printf("TOTAL KEYS TO WRITE: %s\n", humanize(int64(nw)))
	opt := badger.DefaultOptions
	// opt.MapTablesTo = table.Nothing
	opt.Dir = *dir + "/badger"
	opt.ValueDir = opt.Dir
	opt.SyncWrites = false

	var err error

	var init bool

	if *which == "badger" {
		init = true
		fmt.Println("Init Badger")
		y.Check(os.RemoveAll(*dir + "/badger"))
		os.MkdirAll(*dir+"/badger", 0777)
		bdb, err = badger.NewKV(&opt)
		if err != nil {
			log.Fatalf("while opening badger: %v", err)
		}
	} else if *which == "rocksdb" {
		init = true
		fmt.Println("Init Rocks")
		os.RemoveAll(*dir + "/rocks")
		os.MkdirAll(*dir+"/rocks", 0777)
		rdb, err = store.NewStore(*dir + "/rocks")
		y.Check(err)
	} else if *which == "lmdb" {
		init = true
		fmt.Println("Init lmdb")
		os.RemoveAll(*dir + "/lmdb")
		os.MkdirAll(*dir+"/lmdb", 0777)

		lmdbEnv, err = lmdb.NewEnv()
		y.Check(err)
		err = lmdbEnv.SetMaxDBs(1)
		y.Check(err)
		err = lmdbEnv.SetMapSize(1 << 38) // ~273Gb
		y.Check(err)

		err = lmdbEnv.Open(*dir+"/lmdb", lmdb.NoSync, 0777)
		y.Check(err)

		// Acquire handle
		err := lmdbEnv.Update(func(txn *lmdb.Txn) error {
			var err error
			lmdbDBI, err = txn.CreateDBI("bench")
			return err
		})
		y.Check(err)
	} else {
		log.Fatalf("Invalid value for option kv: '%s'", *which)
	}

	if !init {
		log.Fatalf("Invalid arguments. Unable to init any store.")
	}

	rc := ratecounter.NewRateCounter(time.Minute)
	var counter int64
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		var count int64
		t := time.NewTicker(time.Second)
		for {
			select {
			case <-t.C:
				fmt.Printf("[%04d] Write key rate per minute: %s. Total: %s\n",
					count,
					humanize(rc.Rate()),
					humanize(atomic.LoadInt64(&counter)))
				count++
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		if err := http.ListenAndServe("0.0.0.0:8081", nil); err != nil {
			log.Fatalf("While opening http. Error: %v", err)
		}
	}()

	N := 12
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(proc int) {
			entries := make([]*badger.Entry, 1000)
			for i := 0; i < len(entries); i++ {
				e := new(badger.Entry)
				e.Key = make([]byte, 22)
				e.Value = make([]byte, *valueSize)
				entries[i] = e
			}

			var written float64
			for written < nw/float64(N) {
				wrote := float64(writeBatch(entries))

				wi := int64(wrote)
				atomic.AddInt64(&counter, wi)
				rc.Incr(wi)

				written += wrote
			}
			wg.Done()
		}(i)
	}
	// 	wg.Add(1) // Block
	wg.Wait()
	cancel()

	if bdb != nil {
		fmt.Println("closing badger")
		bdb.Close()
	}

	if rdb != nil {
		fmt.Println("closing rocks")
		rdb.Close()
	}

	if lmdbEnv != nil {

		fmt.Println("closing lmdb")
		lmdbEnv.CloseDBI(lmdbDBI)
		lmdbEnv.Close()
	}

	fmt.Printf("\nWROTE %d KEYS\n", atomic.LoadInt64(&counter))
}
