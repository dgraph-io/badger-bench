package db_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/dgraph-io/badger/v2"
	"github.com/stretchr/testify/require"
)

func removeDir(dir string) {
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}
}

func BenchmarkFileNoSync(b *testing.B) {
	dir, err := ioutil.TempDir("", "badger-test")
	require.NoError(b, err)
	defer removeDir(dir)

	f, err := os.OpenFile(fmt.Sprintf("%s/test.file", dir),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	require.NoError(b, err)

	kv := []byte("hello:world\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n, err := f.Write(kv)
		require.NoError(b, err)
		require.Equal(b, n, 12)
		// err = f.Sync()
		require.NoError(b, err)
	}
}

func BenchmarkFileSync(b *testing.B) {
	dir, err := ioutil.TempDir("", "badger-test")
	require.NoError(b, err)
	defer removeDir(dir)

	f, err := os.OpenFile(fmt.Sprintf("%s/test.file", dir),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	require.NoError(b, err)

	kv := []byte("hello:world\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n, err := f.Write(kv)
		require.NoError(b, err)
		require.Equal(b, n, 12)
		err = f.Sync()
		require.NoError(b, err)
	}
}

func BenchmarkBadgerSyncWrites(b *testing.B) {
	dir, err := ioutil.TempDir("", "badger-test")
	require.NoError(b, err)
	defer removeDir(dir)

	opt := badger.DefaultOptions(dir).
		WithSyncWrites(true).
		WithLogger(nil)
	db, err := badger.Open(opt)
	require.NoError(b, err)

	updateFunc := func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry([]byte("hello"), []byte("world")))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := db.Update(updateFunc)
		require.NoError(b, err)
	}
}

func BenchmarkBadgerNoSyncWrites(b *testing.B) {
	dir, err := ioutil.TempDir("", "badger-test")
	require.NoError(b, err)
	defer removeDir(dir)

	opt := badger.DefaultOptions(dir).
		WithSyncWrites(false).
		WithLogger(nil)
	db, err := badger.Open(opt)
	require.NoError(b, err)

	updateFunc := func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry([]byte("hello"), []byte("world")))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := db.Update(updateFunc)
		require.NoError(b, err)
	}
}

func BenchmarkBadgerSyncWritesManually(b *testing.B) {
	dir, err := ioutil.TempDir("", "badger-test")
	require.NoError(b, err)
	defer removeDir(dir)

	opt := badger.DefaultOptions(dir).
		WithSyncWrites(false).
		WithLogger(nil)
	db, err := badger.Open(opt)
	require.NoError(b, err)

	updateFunc := func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry([]byte("hello"), []byte("world")))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := db.Update(updateFunc)
		require.NoError(b, err)
		err = db.Sync()
		require.NoError(b, err)
	}
}
