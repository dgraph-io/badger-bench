// Benchmark batched writing to lmdb using txns against without using them.
//
// This is needed because lmdb does not have support for batched writes,
// and we are trying to simulate it using a sub-txn, going by a hint in
// the Node bindings for lmdb: https://github.com/rvagg/lmdb/blob/master/src/database.cc#L208
package main

import (
	"flag"
	"os"
	"testing"

	"github.com/bmatsuo/lmdb-go/lmdb"
	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/y"
)

func getLmdbEnv() *lmdb.Env {
	os.RemoveAll(*dir + "/lmdb")
	os.MkdirAll(*dir+"/lmdb", 0777)

	var err error
	lmdbEnv, err = lmdb.NewEnv()
	y.Check(err)
	err = lmdbEnv.SetMaxDBs(1)
	y.Check(err)
	err = lmdbEnv.SetMapSize(1 << 36) // ~68Gb
	y.Check(err)

	err = lmdbEnv.Open(*dir+"/lmdb", 0, 0777)
	y.Check(err)
	return lmdbEnv
}

// Create a function that wraps env.Update and sends the resulting error
// over a channel.  Because env.Update is called our update function will
// call runtime.LockOSThread to safely issue the update operation.
var update = func(res chan<- error, op lmdb.TxnOp) {
	res <- lmdbEnv.Update(op)
}

func writeEntries(dbi lmdb.DBI, txn *lmdb.Txn, entries []*badger.Entry) error {
	for _, e := range entries {
		err := txn.Put(dbi, e.Key, e.Value, 0)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeSimpleBatched(entries []*badger.Entry, dbi lmdb.DBI) {
	err := lmdbEnv.Update(func(txn *lmdb.Txn) error {
		return writeEntries(dbi, txn, entries)
	})
	y.Check(err)
}

func writeTxnBatched(entries []*badger.Entry, dbi lmdb.DBI) {
	err := lmdbEnv.Update(func(txn *lmdb.Txn) error {
		return txn.Sub(func(txn *lmdb.Txn) error {
			return writeEntries(dbi, txn, entries)
		})
	})
	y.Check(err)
}

func BenchmarkLmdbBatch(b *testing.B) {
	entries := make([]*badger.Entry, 1000)
	for i := 0; i < len(entries); i++ {
		e := new(badger.Entry)
		e.Key = make([]byte, 22)
		e.Value = make([]byte, *valueSize)
		entries[i] = e
	}

	lmdbEnv := getLmdbEnv()
	defer lmdbEnv.Close()

	var dbi lmdb.DBI
	err := lmdbEnv.Update(func(txn *lmdb.Txn) error {
		var err error
		dbi, err = txn.CreateDBI("bench")
		return err
	})
	y.Check(err)

	b.Run("SimpleBatched", func(b *testing.B) {
		// Do a batched write without txns
		for i := 0; i < b.N; i++ {
			writeSimpleBatched(entries, dbi)
		}
	})

	b.Run("TxnBatched", func(b *testing.B) {
		// Do a batched write with explicit txns
		for i := 0; i < b.N; i++ {
			writeTxnBatched(entries, dbi)
		}
	})
}

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}
