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

	"github.com/dgraph-io/badger/y"
)

var (
	dir           = flag.String("dir", "datafiles", "File to read from")
	numSerial     = flag.Int64("sreads", 0, "Number of serial random reads")
	numParallel   = flag.Int64("preads", 2000000, "Number of parallel random reads")
	numGoroutines = flag.Int("numParallel", 8, "Number of go-routines")
)

var readSize int64 = 4 << 10

func getIndices(flist []*os.File, maxFileSize int64) (*os.File, int64) {
	fidx := rand.Intn(len(flist))
	iidx := rand.Int63n(maxFileSize - readSize)
	return flist[fidx], iidx
}

func Serial(fList []*os.File, maxFileSize int64) {
	startT := time.Now()
	var i int64 = 0
	b := make([]byte, int(readSize))

	rand.Seed(int64(time.Now().Second()))
	for ; i < *numSerial; i++ {
		fd, offset := getIndices(fList, maxFileSize)
		_, err := fd.ReadAt(b, offset)
		if err != nil {
			log.Fatalf("Error reading file: %v", err)
		}
		if i%10000 == 0 {
			log.Printf("Finished %v reads in serial", i)
		}
	}
	fmt.Println("Serial: Number of random reads per second: ", float64(*numSerial)/time.Since(startT).Seconds())
	fmt.Println("Serial: Time Taken: ", time.Since(startT))
}

func Conc2(fList []*os.File, maxFileSize int64) {
	startT := time.Now()
	var i int64
	var wg sync.WaitGroup

	rand.Seed(int64(time.Now().Second()))
	for k := 0; k < *numGoroutines; k++ {
		wg.Add(1)
		go func() {
			b := make([]byte, int(readSize))
			for atomic.LoadInt64(&i) < *numParallel {
				fd, offset := getIndices(fList, maxFileSize)
				_, err := fd.ReadAt(b, offset)
				if err != nil {
					log.Fatalf("Error reading file: %v", err)
				}
				atomic.AddInt64(&i, 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Println("Concurrent 2: Number of random reads per second: ", float64(*numParallel)/time.Since(startT).Seconds())
	fmt.Println("Concurrent 2: Time Taken: ", time.Since(startT))
}

type req struct {
	fd     *os.File
	offset int64
}

func Conc3(fList []*os.File, maxFileSize int64) {
	ch := make(chan req, 10000)
	go func() {
		var i int64
		for i = 0; i < *numParallel; i++ {
			var r req
			r.fd, r.offset = getIndices(fList, maxFileSize)
			ch <- r
		}
		close(ch)
	}()

	startT := time.Now()
	var wg sync.WaitGroup
	for k := 0; k < *numGoroutines; k++ {
		wg.Add(1)
		go func() {
			b := make([]byte, int(readSize))
			for req := range ch {
				_, err := req.fd.ReadAt(b, req.offset)
				if err != nil {
					log.Fatalf("Error reading file: %v", err)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Println("Concurrent 3: Number of random reads per second: ", float64(*numParallel)/time.Since(startT).Seconds())
	fmt.Println("Concurrent 3: Time Taken: ", time.Since(startT))
}

func main() {
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

	Serial(flist, maxFileSize)
	Conc2(flist, maxFileSize)
	Conc3(flist, maxFileSize)
}
