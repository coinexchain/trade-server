package server

import (
	"fmt"
	"testing"
	"time"

	"github.com/coinexchain/trade-server/core"
	toml "github.com/pelletier/go-toml"
	"github.com/stretchr/testify/require"
	db "github.com/tendermint/tm-db"
)

func TestNewConsumerWithMemBuf(t *testing.T) {
	memdb := db.NewMemDB()
	mocSub := &core.MocSubscribeManager{}
	mocSub.HeightSubscribeInfo = make([]core.Subscriber, 1)
	mocSub.HeightSubscribeInfo[0] = &core.PlainSubscriber{ID: 1}

	hub := core.NewHub(memdb, mocSub, 10, 9999999, 10, 4551831, "", 10)
	config := &toml.Tree{}
	consumer, err := NewConsumerWithMemBuf(config, hub)
	require.Nil(t, err)

	key := `height_info`
	val := `{"chain_id":"coinexdex2","height":4551832,"timestamp":1585565183,"last_block_hash":"5580756A652E674631D03791858837F0CE75DC348C7EC562D37DC5A097BC19EB"}`
	consumer.PutMsg([]byte(key), []byte(val))
	consumer.PutMsg([]byte("commit"), []byte("{}"))
	time.Sleep(time.Millisecond)
	mocSub.CompareResult(t, fmt.Sprintf("1: %s", val))
	mocSub.ClearPushList()

	val = `{"chain_id":"coinexdex2","height":4551833,"timestamp":1585565188,"last_block_hash":"0FD3111F517F8FBC6A9FD3905B53A1346F63CA66E602444611483EDFC7D995EB"}`
	consumer.PutMsg([]byte(key), []byte(val))
	consumer.PutMsg([]byte("commit"), []byte("{}"))
	time.Sleep(time.Millisecond)
	mocSub.CompareResult(t, fmt.Sprintf("1: %s", val))
}
