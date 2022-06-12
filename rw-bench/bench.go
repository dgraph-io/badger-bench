package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/y"
)

var (
	numKeys   = flag.Int("keys_mil", 1, "How many million keys to write.")
	valueSize = flag.Int("valsz", 0, "Value size in bytes.")
	//mil        = 1000000
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile `file`")
	memprofile = flag.String("memprofile", "", "write memory profile to `file`")
)

type entry struct {
	Key   []byte
	Value []byte
	Meta  byte
}

/*func fillEntry(e *entry) {
	k := rand.Intn(*numKeys * mil * 10)
	key := fmt.Sprintf("vsz=%05d-k=%010d", *valueSize, k) // 22 bytes.
	if cap(e.Key) < len(key) {
		e.Key = make([]byte, 2*len(key))
	}
	e.Key = e.Key[:len(key)]
	copy(e.Key, key)

	rand.Read(e.Value)
	e.Meta = 0
}*/

var bdg *badger.DB

//var rocks *store.Store

/* func createEntries(entries []*entry) *rdb.WriteBatch {
	rb := rocks.NewWriteBatch()
	for _, e := range entries {
		fillEntry(e)
		rb.Put(e.Key, e.Value)
	}
	return rb
}*/

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	rand.Seed(time.Now().Unix())
	opt := badger.DefaultOptions("tmp/badger")
	// opt.MapTablesTo = table.Nothing
	opt.SyncWrites = false

	var err error
	y.Check(os.RemoveAll("tmp/badger"))
	os.MkdirAll("tmp/badger", 0777)
	bdg, err = badger.Open(opt)
	y.Check(err)

	/*y.Check(os.RemoveAll("tmp/rocks"))
	os.MkdirAll("tmp/rocks", 0777)
	rocks, err = store.NewStore("tmp/rocks")
	y.Check(err)*/

	entries := make([]*entry, *numKeys*1000000)
	for i := 0; i < len(entries); i++ {
		e := new(entry)
		e.Key = make([]byte, 22)
		e.Value = make([]byte, *valueSize)
		entries[i] = e
	}
	//rb := createEntries(entries)
	txn := bdg.NewTransaction(true)
	for _, e := range entries {
		y.Check(txn.Set(e.Key, e.Value))
	}

	fmt.Println("Value size:", *valueSize)
	fmt.Println("RocksDB:")
	//rstart := time.Now()
	//y.Check(rocks.WriteBatch(rb))
	var count int
	//ritr := rocks.NewIterator()
	//ristart := time.Now()
	//for ritr.SeekToFirst(); ritr.Valid(); ritr.Next() {
	//	_ = ritr.Key()
	//	count++
	//}
	//fmt.Println("Num unique keys:", count)
	//fmt.Println("Iteration time: ", time.Since(ristart))
	//fmt.Println("Total time: ", time.Since(rstart))
	//rb.Destroy()
	//rocks.Close()

	fmt.Println("Badger:")
	bstart := time.Now()
	y.Check(txn.Commit())
	iopt := badger.IteratorOptions{}
	bistart := time.Now()
	iopt.PrefetchValues = false
	iopt.PrefetchSize = 1000
	txn = bdg.NewTransaction(false)
	bitr := txn.NewIterator(iopt)
	count = 0
	for bitr.Rewind(); bitr.Valid(); bitr.Next() {
		_ = bitr.Item().Key()
		count++
	}
	fmt.Println("Num unique keys:", count)
	fmt.Println("Iteration time: ", time.Since(bistart))
	fmt.Println("Total time: ", time.Since(bstart))
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		f.Close()
	}
	bdg.Close()
}
