package main

// This script copies data from RocksDB files to Badger.

import (
	"context"
	"flag"
	"fmt"
	"github.com/codahale/hdrhistogram"
	"github.com/dgraph-io/badger/badger"
	"github.com/dgraph-io/badger/table"
	"github.com/dgraph-io/badger/y"
	"github.com/dgraph-io/dgraph/store"
	"gopkg.in/cheggaaa/pb.v1"
	"os"
)

func makeCopy(a []byte) []byte {
	b := make([]byte, len(a))
	copy(b, a)
	return b
}

func printStats(histogram *hdrhistogram.Histogram, buckets int) {
	max := histogram.Max()

	fmt.Printf("Min %d\n", histogram.Min())
	fmt.Printf("Max %d\n", max)
	fmt.Printf("Mean %f\n", histogram.Mean())

	distribution := histogram.Distribution()
	fmt.Println("From")
	for j := 0; j < len(distribution); j++ {
		fmt.Println(distribution[j].From)
	}

	fmt.Println("To")
	for j := 0; j < len(distribution); j++ {
		fmt.Println(distribution[j].To)
	}

	fmt.Println("Count")
	for j := 0; j < len(distribution); j++ {
		fmt.Println(distribution[j].Count)
	}

	fmt.Printf("Bucket counts (bucket size: %d)\n", max/int64(buckets))
	bucketSum := int64(0)
	curBucket := int64(0)

	for j := 0; j < len(distribution); j++ {
		if distribution[j].From > (max/int64(buckets))*curBucket {
			fmt.Println(bucketSum)
			bucketSum = 0
			curBucket++
		}
		bucketSum += distribution[j].Count
	}
}

// Rdb2badger copies data from RocksDB store to Badger.
// Additionaly it computes basic statics of value sizes.
func Rdb2badger(ctx context.Context, rdb *store.Store, bdb *badger.KV, records int) {
	it := rdb.NewIterator()
	defer it.Close()

	it.Seek([]byte(""))

	bar := pb.StartNew(records)

	bufferSize := 1000
	entriesBuffer := make([]*badger.Entry, bufferSize)
	nextFree := 0

	histogram := hdrhistogram.New(0, 10000000, 1)

	for retrieved := 0; it.Valid() && retrieved < records; it.Next() {
		e := new(badger.Entry)
		e.Key = makeCopy(it.Key().Data())
		e.Value = makeCopy(it.Value().Data())

		entriesBuffer[nextFree] = e
		nextFree++

		err := histogram.RecordValue(int64(len(e.Value)))
		y.Check(err)

		if nextFree == bufferSize {
			copied := make([]*badger.Entry, bufferSize)
			copy(copied, entriesBuffer)
			err = bdb.Write(ctx, copied)
			y.Check(err)
			bar.Add(bufferSize)

			nextFree = 0
		}

		y.Check(it.Err())
		retrieved++
	}

	if nextFree > 0 {
		err := bdb.Write(ctx, entriesBuffer[:nextFree])
		y.Check(err)
		bar.Add(bufferSize)
	}

	printStats(histogram, 30)
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
