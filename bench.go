package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime/pprof"
	"sync"
	"time"

	"golang.org/x/net/trace"

	"github.com/dgraph-io/badger/badger"
	"github.com/dgraph-io/badger/value"
	"github.com/pkg/errors"

	"github.com/dgraph-io/dgraph/store"
)

var (
	flagBench      = flag.String("bench", "", "Run which benchmark?")
	flagDB         = flag.String("db", "", "Which DB: rocksdb, badger")
	flagValueSize  = flag.Int("value_size", 100, "Size of each value.")
	flagBatchSize  = flag.Int("batch_size", 100, "Size of writebatch.")
	flagNumWrites  = flag.Int("writes", 50000, "Number of key-value pairs to write.")
	flagNumReads   = flag.Int("reads", 100000, "Number of key-value pairs to read.")
	flagCpuProfile = flag.String("cpu_profile", "", "Write cpu profile to file.")
	flagVerbose    = flag.Bool("verbose", false, "Verbose.")
	flagDir        = flag.String("dir", "bench-tmp", "Where data is temporarily stored.")

	rdbStore *store.Store
)

func Check(err error) {
	if err != nil {
		log.Fatalf("%+v", errors.Wrap(err, "Check error"))
	}
}

func AssertTrue(b bool) {
	if !b {
		log.Fatalf("%+v", errors.Errorf("Assert failed"))
	}
}

func report(d time.Duration, n int) string {
	secs := d.Seconds()
	throughput := float64(n*(16+*flagValueSize)) / ((1 << 20) * secs)
	return fmt.Sprintf("%.2fs, %.2fMb/s", secs, throughput)
}

type Database interface {
	Init(basedir string)
	Close()
	Put(ctx context.Context, key, val []byte)
	BatchPut(ctx context.Context, key, val [][]byte)
	Get(ctx context.Context, key []byte)
}

type RocksDBAdapter struct {
	rdb *store.Store
	dir string
}

func (s *RocksDBAdapter) Init(basedir string) {
	var err error
	s.dir, err = ioutil.TempDir(basedir, "storetest_")
	Check(err)
	s.rdb, err = store.NewSyncStore(s.dir)
	Check(err)
}

func (s *RocksDBAdapter) Close() {
	//	s.rdb.Close()
}

func (s *RocksDBAdapter) Put(ctx context.Context, key, val []byte) {
	s.rdb.SetOne(key, val)
}

func (s *RocksDBAdapter) BatchPut(ctx context.Context, key, val [][]byte) {
	wb := s.rdb.NewWriteBatch()
	AssertTrue(len(key) == len(val))
	for i := 0; i < len(key); i++ {
		wb.Put(key[i], val[i])
	}
	Check(s.rdb.WriteBatch(wb))
}

func (s *RocksDBAdapter) Get(ctx context.Context, key []byte) {
	_, err := s.rdb.Get(key)
	Check(err)
}

type BadgerAdapter struct {
	db  *badger.KV
	dir string
}

func (s *BadgerAdapter) Init(basedir string) {
	opt := badger.DefaultOptions
	opt.Verbose = true
	dir, err := ioutil.TempDir(basedir, "badger")
	Check(err)
	opt.Dir = dir

	fmt.Printf("Dir: %s\n", *flagDir)
	s.db = badger.NewKV(&opt)
}

func (s *BadgerAdapter) Close() {
}

func (s *BadgerAdapter) Put(ctx context.Context, key, val []byte) {
	s.db.Put(ctx, key, val)
}

func (s *BadgerAdapter) BatchPut(ctx context.Context, key, val [][]byte) {
	var entries []value.Entry
	AssertTrue(len(key) == len(val))
	for i := 0; i < len(key); i++ {
		entries = append(entries, value.Entry{
			Key:   key[i],
			Value: val[i],
		})
	}
	Check(s.db.Write(ctx, entries))
}

func (s *BadgerAdapter) Get(ctx context.Context, key []byte) {
	s.db.Get(ctx, key)
}

// No batching.
func WriteRandom(database Database) {
	ctx := context.Background()
	fmt.Println("WriteRandom test")
	timeStart := time.Now()
	timeLog := timeStart
	timeLogI := 0
	// If you use b.N, you might add too few samples and be working only in memory.
	// We need to fix a large number of pairs. This is what LevelDB benchmark does as well.
	for i := 0; i < *flagNumWrites; i++ {
		key := []byte(fmt.Sprintf("%016d", rand.Int()%*flagNumWrites))
		val := make([]byte, *flagValueSize)
		rand.Read(val)
		database.Put(ctx, key, val)
		timeElapsed := time.Since(timeLog)
		if timeElapsed > 5*time.Second {
			log.Printf("%.2f%% : %s\n", float64(i)*100.0/float64(*flagNumWrites),
				report(timeElapsed, i-timeLogI))
			timeLog = time.Now()
			timeLogI = i
		}
	}
	log.Printf("Overall: %s\n", report(time.Since(timeStart), *flagNumWrites))
}

