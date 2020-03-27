package core

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestParseHeightInfo(t *testing.T) {
	subMan := &MocSubscribeManager{}
	subMan.HeightSubscribeInfo = make([]Subscriber, 1)
	subMan.HeightSubscribeInfo[0] = &PlainSubscriber{1}
	hub := getHub(t, subMan)
	defer os.RemoveAll("tmp")
	hub.oldChainID = "coinexdex-test1"

	//timeStr := "2019-08-21T08:00:51.648298Z"
	key := "height_info"
	val := `{"chain_id":"coinexdex-test1","height":6,"timestamp":"2019-08-21T08:00:51.648298Z","last_block_hash":"1AEE872130EEA53168AD546A453BB343B4ABAE075949AF7AB995EF855790F5A4"}`
	consumeMsg(hub, key, val)

	date, _ := time.Parse(time.RFC3339, "2019-08-21T08:00:51.648298Z")
	model := `{"chain_id":"%s","height":%d,"timestamp":%d,"last_block_hash":"1AEE872130EEA53168AD546A453BB343B4ABAE075949AF7AB995EF855790F5A4"}`
	expectVal := fmt.Sprintf(model, "coinexdex-test1", 6, date.Unix())
	subMan.compareResult(t, fmt.Sprintf("1: %s", expectVal))
	subMan.clearPushList()

	hub.currBlockHeight = 8
	val = fmt.Sprintf(model, "coinexdex-test2", 9, date.Unix())
	consumeMsg(hub, key, val)
	expectVal = fmt.Sprintf(model, "coinexdex-test2", 9, date.Unix())
	subMan.compareResult(t, fmt.Sprintf("1: %s", expectVal))
	subMan.clearPushList()
}

func consumeMsg(hub *Hub, key, val string) {
	hub.ConsumeMessage(key, []byte(val))
	fillCommitInfo(hub)

	time.Sleep(time.Millisecond)
}
