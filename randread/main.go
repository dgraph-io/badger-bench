package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/codahale/hdrhistogram"
	"github.com/dgraph-io/badger/y"
	"github.com/pkg/profile"
	"github.com/traetox/goaio"
)

var (
	dir           = flag.String("dir", "datafiles", "File to read from")
	mode          = flag.Int("mode", 1, "0 = serial, 1 = parallel, 2 = parallel via channel")
	duration      = flag.Int64("seconds", 30, "Number of seconds to run for")
	numGoroutines = flag.Int("jobs", 8, "Number of Goroutines")
	profilemode   = flag.String("profile.mode", "", "Enable profiling mode, one of [cpu, mem, mutex, block, trace]")
)
var done int32
var readSize int64 = 1 << 10

func getIndices(r *rand.Rand, flist []*os.File, maxFileSize int64) (*os.File, int64) {
	fidx := r.Intn(len(flist))
	iidx := r.Int63n(maxFileSize - readSize)
	return flist[fidx], iidx
}

func getAIOIndices(r *rand.Rand, aios []*goaio.AIO, maxFileSize int64) (*goaio.AIO, int64) {
	fidx := r.Intn(len(aios))
	iidx := r.Int63n(maxFileSize - readSize)
	return aios[fidx], iidx
}

func Serial(fList []*os.File, maxFileSize int64) {
	startT := time.Now()
	var count int64 = 0
	b := make([]byte, int(readSize))
	hist := hdrhistogram.New(1, 1000000, 4) // in microseconds.

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		if atomic.LoadInt32(&done) == 1 {
			break
		}
		count++
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
		if count%10000 == 0 {
			fmt.Printf("Serial: Number of random reads per second: %f\r",
				float64(hist.TotalCount())/time.Since(startT).Seconds())
		}
	}
	fmt.Println("Serial: Number of random reads per second: ",
		float64(count)/time.Since(startT).Seconds())
	fmt.Println("Serial: Time Taken: ", time.Since(startT))
}

func Conc2(fList []*os.File, maxFileSize int64) {
	startT := time.Now()
	var wg sync.WaitGroup
	hist := hdrhistogram.New(1, 1000000, 4) // in microseconds.

	for k := 0; k < *numGoroutines; k++ {
		wg.Add(1)
		go func() {
			b := make([]byte, int(readSize))
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for {
				if atomic.LoadInt32(&done) == 1 {
					break
				}
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
	go func() {
		d := time.NewTicker(time.Second)
		for range d.C {
			fmt.Printf("Concurrent 2: Number of random reads per second: %f\r",
				float64(hist.TotalCount())/time.Since(startT).Seconds())
			if atomic.LoadInt32(&done) == 1 {
				fmt.Println()
				fmt.Println("Concurrent 2: Time Taken: ", time.Since(startT))
				fmt.Println("Total count by hist", hist.TotalCount())
				break
			}
		}
	}()

	wg.Wait()
	for _, b := range hist.CumulativeDistribution() {
		fmt.Printf("[%f] %d\n", b.Quantile, b.ValueAt)
	}
	fmt.Printf("=> [0.9] %d\n", hist.ValueAtQuantile(90.0))
}

type aioReq struct {
	aio    *goaio.AIO
	offset int64
	start  time.Time
}

func ConcAio(fList []*os.File, maxFileSize int64) {
	startT := time.Now()
	var wg sync.WaitGroup
	hist := hdrhistogram.New(1, 1000000, 4) // in microseconds.

	aioCfg := goaio.AIOExtConfig{
		QueueDepth: 32,
	}
	aioList := make([]*goaio.AIO, len(fList))
	for i := range fList {
		aio, err := goaio.NewAIOExt(fList[i].Name(), aioCfg, os.O_RDONLY, 0755)
		if err != nil {
			fmt.Println("Failed to create new AIO for", fList[i].Name(), err)
			return
		}
		aioList[i] = aio
	}

	for k := 0; k < *numGoroutines; k++ {
		wg.Add(1)
		go func() {
			b := make([]byte, int(readSize))
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			ids := make([]goaio.RequestId, 32)
			for {
				if atomic.LoadInt32(&done) == 1 {
					break
				}
				aio, offset := getAIOIndices(r, aioList, maxFileSize)
				// if !aio.Ready() {
				// 	if _, err := aio.WaitAny(ids); err != nil {
				// 		log.Fatalf("Failed to collect requests", err)
				// 	}
				// }
				// r.start = time.Now()
				if _, err := aio.ReadAt(b, offset); err != nil {
					log.Fatalf("Unable to read: %v", err)
				}
			}

			for _, aio := range aioList {
				for {
					n, err := aio.WaitAny(ids)
					if err != nil {
						log.Fatalf("Failed to collect reqs: %v", err)
					}
					if n == 0 {
						break
					}
				}
			}
			// dur := time.Since(start).Nanoseconds() / 1000
			// if dur > 1000000 {
			// 	dur = 1000000
			// }
			// if err = hist.RecordValue(dur); err != nil {
			// 	log.Fatalf("Unable to record hist: %v", err)
			// }
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Println("Concurrent 2: Number of random reads per second: ",
		float64(hist.TotalCount())/time.Since(startT).Seconds())
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

	done = 0
	go func() {
		time.Sleep(time.Duration(*duration) * time.Second)
		atomic.StoreInt32(&done, 1)
	}()

	switch *mode {
	case 0:
		Serial(flist, maxFileSize)
	case 1:
		Conc2(flist, maxFileSize)
	case 2:
		ConcAio(flist, maxFileSize)
	default:
		log.Fatalf("Unknown mode")
	}
}