// With batching.
func BatchWriteRandom(database Database) {
	fmt.Println("BatchWriteRandom test")
	timeStart := time.Now()
	timeLog := timeStart
	timeLogI := 0
	keys := make([][]byte, *flagBatchSize)
	vals := make([][]byte, *flagBatchSize)

	for i := 0; i < *flagNumWrites; i++ {
		for j := 0; j < *flagBatchSize; j++ {
			keys[j] = []byte(fmt.Sprintf("%016d", rand.Int()%*flagNumWrites))
			vals[j] = make([]byte, *flagValueSize)
			rand.Read(vals[j])
		}
		tr := trace.New("BatchWrite", "BatchPut")
		ctx := trace.NewContext(context.Background(), tr)
		database.BatchPut(ctx, keys, vals)
		// tr.Finish()
		timeElapsed := time.Since(timeLog)
		if timeElapsed > 5*time.Second {
			log.Printf("%.2f%% : %s\n", float64(i)*100.0/float64(*flagNumWrites),
				report(timeElapsed, (i-timeLogI)*(*flagBatchSize)))
			timeLog = time.Now()
			timeLogI = i
		}
	}
	log.Printf("Overall: %s\n", report(time.Since(timeStart), *flagNumWrites*(*flagBatchSize)))
}

// No batching.
func ReadRandom(database Database) {
	fmt.Println("ReadRandom test")
	ctx := context.Background()

	keys := make([][]byte, *flagBatchSize)
	vals := make([][]byte, *flagBatchSize)
	// Write some key-value pairs first.
	// TODO: Allow user to just specify a database to open.
	timeLog := time.Now()
	log.Printf("Preparing database")
	for i := 0; i < *flagNumWrites; i++ {
		for j := 0; j < *flagBatchSize; j++ {
			keys[j] = []byte(fmt.Sprintf("%016d", rand.Int()%*flagNumWrites))
			vals[j] = make([]byte, *flagValueSize)
			rand.Read(vals[j])
		}
		database.BatchPut(ctx, keys, vals)
		timeElapsed := time.Since(timeLog)
		if timeElapsed > 5*time.Second {
			log.Printf("%.2f percent written\n", float64(i)*100.0/float64(*flagNumWrites))
			timeLog = time.Now()
		}
	}

	log.Printf("Slight pause\n")
	time.Sleep(time.Second)

	log.Printf("Starting reads")
	timeLogI := 0
	timeStart := time.Now()
	timeLog = timeStart
	for i := 0; i < *flagNumReads; i++ {
		key := []byte(fmt.Sprintf("%016d", rand.Int()%*flagNumReads))
		database.Get(ctx, key)
		timeElapsed := time.Since(timeLog)
		if timeElapsed > 5*time.Second {
			log.Printf("%.2f percent: %s\n", float64(i)*100.0/float64(*flagNumReads),
				report(timeElapsed, (i-timeLogI)))
			timeLog = time.Now()
			timeLogI = i
		}
	}
	log.Printf("Overall: %s\n", report(time.Since(timeStart), *flagNumReads))
}

func main() {
	flag.Parse()
	AssertTrue(len(*flagBench) > 0)

	if *flagCpuProfile != "" {
		f, err := os.Create(*flagCpuProfile)
		if err != nil {
			log.Fatalf("Profiler error: %v", err)
		}
		fmt.Printf("CPU profiling started")
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		http.ListenAndServe(":8080", nil)
		wg.Done()
	}()

	var database Database
	switch *flagDB {
	case "badger":
		database = new(BadgerAdapter)
	case "rocksdb":
		database = new(RocksDBAdapter)
	default:
		log.Fatalf("Database invalid: %v", *flagDB)
	}
	database.Init(*flagDir)
	defer database.Close()

	AssertTrue(*flagDB == "rocksdb" || *flagDB == "badger")
	switch *flagBench {
	case "writerandom":
		WriteRandom(database)
	case "batchwriterandom":
		BatchWriteRandom(database)
	case "readrandom":
		ReadRandom(database)
	default:
		log.Fatalf("Unknown benchmark: %v", *flagBench)
	}
	wg.Wait()
}
