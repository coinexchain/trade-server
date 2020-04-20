package server

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coinexchain/trade-server/core"
	toml "github.com/pelletier/go-toml"
	db "github.com/tendermint/tm-db"
)

func TestNewConsumerWithMemBuf(t *testing.T) {
	memdb := db.NewMemDB()
	mocSub := &core.MocSubscribeManager{}
	hub := core.NewHub(memdb, mocSub, 10, 10, 10, 10, "", 10)
	config := &toml.Tree{}
	consumer, err := NewConsumerWithMemBuf(config, hub)
	require.Nil(t, err)

	key := `height_info`
	val := `{"height":1199,"timestamp":"2019-09-10T08:28:46.518409Z","last_block_hash":"237C23E275808F6FC50466C48231FF3449665B6B1D7819DA3FF363CFFC5323A9"}`
	consumer.PutMsg([]byte(key), []byte(val))
	consumer.PutMsg([]byte("commit"), []byte("{}"))

	val = `{"height":1399,"timestamp":"2019-09-10T08:28:46.518409Z","last_block_hash":"237C23E275808F6FC50466C48231FF3449665B6B1D7819DA3FF363CFFC5323A9"}`
	consumer.PutMsg([]byte(key), []byte(val))
	consumer.PutMsg([]byte("commit"), []byte("{}"))

}

func TestSliceHub(t *testing.T) {
	data := make([]int, 10)
	for i := 0; i < 10; i++ {
		data[i] = i + 1
	}
	fmt.Printf("%p\n", data)
	sliceHub(data)
}

func sliceHub(data []int) {
	data = data[:0]
	fmt.Println("end, ", data)
	fmt.Printf("%p\n", data)
}
