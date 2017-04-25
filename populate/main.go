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

	"golang.org/x/net/trace"

	"github.com/dgraph-io/badger/badger"
	"github.com/dgraph-io/badger/table"
	"github.com/dgraph-io/badger/y"
	"github.com/dgraph-io/dgraph/store"
	"github.com/pkg/profile"
)

const mil float64 = 1000000

var (
	which     = flag.String("kv", "both", "Which KV store to use. Options: both, badger, rocksdb")
	numKeys   = flag.Float64("keys_mil", 10.0, "How many million keys to write.")
	valueSize = flag.Int("valsz", 128, "Value size in bytes.")
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
	e.Offset = 0
}

var ctx = context.Background()
var bdb *badger.KV
var rdb *store.Store

func writeBatch(entries []*badger.Entry) int {
	rb := rdb.NewWriteBatch()
	defer rb.Destroy()

	for _, e := range entries {
		fillEntry(e)
		rb.Put(e.Key, e.Value)
	}
	if bdb != nil {
		y.Check(bdb.Write(ctx, entries))
	}
	if rdb != nil {
		y.Check(rdb.WriteBatch(rb))
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
	opt.SyncWrites = false

	var err error

	if *which == "badger" || *which == "both" {
		fmt.Println("Init Badger")
		y.Check(os.RemoveAll("tmp/badger"))
		os.MkdirAll("tmp/badger", 0777)
		bdb = badger.NewKV(&opt)
	}
	if *which == "rocksdb" || *which == "both" {
		fmt.Println("Init Rocks")
		os.RemoveAll("tmp/rocks")
		os.MkdirAll("tmp/rocks", 0777)
		rdb, err = store.NewStore("tmp/rocks")
		// rdb, err = store.NewSyncStore("tmp/rocks")
		y.Check(err)
	}

	http.HandleFunc("/debug/events", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		trace.RenderEvents(w, req, true)
	})
	http.HandleFunc("/debug/requests", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		trace.Render(w, req, true)
	})

	go http.ListenAndServe("0.0.0.0:8080", nil)

	N := 10
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(proc int) {
			entries := make([]*badger.Entry, 100)
			for i := 0; i < len(entries); i++ {
				e := new(badger.Entry)
				e.Key = make([]byte, 22)
				e.Value = make([]byte, *valueSize)
				entries[i] = e
			}

			var written float64
			for written < nw/float64(N) {
				written += float64(writeBatch(entries))
				if int(written)%100000 == 0 {
					fmt.Printf("[%d] Written %5.2fM key-val pairs\n", proc, written/mil)
				}
			}
			fmt.Printf("[%d] Written %5.2fM key-val pairs\n", proc, written/mil)
			wg.Done()
		}(i)
	}
	// 	wg.Add(1) // Block
	wg.Wait()
	if bdb != nil {
		fmt.Println("closing badger")
		bdb.Close()
	}
	if rdb != nil {
		fmt.Println("closing rocks")
		rdb.Close()
	}
	time.Sleep(10 * time.Second)
}
