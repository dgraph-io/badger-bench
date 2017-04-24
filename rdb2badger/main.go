package main

// This script copies data from RocksDB files to Badger.

import (
	"context"
	"flag"
	"github.com/dgraph-io/badger/badger"
	"github.com/dgraph-io/badger/table"
	"github.com/dgraph-io/badger/y"
	"github.com/dgraph-io/dgraph/store"
	"gopkg.in/cheggaaa/pb.v1"
	"os"
)

// Rdb2badger copies data from RocksDB store to Badger.
func Rdb2badger(ctx context.Context, rdb *store.Store, bdb *badger.KV, records int) {
	it := rdb.NewIterator()
	defer it.Close()

	it.Seek([]byte(""))

	bar := pb.StartNew(records)

	for retrieved := 0; it.Valid() && retrieved < records; it.Next() {
		err := bdb.Put(ctx, it.Key().Data(), it.Value().Data())
		y.Check(err)
		y.Check(it.Err())

		bar.Increment()
		retrieved++
	}
}

var (
	input  = flag.String("input", "tmp/rocksdb", "Path to RocksDB data.")
	output = flag.String("output", "tmp/badger", "Path to Badger data.")
	limit  = flag.Int("limit", 100, "Limit of number of records to retrieve.")
)

func main() {
	flag.Parse()

	var ctx = context.Background()

	rdb, err := store.NewReadOnlyStore(*input)
	y.Check(err)
	defer rdb.Close()

	opt := badger.DefaultOptions
	opt.NumMemtables = 3
	opt.MapTablesTo = table.Nothing
	opt.Verbose = true
	opt.Dir = *output
	opt.SyncWrites = false

	y.Check(os.RemoveAll(*output))
	err = os.MkdirAll(*output, 0777)
	y.Check(err)
	bdb := badger.NewKV(&opt)
	defer bdb.Close()

	Rdb2badger(ctx, rdb, bdb, *limit)
}
