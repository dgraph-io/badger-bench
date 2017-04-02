package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"

	"github.com/dgraph-io/badger/badger"

	"github.com/dgraph-io/dgraph/store"
	"github.com/dgraph-io/dgraph/x"
)

var (
	flagBench      = flag.String("bench", "", "Run which benchmark?")
	flagDB         = flag.String("db", "", "Which DB: rocksdb, badger")
	flagValueSize  = flag.Int("value_size", 100, "Size of each value.")
	flagBatchSize  = flag.Int("batch_size", 1, "Size of writebatch.")
	flagNum        = flag.Int("num", 1000000, "Number of key-value pairs to write.")
	flagRandSize   = flag.Int("rand_size", 1000000, "Size of rng buffer.")
	flagCpuProfile = flag.String("cpu_profile", "", "Write cpu profile to file.")
	flagVerbose    = flag.Bool("verbose", false, "Verbose.")

	rdbStore *store.Store
	rng      randomGenerator
)

type randomGenerator struct {
	data []byte
	idx  int
}

func (s *randomGenerator) Init() {
	if *flagRandSize <= 0 {
		// Will not precompute the randomness.
		return
	}
	s.data = make([]byte, *flagRandSize)
	n, err := rand.Read(s.data)
	x.Check(err)
	x.AssertTrue(n == *flagRandSize)
	s.idx = 0
}

// Bytes generates len(out) random bytes and writes to out.
func (s *randomGenerator) Bytes(out []byte) {
	if *flagRandSize == 0 {
		n, err := rand.Read(out)
		x.AssertTrue(n == len(out))
		x.Check(err)
		return
	}
	size := len(out)
	if s.idx+size > len(s.data) {
		s.idx = 0
	}
	x.AssertTrue(size == copy(out, s.data[s.idx:s.idx+size]))
	s.idx += size
}

func (s *randomGenerator) Int() int {
	var buf [4]byte
	s.Bytes(buf[:])
	return int(binary.LittleEndian.Uint32(buf[:]))
}

func main() {
	x.Init()
	x.AssertTrue(len(*flagBench) > 0)
	x.AssertTrue(*flagValueSize > 0)
	rng.Init()

	if *flagCpuProfile != "" {
		f, err := os.Create(*flagCpuProfile)
		if err != nil {
			x.Fatalf("Profiler error: %v", err)
		}
		pprof.StartCPUProfile(f)
	}

	var database Database
	switch *flagDB {
	case "badger":
		database = new(BadgerAdapter)
	case "rocksdb":
		database = new(RocksDBAdapter)
	default:
		x.Fatalf("Database invalid: %v", *flagDB)
	}
	database.Init()
	defer database.Close()

	x.AssertTrue(*flagDB == "rocksdb" || *flagDB == "badger")
	label := fmt.Sprintf("%s_%s", *flagBench, *flagDB)
	switch *flagBench {
	case "writerandom":
		WriteRandom(database)
	case "batchwriterandom":
		BatchWriteRandom(database)
	default:
		x.Fatalf("Unknown benchmark: %v", *flagBench)
	}
	if *flagCpuProfile != "" {
		pprof.StopCPUProfile()
	}
}

func report(d time.Duration, n int) string {
	secs := d.Seconds()
	throughput := float64(n*(16+*flagValueSize)) / ((1 << 20) * secs)
	return fmt.Sprintf("%.2fs, %.2fMb/s", secs, throughput)
}

type Database interface {
	Init()
	Close()
	Put(key, val []byte)
	BatchPut(key, val [][]byte)
}

type RocksDBAdapter struct {
	rdb *store.Store
	dir string
}

func (s *RocksDBAdapter) Init() {
	var err error
	s.dir, err = ioutil.TempDir("/tmp", "storetest_")
	x.Check(err)
	s.rdb, err = store.NewStore(s.dir)
	x.Check(err)
}

func (s *RocksDBAdapter) Close() {
	os.RemoveAll(s.dir)
	s.rdb.Close()
}

func (s *RocksDBAdapter) Put(key, val []byte) {
	s.rdb.SetOne(key, val)
}

func (s *RocksDBAdapter) BatchPut(key, val [][]byte) {
	wb := s.rdb.NewWriteBatch()
	x.AssertTrue(len(key) == len(val))
	for i := 0; i < len(key); i++ {
		wb.Put(key[i], val[i])
	}
	x.Check(s.rdb.WriteBatch(wb))
}

type BadgerAdapter struct {
	db  *badger.DB
	dir string
}

func (s *BadgerAdapter) Init() {
	opt := badger.DBOptions{
		WriteBufferSize:         1 << 20, // Size of each memtable.
		NumLevelZeroTables:      5,
		NumLevelZeroTablesStall: 6,
		LevelOneSize:            5 << 20,
		MaxLevels:               7,
		NumCompactWorkers:       6,
		MaxTableSize:            2 << 20,
		LevelSizeMultiplier:     5,
		Verbose:                 *flagVerbose,
		Dir:                     "/tmp/badger_bench",
	}
	s.db = badger.NewDB(opt)
}

func (s *BadgerAdapter) Close() {
}

func (s *BadgerAdapter) Put(key, val []byte) {
	s.db.Put(key, val)
}

func (s *BadgerAdapter) BatchPut(key, val [][]byte) {
	wb := badger.NewWriteBatch(len(key))
	x.AssertTrue(len(key) == len(val))
	for i := 0; i < len(key); i++ {
		wb.Put(key[i], val[i])
	}
	x.Check(s.db.Write(wb))
}

// No batching.
func WriteRandom(database Database) {
	database.Init()
	defer database.Close()
	timeStart := time.Now()
	timeLog := timeStart
	timeLogI := 0
	// If you use b.N, you might add too few samples and be working only in memory.
	// We need to fix a large number of pairs. This is what LevelDB benchmark does as well.
	for i := 0; i < *flagNum; i++ {
		key := []byte(fmt.Sprintf("%016d", rng.Int()%*flagNum))
		val := make([]byte, *flagValueSize)
		rng.Bytes(val)
		database.Put(key, val)
		timeElapsed := time.Since(timeLog)
		if timeElapsed > 5*time.Second {
			x.Printf("%.2f percent: %s\n", float64(i)*100.0/float64(*flagNum),
				report(timeElapsed, i-timeLogI))
			timeLog = time.Now()
			timeLogI = i
		}
	}
	x.Printf("Overall: %s\n", report(time.Since(timeStart), *flagNum))
}

// With batching.
func BatchWriteRandom(database Database) {
	database.Init()
	defer database.Close()
	timeStart := time.Now()
	timeLog := timeStart
	timeLogI := 0
	keys := make([][]byte, *flagBatchSize)
	vals := make([][]byte, *flagBatchSize)

	for i := 0; i < *flagNum; i++ {
		for j := 0; j < *flagBatchSize; j++ {
			keys[j] = []byte(fmt.Sprintf("%016d", rng.Int()%*flagNum))
			vals[j] = make([]byte, *flagValueSize)
			rng.Bytes(vals[j])
		}
		database.BatchPut(keys, vals)
		timeElapsed := time.Since(timeLog)
		if timeElapsed > 5*time.Second {
			x.Printf("%.2f percent: %s\n", float64(i)*100.0/float64(*flagNum),
				report(timeElapsed, (i-timeLogI)*(*flagBatchSize)))
			timeLog = time.Now()
			timeLogI = i
		}
	}
	x.Printf("Overall: %s\n", report(time.Since(timeStart), *flagNum*(*flagBatchSize)))
}
