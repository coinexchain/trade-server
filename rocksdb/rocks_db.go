package rocksdb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/coinexchain/trade-server/core"
	"github.com/tecbot/gorocksdb"
	dbm "github.com/tendermint/tm-db"
)

// We use rocksdb's customizable compact filter to prune old records
type TimestampCompactionFilter struct {
	pruneTimestamp uint64
}

func (f *TimestampCompactionFilter) Name() string {
	return "TimestampPruneFilter"
}

var PruneableKeys map[byte]struct{}

func init() {
	PruneableKeys = make(map[byte]struct{})
	PruneableKeys[core.CandleStickByte] = struct{}{}
	PruneableKeys[core.DealByte] = struct{}{}
	PruneableKeys[core.OrderByte] = struct{}{}
	PruneableKeys[core.BancorInfoByte] = struct{}{}
	PruneableKeys[core.BancorTradeByte] = struct{}{}
	PruneableKeys[core.IncomeByte] = struct{}{}
	PruneableKeys[core.TxByte] = struct{}{}
	PruneableKeys[core.CommentByte] = struct{}{}
	PruneableKeys[core.SlashByte] = struct{}{}
	PruneableKeys[core.BancorDealByte] = struct{}{}
	PruneableKeys[core.RedelegationByte] = struct{}{}
	PruneableKeys[core.UnbondingByte] = struct{}{}
	PruneableKeys[core.UnlockByte] = struct{}{}
	PruneableKeys[core.LockedByte] = struct{}{}
	PruneableKeys[core.DonationByte] = struct{}{}
	PruneableKeys[core.DelistByte] = struct{}{}
	PruneableKeys[core.DetailByte] = struct{}{} // it's special
}

// If the corresponding time is too early, we prune the record
func (f *TimestampCompactionFilter) Filter(level int, key, val []byte) (remove bool, newVal []byte) {
	if f.pruneTimestamp == 0 {
		return false, val
	}
	if len(key) < 1 {
		return false, val
	}
	if _, ok := PruneableKeys[key[0]]; !ok {
		return false, val
	}
	start := len(key) - 1 /*lastByte*/ - 8 /*sid*/ - 8 /*timeBytes*/
	if key[0] == core.DetailByte {
		start = len(key) - 8 /*timeBytes*/
	}
	t := binary.BigEndian.Uint64(key[start : start+8])
	if f.pruneTimestamp > t {
		return true, nil
	}
	return false, val
}

var _ dbm.DB = (*RocksDB)(nil)

type RocksDB struct {
	db     *gorocksdb.DB
	ro     *gorocksdb.ReadOptions
	wo     *gorocksdb.WriteOptions
	woSync *gorocksdb.WriteOptions
	filter *TimestampCompactionFilter
}

func NewRocksDB(name string, dir string) (*RocksDB, error) {
	// default rocksdb option, good enough for most cases, including heavy workloads.
	// 1GB table cache, 512MB write buffer(may use 50% more on heavy workloads).
	// compression: snappy as default, need to -lsnappy to enable.
	bbto := gorocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetBlockCache(gorocksdb.NewLRUCache(1 << 30))
	bbto.SetFilterPolicy(gorocksdb.NewBloomFilter(10))

	opts := gorocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCreateIfMissing(true)
	opts.IncreaseParallelism(runtime.NumCPU())
	// 1.5GB maximum memory use for writebuffer.
	opts.OptimizeLevelStyleCompaction(512 * 1024 * 1024)
	return NewRocksDBWithOptions(name, dir, opts)
}

