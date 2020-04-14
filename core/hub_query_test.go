package core

import (
	"encoding/binary"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tm-db"
)

func storeHeightInfo(db dbm.DB, height, timestamp uint64) {
	heightBytes := Int64ToBigEndianBytes(int64(height))
	key := append([]byte{BlockHeightByte}, heightBytes...)
	val := make([]byte, 8)
	binary.LittleEndian.PutUint64(val[:], timestamp)
	db.Set(key, val)
}

func storeLatestHeight(db dbm.DB, height int64) {
	heightBytes := Int64ToBigEndianBytes(height)
	db.Set([]byte{LatestHeightByte}, heightBytes)
}

func TestQueryBlockInfo(t *testing.T) {
	db, err := dbm.NewGoLevelDB("test", "tmp")
	require.Nil(t, err)
	defer func() {
		os.RemoveAll("tmp")
	}()
	hub := Hub{db: db}

	// store height
	storeHeightInfo(db, 3, 1990)
	storeHeightInfo(db, 4, 1991)
	storeHeightInfo(db, 5, 1992)

	// query block info
	info := hub.QueryBlockInfo()
	require.EqualValues(t, 5, info.Height)
	require.EqualValues(t, 1992, info.TimeStamp)
}

func TestQueryLatestHeight(t *testing.T) {
	db, err := dbm.NewGoLevelDB("test", "tmp")
	require.Nil(t, err)
	defer func() {
		os.RemoveAll("tmp")
	}()
	hub := Hub{db: db}

	// store latest height
	storeLatestHeight(db, 1000)
	storeLatestHeight(db, 1001)

	// query latest height
	height := hub.QueryLatestHeight()
	require.EqualValues(t, 1001, height)
}

func TestQueryBlockTime(t *testing.T) {
	db, err := dbm.NewGoLevelDB("test", "tmp")
	require.Nil(t, err)
	defer func() {
		os.RemoveAll("tmp")
	}()
	hub := Hub{db: db}

	// store height
	storeHeightInfo(db, 0, 1990)
	storeHeightInfo(db, 1, 1991)
	storeHeightInfo(db, 2, 1992)
	storeHeightInfo(db, 3, 1993)
	storeHeightInfo(db, 4, 1994)
	storeHeightInfo(db, 5, 1995)

	// query latest height
	blockTimes := hub.QueryBlockTime(5, 100)
	bytes, _ := json.Marshal(blockTimes)
	correct := `[1995,1994,1993,1992,1991,1990]`
	require.EqualValues(t, correct, string(bytes))
}
