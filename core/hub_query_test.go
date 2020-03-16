package core

import (
	"encoding/binary"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tm-db"
)

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

func storeHeightInfo(db dbm.DB, height, timestamp uint64) {
	heightBytes := Int64ToBigEndianBytes(int64(height))
	key := append([]byte{BlockHeightByte}, heightBytes...)
	val := make([]byte, 8)
	binary.LittleEndian.PutUint64(val[:], timestamp)
	db.Set(key, val)
}
