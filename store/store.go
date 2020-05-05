/*
 * Copyright (C) 2017 Dgraph Labs, Inc. and Contributors
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package store

import (
	"strconv"

	rdb "github.com/tecbot/gorocksdb"
)

// Store contains some handles to RocksDB.
type Store struct {
	db       *rdb.DB
	opt      *rdb.Options // Contains blockopt.
	blockopt *rdb.BlockBasedTableOptions
	ropt     *rdb.ReadOptions
	wopt     *rdb.WriteOptions
}

func (s *Store) setOpts() {
	s.opt = rdb.NewDefaultOptions()
	s.blockopt = rdb.NewDefaultBlockBasedTableOptions()
	s.opt.SetBlockBasedTableFactory(s.blockopt)

	// If you want to access blockopt.blockCache, you need to grab handles to them
	// as well. Otherwise, they will be nil. However, for now, we do not really need
	// to do this.
	// s.blockopt.SetBlockCache(rocksdb.NewLRUCache(blockCacheSize))
	// s.blockopt.SetBlockCacheCompressed(rocksdb.NewLRUCache(blockCacheSize))

	s.opt.SetCreateIfMissing(true)
	fp := rdb.NewBloomFilter(16)
	s.blockopt.SetFilterPolicy(fp)

	s.ropt = rdb.NewDefaultReadOptions()
	s.wopt = rdb.NewDefaultWriteOptions()
	s.wopt.SetSync(false) // We don't need to do synchronous writes.
}

// NewStore constructs a Store object at filepath, given some options.
func NewStore(filepath string) (*Store, error) {
	s := &Store{}
	s.setOpts()
	var err error
	s.db, err = rdb.OpenDb(s.opt, filepath)
	return s, err
}

func NewSyncStore(filepath string) (*Store, error) {
	s := &Store{}
	s.setOpts()
	s.wopt.SetSync(true) // Do synchronous writes.
	var err error
	s.db, err = rdb.OpenDb(s.opt, filepath)
	return s, err
}

// NewReadOnlyStore constructs a readonly Store object at filepath, given options.
func NewReadOnlyStore(filepath string) (*Store, error) {
	s := &Store{}
	s.setOpts()
	var err error
	s.db, err = rdb.OpenDbForReadOnly(s.opt, filepath, false)
	return s, err
}

// Get returns the value given a key for RocksDB.
func (s *Store) Get(key []byte) (*rdb.Slice, error) {
	valSlice, err := s.db.Get(s.ropt, key)
	if err != nil {
		return nil, err
	}

	return valSlice, nil
}

// SetOne adds a key-value to data store.
func (s *Store) SetOne(k []byte, val []byte) error { return s.db.Put(s.wopt, k, val) }

// Delete deletes a key from data store.
func (s *Store) Delete(k []byte) error { return s.db.Delete(s.wopt, k) }

// NewIterator initializes a new iterator and returns it.
func (s *Store) NewIterator() *rdb.Iterator {
	ro := rdb.NewDefaultReadOptions()
	// SetFillCache should be set to false for bulk reads to avoid caching data
	// while doing bulk scans.
	ro.SetFillCache(false)
	return s.db.NewIterator(ro)
}

// Close closes our data store.
func (s *Store) Close() { s.db.Close() }

// Memtable returns the memtable size.
func (s *Store) MemtableSize() uint64 {
	memTableSize, _ := strconv.ParseUint(s.db.GetProperty("rocksdb.cur-size-all-mem-tables"), 10, 64)
	return memTableSize
}

// IndexFilterblockSize returns the filter block size.
func (s *Store) IndexFilterblockSize() uint64 {
	blockSize, _ := strconv.ParseUint(s.db.GetProperty("rocksdb.estimate-table-readers-mem"), 10, 64)
	return blockSize
}

// NewWriteBatch creates a new WriteBatch object and returns a pointer to it.
func (s *Store) NewWriteBatch() *rdb.WriteBatch { return rdb.NewWriteBatch() }

// WriteBatch does a batch write to RocksDB from the data in WriteBatch object.
func (s *Store) WriteBatch(wb *rdb.WriteBatch) error {
	return s.db.Write(s.wopt, wb)
}

// NewCheckpoint creates new checkpoint from current store.
func (s *Store) NewCheckpoint() (*rdb.Checkpoint, error) { return s.db.NewCheckpoint() }

// NewSnapshot creates new snapshot from current store.
func (s *Store) NewSnapshot() *rdb.Snapshot { return s.db.NewSnapshot() }

// SetSnapshot updates default read options to use the given snapshot.
func (s *Store) SetSnapshot(snapshot *rdb.Snapshot) { s.ropt.SetSnapshot(snapshot) }
