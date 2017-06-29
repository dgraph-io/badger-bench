package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/codahale/hdrhistogram"
	"github.com/dgraph-io/badger/y"
	"github.com/pkg/profile"
)

var (
	dir           = flag.String("dir", "datafiles", "File to read from")
	numReads      = flag.Int64("num", 2000000, "Number of reads")
	mode          = flag.Int("mode", 1, "0 = serial, 1 = parallel, 2 = parallel via channel")
	numGoroutines = flag.Int("jobs", 8, "Number of Goroutines")
	profilemode   = flag.String("profile.mode", "", "Enable profiling mode, one of [cpu, mem, mutex, block, trace]")
)

var readSize int64 = 4 << 10

func getIndices(r *rand.Rand, flist []*os.File, maxFileSize int64) (*os.File, int64) {
	fidx := r.Intn(len(flist))
	iidx := r.Int63n(maxFileSize - readSize)
	return flist[fidx], iidx
}

func Serial(fList []*os.File, maxFileSize int64) {
	startT := time.Now()
	var i int64 = 0
	b := make([]byte, int(readSize))

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for ; i < *numReads; i++ {
		fd, offset := getIndices(r, fList, maxFileSize)
		_, err := fd.ReadAt(b, offset)
		if err != nil {
			log.Fatalf("Error reading file: %v", err)
		}
		if i%10000 == 0 {
			log.Printf("Finished %v reads in serial", i)
		}
	}
	fmt.Println("Serial: Number of random reads per second: ",
		float64(*numReads)/time.Since(startT).Seconds())
	fmt.Println("Serial: Time Taken: ", time.Since(startT))
}

func Conc2(fList []*os.File, maxFileSize int64) {
	startT := time.Now()
	var wg sync.WaitGroup
	countPerGo := *numReads / int64(*numGoroutines)
	fmt.Printf("Concurrent mode: Reads per goroutine: %d\n", countPerGo)
	hist := hdrhistogram.New(1, 1000000, 4) // in microseconds.

	for k := 0; k < *numGoroutines; k++ {
		wg.Add(1)
		go func() {
			b := make([]byte, int(readSize))
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			var i int64
			for ; i < countPerGo; i++ {
				fd, offset := getIndices(r, fList, maxFileSize)
				start := time.Now()
				_, err := fd.ReadAt(b, offset)
				if err != nil {
					log.Fatalf("Error reading file: %v", err)
				}
				dur := time.Since(start).Nanoseconds() / 1000
				if dur > 1000000 {
					dur = 1000000
				}
				if err = hist.RecordValue(dur); err != nil {
					log.Fatalf("Unable to record hist: %v", err)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Println("Concurrent 2: Number of random reads per second: ",
		float64(*numReads)/time.Since(startT).Seconds())
	fmt.Println("Concurrent 2: Time Taken: ", time.Since(startT))
	total := float64(hist.TotalCount())
	fmt.Println("Total count by hist", total)
	for _, b := range hist.CumulativeDistribution() {
		fmt.Printf("[%f] %d\n", b.Quantile, b.ValueAt)
	}
	fmt.Printf("=> [0.9] %d\n", hist.ValueAtQuantile(90.0))
}

func main() {
	flag.Parse()

	var flist []*os.File
	var maxFileSize int64
	getFile := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Print(err)
			return nil
		}

		if !info.IsDir() {
			f, err := os.Open(path)
			y.AssertTruef(err == nil, "Error opening file: %v", path)
			flist = append(flist, f)
			log.Println("Opened file:", path, "Size:", info.Size()/(1<<20), "MB")
			maxFileSize = info.Size()
		}
		return nil
	}

	err := filepath.Walk(*dir, getFile)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if len(flist) == 0 {
		log.Fatalf("Must have files already created")
	}

	switch *profilemode {
	case "cpu":
		defer profile.Start(profile.CPUProfile).Stop()
	case "mem":
		defer profile.Start(profile.MemProfile).Stop()
	case "mutex":
		defer profile.Start(profile.MutexProfile).Stop()
	case "block":
		defer profile.Start(profile.BlockProfile).Stop()
	case "trace":
		defer profile.Start(profile.TraceProfile).Stop()
	}

	switch *mode {
	case 0:
		Serial(flist, maxFileSize)
	case 1:
		Conc2(flist, maxFileSize)
	default:
		log.Fatalf("Unknown mode")
	}
}
