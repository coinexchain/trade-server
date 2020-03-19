package core

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tm-db"
)

func getHub(t *testing.T, subMan *MocSubscribeManager) *Hub {

	db, err := dbm.NewGoLevelDB("test", "tmp")
	require.Nil(t, err)

	hub := NewHub(db, subMan, 1, 0, 1, 1)
	hub.currBlockHeight = 5
	hub.StoreLeastHeight()
	hub.skipHeight = false
	return hub
}

func fillCommitInfo(hub *Hub) {
	hub.ConsumeMessage("commit", []byte("{}"))
}

func TestHub_PushHeightInfoMsg(t *testing.T) {
	subMan := MocSubscribeManager{}
	subMan.HeightSubscribeInfo = make([]Subscriber, 1)
	subMan.HeightSubscribeInfo[0] = &PlainSubscriber{1}
	hub := getHub(t, &subMan)
	defer os.RemoveAll("tmp")

	// consumer height msg
	key := "height_info"
	val := `{"chain_id":"coinexdex-test1","height":6,"timestamp":"2020-03-10T10:56:57.067926Z","last_block_hash":"1AEE872130EEA53168AD546A453BB343B4ABAE075949AF7AB995EF855790F5A4"}`
	hub.ConsumeMessage(key, []byte(val))
	fillCommitInfo(hub)

	time.Sleep(time.Millisecond)
	subMan.compareResult(t, fmt.Sprintf("1: %s", val))
}

func TestHub_PushBancorTradeInfoMsg(t *testing.T) {
	addr1 := "coinex1x6rhu5m53fw8qgpwuljauaptvxyur57zym4jly"
	subMan := MocSubscribeManager{}
	subMan.BancorTradeSubscribeInfo = make(map[string][]Subscriber)
	subMan.BancorTradeSubscribeInfo[addr1] = make([]Subscriber, 1)
	subMan.BancorTradeSubscribeInfo[addr1][0] = &PlainSubscriber{ID: 1}
	hub := getHub(t, &subMan)
	defer os.RemoveAll("tmp")

	key := "bancor_trade"
	val := `{"sender":"coinex1x6rhu5m53fw8qgpwuljauaptvxyur57zym4jly","stock":"abc","money":"cet","amount":60,"side":1,"money_limit":100,"transaction_price":"5.300000000000000000","block_height":290}`
	hub.ConsumeMessage(key, []byte(val))
	fillCommitInfo(hub)

	time.Sleep(time.Millisecond)
	subMan.compareResult(t, fmt.Sprintf("1: %s", val))
}

func TestHub_PushBancorDealMsg(t *testing.T) {
	subMan := MocSubscribeManager{}
	subMan.BancorDealSubscribeInfo = make(map[string][]Subscriber)
	subMan.BancorDealSubscribeInfo["abc/cet"] = make([]Subscriber, 1)
	subMan.BancorDealSubscribeInfo["abc/cet"][0] = &PlainSubscriber{1}
	hub := getHub(t, &subMan)
	defer os.RemoveAll("tmp")

	key := "bancor_trade"
	val := `{"sender":"coinex1x6rhu5m53fw8qgpwuljauaptvxyur57zym4jly","stock":"abc","money":"cet","amount":60,"side":1,"money_limit":100,"transaction_price":"5.300000000000000000","block_height":290}`
	hub.ConsumeMessage(key, []byte(val))
	fillCommitInfo(hub)

	time.Sleep(time.Millisecond)
	subMan.compareResult(t, fmt.Sprintf("1: %s", val))
}
