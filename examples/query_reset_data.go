package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	db "github.com/tendermint/tm-db"

	"github.com/coinexchain/trade-server/core"
	"github.com/coinexchain/trade-server/server"
	toml "github.com/pelletier/go-toml"
)

func main() {
	db := initDB()
	defer db.Close()

	QueryHubDumpData(db)
	//ResetOffset(db)
}

func initDB() db.DB {
	content := `"data-dir" = "/Users/matrix/cetchain/dex/data"
"use-rocksdb" = false`

	conf, err := toml.Load(content)
	if err != nil {
		panic(err)
	}
	fmt.Println(conf.Get("data-dir"))
	fmt.Println(conf.Get("use-rocksdb"))
	db, err := server.InitDB(conf)
	if err != nil {
		panic(err)
	}

	return db
}

func QueryHubDumpData(db db.DB) {
	hub := core.NewHub(db, nil, 1, 2, 3, 4, "", 2)
	if hub == nil {
		panic("init hub failed")
	}
	dumpData, err := server.GetHubDumpData(hub)
	if err != nil {
		panic("get hub data failed")
	}
	bz, _ := json.Marshal(dumpData)
	fmt.Println(string(bz))
}

func ResetOffset(db db.DB) {
	offsetKey := core.GetOffsetKey(0)
	if !db.Has(offsetKey) {
		panic("no data in store")
	}

	// get original
	offsetBuf := db.Get(offsetKey)
	fmt.Println(offsetBuf)
	fmt.Printf("original offset: %d\n", binary.BigEndian.Uint64(offsetBuf))

	// set new offset
	offsetBuf = core.Int64ToBigEndianBytes(0)
	db.SetSync(offsetKey, offsetBuf)

	// validator
	offsetBuf = db.Get(offsetKey)
	fmt.Printf("Update offset: %d\n", binary.BigEndian.Uint64(offsetBuf))
	if binary.BigEndian.Uint64(offsetBuf) != 0 {
		panic("should be 0")
	}
}
