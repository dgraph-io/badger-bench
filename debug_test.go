package main

import (
	"fmt"
	"testing"

	"github.com/bmatsuo/lmdb-go/lmdb"
	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/y"
)

// Print all keys in lmdb store
func TestPrintLmdbKeys(t *testing.T) {
	lmdbEnv := getLmdb()

	var lmdbDBI lmdb.DBI
	// Acquire handle
	err := lmdbEnv.View(func(txn *lmdb.Txn) error {
		var err error
		lmdbDBI, err = txn.OpenDBI("bench", 0)
		return err
	})
	y.Check(err)
	defer lmdbEnv.CloseDBI(lmdbDBI)

	err = lmdbEnv.View(func(txn *lmdb.Txn) error {
		txn.RawRead = true

		cur, err := txn.OpenCursor(lmdbDBI)
		if err != nil {
			return err
		}
		defer cur.Close()

		for {
			k1, _, err := cur.Get(nil, nil, lmdb.Next)
			if lmdb.IsNotFound(err) {
				return nil
			}
			if err != nil {
				return err
			}

			fmt.Printf("%s\n", string(k1))
		}
	})
	y.Check(err)
}

// Print all keys in badger store
func TestPrintBadgerKeys(t *testing.T) {
	bdb, err := getBadger()
	y.Check(err)

	opt := badger.IteratorOptions{}
	opt.PrefetchSize = 10000
	opt.FetchValues = true
	itr := bdb.NewIterator(opt)
	for itr.Rewind(); itr.Valid(); itr.Next() {
		item := itr.Item()
		fmt.Printf("%s\n", string(item.Key()))
	}
}
