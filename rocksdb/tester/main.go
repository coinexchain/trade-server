package main

import (
	"encoding/binary"
	"os"
	"path/filepath"

	"github.com/coinexchain/trade-server/rocksdb"
	"github.com/tecbot/gorocksdb"
)

func main() {
	db, err := rocksdb.NewRocksDB("test", ".")
	if err != nil {
		panic(err)
	}
	var buf [8]byte
	for i := int64(1000 * 1000); i > 0; i-- {
		binary.BigEndian.PutUint64(buf[:], uint64(i))
		db.Set(buf[:], buf[:])
	}
	db.SetSync(buf[:], buf[:])
	db.CompactRange(gorocksdb.Range{Start: []byte{}, Limit: []byte{255, 255, 255, 255, 255, 255, 255, 255}})
	db.Close()
	os.RemoveAll(filepath.Join(".", "test"+".db"))
}
