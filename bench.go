package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/y"

	"github.com/dgraph-io/badger"
)

var (
	flagDir = flag.String("dir", "bench-tmp", "Where data is temporarily stored.")
)

func main() {
	fmt.Println("badger-bench")
	flag.Parse()
	opt := badger.DefaultOptions
	opt.Dir = *flagDir + "/badger"
	bdb, err := badger.NewKV(&opt)
	y.Check(err)
	defer bdb.Close()
	time.Sleep(10 * time.Second)
}
