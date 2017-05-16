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
	dir         = flag.String("dir", "datafiles", "File to read from")
	numSRead    = flag.Int64("numSread", 100000, "Number of random reads")
	numPRead    = flag.Int64("numPread", 2160000, "Number of random reads")
	numParallel = flag.Int("numParallel", 8, "Number of go-routines")
)

var maxAddr int64 = 32 << 30
var fileSize int64 = 4 << 30
var readSize int64 = 4 << 10

func Serial(fList []*os.File) {
	startT := time.Now()
	var i int64 = 0
	b := make([]byte, int(readSize))

	rand.Seed(int64(time.Now().Second()))
	for ; i < *numSRead; i++ {
		idx := rand.Int63n(maxAddr)
		fidx := idx / fileSize
		iidx := idx % fileSize
		if iidx >= fileSize-readSize {
			iidx -= readSize
		}
		_, err := fList[fidx].ReadAt(b, iidx)
		if err != nil {
			log.Fatalf("Error reading file: %v", err)
		}
		if i%10000 == 0 {
			log.Printf("Finished %v reads in serial", i)
		}
	}
	fmt.Println("Serial: Number of random reads per second: ", float64(*numSRead)/time.Since(startT).Seconds())
	fmt.Println("Serial: Time Taken: ", time.Since(startT))
}

func Conc1(fList []*os.File) {
	startT := time.Now()
	var i int64
	var j int64
	b := make([]byte, int(readSize))
	var wg sync.WaitGroup

	rand.Seed(int64(time.Now().Second()))
	for i < *numPRead {
		if atomic.LoadInt64(&j) < int64(*numParallel) {
			wg.Add(1)
			atomic.AddInt64(&j, 1)
			go func() {
				idx := rand.Int63n(maxAddr)
				fidx := idx / fileSize
				iidx := idx % fileSize
				if iidx >= fileSize-readSize {
					iidx -= readSize
				}
				// We use shared 'b' but it should be okay
				_, err := fList[fidx].ReadAt(b, iidx)
				if err != nil {
					log.Fatalf("Error reading file: %v", err)
				}
				atomic.AddInt64(&i, 1)
				atomic.AddInt64(&j, -1)
				wg.Done()
			}()
		}
	}
	wg.Wait()
	fmt.Println("Concurrent 1: Number of random reads per second: ", float64(*numPRead)/time.Since(startT).Seconds())
	fmt.Println("Concurrent 1: Time Taken: ", time.Since(startT))
}

func Conc2(fList []*os.File) {
	startT := time.Now()
	var i int64
	var wg sync.WaitGroup

	rand.Seed(int64(time.Now().Second()))
	for k := 0; k < *numParallel; k++ {
		wg.Add(1)
		go func() {
			b := make([]byte, int(readSize))
			for atomic.LoadInt64(&i) < *numPRead {
				idx := rand.Int63n(maxAddr)
				fidx := idx / fileSize
				iidx := idx % fileSize
				if iidx >= fileSize-readSize {
					iidx -= readSize
				}
				_, err := fList[fidx].ReadAt(b, iidx)
				if err != nil {
					log.Fatalf("Error reading file: %v", err)
				}
				atomic.AddInt64(&i, 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Println("Concurrent 2: Number of random reads per second: ", float64(*numPRead)/time.Since(startT).Seconds())
	fmt.Println("Concurrent 2: Time Taken: ", time.Since(startT))
}

func Conc3(fList []*os.File, ch chan int64) {
	startT := time.Now()
	var i int64
	var wg sync.WaitGroup

	for k := 0; k < *numParallel; k++ {
		wg.Add(1)
		go func() {
			b := make([]byte, int(readSize))
			for idx := range ch {
				fidx := idx / fileSize
				iidx := idx % fileSize
				if iidx >= fileSize-readSize {
					iidx -= readSize
				}
				_, err := fList[fidx].ReadAt(b, iidx)
				if err != nil {
					log.Fatalf("Error reading file: %v", err)
				}
				atomic.AddInt64(&i, 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Println("Concurrent 3: Number of random reads per second: ", float64(*numPRead)/time.Since(startT).Seconds())
	fmt.Println("Concurrent 3: Time Taken: ", time.Since(startT))
}

func main() {
	var fList []*os.File
	printFile := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Print(err)
			return nil
		}

		if !info.IsDir() {
			f, err := os.Open(path)
			y.AssertTruef(err == nil, "Error opening file: %v", path)
			fList = append(fList, f)
			log.Println("Opened file:", path, "Size:", info.Size()/(1<<20), "MB")
		}
		return nil
	}

	err := filepath.Walk(*dir, printFile)
	if err != nil {
		log.Fatalf("%v", err)
	}
	//	Serial(fList)
	//Conc1(fList)
	Conc2(fList)

	rand.Seed(int64(time.Now().Second()))
	ch := make(chan int64, 10000)
	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		Conc3(fList, ch)
		wg.Done()
	}()
	var i int64 = 0
	for ; i < *numPRead; i++ {
		idx := rand.Int63n(maxAddr)
		ch <- idx
	}
	close(ch)
	wg.Wait()
}
