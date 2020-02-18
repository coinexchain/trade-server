package main

import (
	"encoding/binary"

	"github.com/tecbot/gorocksdb"
	"github.com/coinexchain/trade-server/rocksdb"
)

func main() {
	db, err := rocksdb.NewRocksDB("test", "rocksdb")
	if err != nil {
		panic(err)
	}
	var buf [8]byte
	for i := int64(1000*1000); i > 0; i-- {
		binary.BigEndian.PutUint64(buf[:], uint64(i))
		db.Set(buf[:], buf[:])
	}
	db.SetSync(buf[:], buf[:])
	db.CompactRange(gorocksdb.Range{Start: []byte{}, Limit: []byte{255,255,255,255,255,255,255,255}})
	db.Close()
}
