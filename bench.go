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

	"github.com/dgraph-io/badger/db"

	"github.com/dgraph-io/dgraph/store"
	"github.com/dgraph-io/dgraph/x"
)

var (
	flagBench      = flag.String("bench", "", "Run which benchmark?")
	flagDB         = flag.String("db", "", "Which DB: rocksdb, badger")
	flagValueSize  = flag.Int("value_size", 100, "Size of each value.")
	flagNum        = flag.Int("num", 1000000, "Number of key-value pairs to write.")
	flagRandSize   = flag.Int("rand_size", 1000000, "Size of rng buffer.")
	flagCpuProfile = flag.String("cpu_profile", "", "Write cpu profile to file.")

	rdbStore *store.Store
	rng      randomGenerator
)

type randomGenerator struct {
	data []byte
	idx  int
}

func (s *randomGenerator) Init() {
	s.data = make([]byte, *flagRandSize)
	n, err := rand.Read(s.data)
	x.Check(err)
	x.AssertTrue(n == *flagRandSize)
	s.idx = 0
}

func (s *randomGenerator) Bytes(size int) []byte {
	if s.idx+size > len(s.data) {
		s.idx = 0
	}
	return s.data[s.idx : s.idx+size]
}

func (s *randomGenerator) Int() int {
	b := s.Bytes(4)
	return int(binary.LittleEndian.Uint32(b))
}

func main() {
	x.Init()
	x.AssertTrue(len(*flagBench) > 0)

	rng.Init()

	if *flagCpuProfile != "" {
		f, err := os.Create(*flagCpuProfile)
		if err != nil {
			x.Fatalf("Profiler error: %v", err)
		}
		pprof.StartCPUProfile(f)
	}

	x.AssertTrue(*flagDB == "rocksdb" || *flagDB == "badger")
	switch *flagBench {
	case "writerandom":
		if *flagDB == "badger" {
			report("BadgerWriteRandom", BadgerWriteRandom())
		} else {
			report("RocksDBWriteRandom", RocksDBWriteRandom())
		}
	default:
		x.Fatalf("Unknown benchmark: %v", *flagBench)
	}

	if *flagCpuProfile != "" {
		pprof.StopCPUProfile()
	}
}

func report(label string, d time.Duration) {
	secs := d.Seconds()
	throughput := float64(*flagNum*(16+*flagValueSize)) / ((1 << 20) * secs)
	fmt.Printf("%s: %.2fs, %.2fMb/s\n", label, secs, throughput)
}

// No batching.
func RocksDBWriteRandom() time.Duration {
	dir, err := ioutil.TempDir("", "storetest_")
	x.Check(err)
	defer os.RemoveAll(dir)

	rdbStore, err = store.NewStore(dir)
	x.Check(err)
	defer rdbStore.Close()

	timeStart := time.Now()

	// If you use b.N, you might add too few samples and be working only in memory.
	// We need to fix a large number of pairs. This is what LevelDB benchmark does as well.
	for i := 0; i < *flagNum; i++ {
		key := []byte(fmt.Sprintf("%016d", rng.Int()%*flagNum))
		val := rng.Bytes(*flagValueSize)
		rdbStore.SetOne(key, val)
	}

	return time.Since(timeStart)
}

// No batching.
func BadgerWriteRandom() time.Duration {
	opt := db.DBOptions{
		WriteBufferSize: 1 << 20, // Size of each memtable.
		CompactOpt: db.CompactOptions{
			NumLevelZeroTables:      5,
			NumLevelZeroTablesStall: 10,
			LevelOneSize:            10 << 20,
			MaxLevels:               4,
			NumCompactWorkers:       3,
			MaxTableSize:            2 << 20,
			//			Verbose:                 true,
			Verbose: false,
		},
	}
	ps := db.NewDB(opt)

	timeStart := time.Now()

	// If you use b.N, you might add too few samples and be working only in memory.
	// We need to fix a large number of pairs. This is what LevelDB benchmark does as well.
	for i := 0; i < *flagNum; i++ {
		key := []byte(fmt.Sprintf("%016d", rng.Int()%*flagNum))
		val := rng.Bytes(*flagValueSize)
		ps.Put(key, val)
	}

	return time.Since(timeStart)
}
