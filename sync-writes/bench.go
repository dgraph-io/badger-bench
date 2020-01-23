package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dgraph-io/badger/v2"
)

var (
	num        = flag.Int("num", 1000000, "Number of iterations")
	fileTest   = flag.Bool("file", false, "Write to file")
	badgerTest = flag.Bool("badger", false, "Write to Badger")
)

func removeDir(dir string) {
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}
}

func writeFile() {
	dir := "."
	err := os.Remove(fmt.Sprintf("%s/test.file", dir))
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile(fmt.Sprintf("%s/test.file", dir),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	kv := []byte("hello:world\n")

	for i := 0; i < *num; i++ {
		n, err := f.Write(kv)
		if n != len(kv) {
			log.Fatal(err)
		}
		if err != nil {
			log.Fatal(err)
		}

		start := time.Now()
		err = f.Sync()
		duration := time.Since(start)
		fmt.Printf("%v\n", duration)
	}
}

func writeBadger() {
	dir := "badgerdb"
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}
	opt := badger.DefaultOptions(dir).
		WithSyncWrites(false)
	db, err := badger.Open(opt)
	if err != nil {
		log.Fatal(err)
	}

	updateFunc := func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry([]byte("hello"), []byte("world")))
	}

	for i := 0; i < *num; i++ {
		err := db.Update(updateFunc)
		if err != nil {
			log.Fatal(err)
		}

		start := time.Now()
		err = db.Sync()
		duration := time.Since(start)
		fmt.Printf("%v\n", duration)
	}
}

func main() {
	flag.Parse()

	if *fileTest {
		writeFile()
	}
	if *badgerTest {
		writeBadger()
	}
}