func NewRocksDBWithOptions(name string, dir string, opts *gorocksdb.Options) (*RocksDB, error) {
	dbPath := filepath.Join(dir, name+".db")
	filter := TimestampCompactionFilter{pruneTimestamp: 0}
	opts.SetCompactionFilter(&filter) // use a customized compaction filter
	db, err := gorocksdb.OpenDb(opts, dbPath)
	if err != nil {
		return nil, err
	}
	ro := gorocksdb.NewDefaultReadOptions()
	wo := gorocksdb.NewDefaultWriteOptions()
	woSync := gorocksdb.NewDefaultWriteOptions()
	woSync.SetSync(true)
	database := &RocksDB{
		db:     db,
		ro:     ro,
		wo:     wo,
		woSync: woSync,
		filter: &filter,
	}
	return database, nil
}

func (db *RocksDB) SetPruneTimestamp(t uint64) {
	if db.filter.pruneTimestamp < t {
		db.filter.pruneTimestamp = t
	}
}

func (db *RocksDB) GetPruneTimestamp() uint64 {
	return db.filter.pruneTimestamp
}

func (db *RocksDB) CompactRange(r gorocksdb.Range) {
	db.db.CompactRange(r)
}

// Implements DB.
func (db *RocksDB) Get(key []byte) []byte {
	key = nonNilBytes(key)
	res, err := db.db.Get(db.ro, key)
	if err != nil {
		panic(err)
	}
	return moveSliceToBytes(res)
}

// Implements DB.
func (db *RocksDB) Has(key []byte) bool {
	return db.Get(key) != nil
}

// Implements DB.
func (db *RocksDB) Set(key []byte, value []byte) {
	key = nonNilBytes(key)
	value = nonNilBytes(value)
	err := db.db.Put(db.wo, key, value)
	if err != nil {
		panic(err)
	}
}

// Implements DB.
func (db *RocksDB) SetSync(key []byte, value []byte) {
	key = nonNilBytes(key)
	value = nonNilBytes(value)
	err := db.db.Put(db.woSync, key, value)
	if err != nil {
		panic(err)
	}
}

// Implements DB.
func (db *RocksDB) Delete(key []byte) {
	key = nonNilBytes(key)
	err := db.db.Delete(db.wo, key)
	if err != nil {
		panic(err)
	}
}

// Implements DB.
func (db *RocksDB) DeleteSync(key []byte) {
	key = nonNilBytes(key)
	err := db.db.Delete(db.woSync, key)
	if err != nil {
		panic(err)
	}
}

func (db *RocksDB) DB() *gorocksdb.DB {
	return db.db
}

// Implements DB.
func (db *RocksDB) Close() {
	db.ro.Destroy()
	db.wo.Destroy()
	db.woSync.Destroy()
	db.db.Close()
}

// Implements DB.
func (db *RocksDB) Print() {
	itr := db.Iterator(nil, nil)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		key := itr.Key()
		value := itr.Value()
		fmt.Printf("[%X]:\t[%X]\n", key, value)
	}
}

// Implements DB.
func (db *RocksDB) Stats() map[string]string {
	keys := []string{"rocksdb.stats"}
	stats := make(map[string]string, len(keys))
	for _, key := range keys {
		stats[key] = db.db.GetProperty(key)
	}
	return stats
}

//----------------------------------------
// Batch

// Implements DB.
func (db *RocksDB) NewBatch() dbm.Batch {
	batch := gorocksdb.NewWriteBatch()
	return &rocksDBBatch{db, batch}
}

type rocksDBBatch struct {
	db    *RocksDB
	batch *gorocksdb.WriteBatch
}

// Implements Batch.
func (mBatch *rocksDBBatch) Set(key, value []byte) {
	mBatch.batch.Put(key, value)
}

// Implements Batch.
func (mBatch *rocksDBBatch) Delete(key []byte) {
	mBatch.batch.Delete(key)
}

// Implements Batch.
func (mBatch *rocksDBBatch) Write() {
	err := mBatch.db.db.Write(mBatch.db.wo, mBatch.batch)
	if err != nil {
		panic(err)
	}
}

// Implements Batch.
func (mBatch *rocksDBBatch) WriteSync() {
	err := mBatch.db.db.Write(mBatch.db.woSync, mBatch.batch)
	if err != nil {
		panic(err)
	}
}

// Implements Batch.
func (mBatch *rocksDBBatch) Close() {
	mBatch.batch.Destroy()
}

//----------------------------------------
// Iterator
// NOTE This is almost identical to db/go_level_db.Iterator
// Before creating a third version, refactor.

func (db *RocksDB) Iterator(start, end []byte) dbm.Iterator {
	itr := db.db.NewIterator(db.ro)
	return newRocksDBIterator(itr, start, end, false)
}

func (db *RocksDB) ReverseIterator(start, end []byte) dbm.Iterator {
	itr := db.db.NewIterator(db.ro)
	return newRocksDBIterator(itr, start, end, true)
}

var _ dbm.Iterator = (*rocksDBIterator)(nil)

type rocksDBIterator struct {
	source     *gorocksdb.Iterator
	start, end []byte
	isReverse  bool
	isInvalid  bool
}

func newRocksDBIterator(source *gorocksdb.Iterator, start, end []byte, isReverse bool) *rocksDBIterator {
	if isReverse {
		if end == nil {
			source.SeekToLast()
		} else {
			source.Seek(end)
			if source.Valid() {
				eoakey := moveSliceToBytes(source.Key()) // end or after key
				if bytes.Compare(end, eoakey) <= 0 {
					source.Prev()
				}
			} else {
				source.SeekToLast()
			}
		}
	} else {
		if start == nil {
			source.SeekToFirst()
		} else {
			source.Seek(start)
		}
	}
	return &rocksDBIterator{
		source:    source,
		start:     start,
		end:       end,
		isReverse: isReverse,
		isInvalid: false,
	}
}

func (itr rocksDBIterator) Domain() ([]byte, []byte) {
	return itr.start, itr.end
}

func (itr rocksDBIterator) Valid() bool {

	// Once invalid, forever invalid.
	if itr.isInvalid {
		return false
	}

	// Panic on DB error.  No way to recover.
	itr.assertNoError()

	// If source is invalid, invalid.
	if !itr.source.Valid() {
		itr.isInvalid = true
		return false
	}

	// If key is end or past it, invalid.
	var start = itr.start
	var end = itr.end
	var key = moveSliceToBytes(itr.source.Key())
	if itr.isReverse {
		if start != nil && bytes.Compare(key, start) < 0 {
			itr.isInvalid = true
			return false
		}
	} else {
		if end != nil && bytes.Compare(end, key) <= 0 {
			itr.isInvalid = true
			return false
		}
	}

	// It's valid.
	return true
}

func (itr rocksDBIterator) Key() []byte {
	itr.assertNoError()
	itr.assertIsValid()
	return moveSliceToBytes(itr.source.Key())
}

func (itr rocksDBIterator) Value() []byte {
	itr.assertNoError()
	itr.assertIsValid()
	return moveSliceToBytes(itr.source.Value())
}

func (itr rocksDBIterator) Next() {
	itr.assertNoError()
	itr.assertIsValid()
	if itr.isReverse {
		itr.source.Prev()
	} else {
		itr.source.Next()
	}
}

func (itr rocksDBIterator) Close() {
	itr.source.Close()
}

func (itr rocksDBIterator) assertNoError() {
	if err := itr.source.Err(); err != nil {
		panic(err)
	}
}

func (itr rocksDBIterator) assertIsValid() {
	if !itr.Valid() {
		panic("rocksDBIterator is invalid")
	}
}

// moveSliceToBytes will free the slice and copy out a go []byte
// This function can be applied on *Slice returned from Key() and Value()
// of an Iterator, because they are marked as freed.
func moveSliceToBytes(s *gorocksdb.Slice) []byte {
	defer s.Free()
	if !s.Exists() {
		return nil
	}
	v := make([]byte, len(s.Data()))
	copy(v, s.Data())
	return v
}

func nonNilBytes(bz []byte) []byte {
	if bz == nil {
		return []byte{}
	}
	return bz
}
