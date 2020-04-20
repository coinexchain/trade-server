package core

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func consumeMsg(hub *Hub, key, val string) {
	hub.ConsumeMessage(key, []byte(val))
	fillCommitInfo(hub)

	time.Sleep(time.Millisecond)
}

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
	subMan.CompareResult(t, fmt.Sprintf("1: %s", expectVal))
	subMan.ClearPushList()

	hub.currBlockHeight = 8
	val = fmt.Sprintf(model, "coinexdex-test2", 9, date.Unix())
	consumeMsg(hub, key, val)
	expectVal = fmt.Sprintf(model, "coinexdex-test2", 9, date.Unix())
	subMan.CompareResult(t, fmt.Sprintf("1: %s", expectVal))
	subMan.ClearPushList()
}

func TestParseBancorInfoForKafka(t *testing.T) {
	subMan := &MocSubscribeManager{}
	subMan.BancorInfoSubscribeInfo = make(map[string][]Subscriber)
	subMan.BancorInfoSubscribeInfo["abc/cet"] = make([]Subscriber, 1)
	subMan.BancorInfoSubscribeInfo["abc/cet"][0] = &PlainSubscriber{ID: 1}
	hub := getHub(t, subMan)
	defer os.RemoveAll("tmp")
	hub.oldChainID = "coinex-test"
	hub.chainID = "coinex-test"

	key := "bancor_info"
	val := `{"sender":"coinex1yj66ancalgk7dz3383s6cyvdd0nd93q0tk4x0c","stock":"abc","money":"cet","init_price":"1.000000000000000000","max_supply":"10000000000000","stock_precision":6,"max_price":"500.000000000000000000","price":"1.000000002994000000","stock_in_pool":"9999999999940","money_in_pool":"60","earliest_cancel_time":1917014400}`
	consumeMsg(hub, key, val)
	expectVal := `{"owner":"coinex1yj66ancalgk7dz3383s6cyvdd0nd93q0tk4x0c","stock":"abc","money":"cet","init_price":"1.000000000000000000","max_supply":"10000000000000","stock_precision":"6","max_price":"500.000000000000000000","current_price":"1.000000002994000000","stock_in_pool":"9999999999940","money_in_pool":"60","earliest_cancel_time":1917014400}`
	subMan.CompareResult(t, fmt.Sprintf("1: %s", expectVal))
	subMan.ClearPushList()

	hub.chainID = "coinex-test-new"
	consumeMsg(hub, key, expectVal)
	subMan.CompareResult(t, fmt.Sprintf("1: %s", expectVal))
	subMan.ClearPushList()
}

func TestParseNotificationBeginUnbonding(t *testing.T) {
	addr := "coinex1tlegt4y40m3qu3dd4zddmjf6u3rswdqk8xxvzw"
	subMan := &MocSubscribeManager{}
	subMan.UnbondingSubscribeInfo = make(map[string][]Subscriber)
	subMan.UnbondingSubscribeInfo[addr] = make([]Subscriber, 1)
	subMan.UnbondingSubscribeInfo[addr][0] = &PlainSubscriber{1}
	hub := getHub(t, subMan)
	defer os.RemoveAll("tmp")
	hub.oldChainID = "coinex-test"
	hub.chainID = "coinex-test"

	//u := int64(1566374450)
	//fmt.Println(time.Unix(u, 0).Format(time.RFC3339))

	key := "begin_unbonding"
	val := `{"delegator":"coinex1tlegt4y40m3qu3dd4zddmjf6u3rswdqk8xxvzw","validator":"coinexvaloper1yj66ancalgk7dz3383s6cyvdd0nd93q0sekwpv","amount":"100000","completion_time":"2019-08-21T16:00:50+08:00"}`
	consumeMsg(hub, key, val)
	key2 := "complete_unbonding"
	val2 := `{"delegator":"coinex1tlegt4y40m3qu3dd4zddmjf6u3rswdqk8xxvzw","validator":"coinexvaloper1yj66ancalgk7dz3383s6cyvdd0nd93q0sekwpv"}`
	hub.lastBlockTime = time.Unix(1566374450, 0)
	hub.currBlockTime = time.Unix(1566374456, 0)
	consumeMsg(hub, key2, val2)
	date, _ := time.Parse(time.RFC3339, "2019-08-21T16:00:50+08:00")
	mode := `{"delegator":"coinex1tlegt4y40m3qu3dd4zddmjf6u3rswdqk8xxvzw","validator":"coinexvaloper1yj66ancalgk7dz3383s6cyvdd0nd93q0sekwpv","amount":"100000","completion_time":%d}`
	expectVal := fmt.Sprintf(mode, date.Unix())
	subMan.CompareResult(t, fmt.Sprintf("1: %s", expectVal))
	subMan.ClearPushList()

	hub.chainID = "coinex-test-new"
	date, _ = time.Parse(time.RFC3339, "2019-08-23T16:00:50+08:00")
	expectVal = fmt.Sprintf(mode, date.Unix())
	consumeMsg(hub, key, expectVal)
	hub.lastBlockTime = date
	hub.currBlockTime = time.Unix(date.Unix()+6, 0)
	consumeMsg(hub, key2, val2)
	subMan.CompareResult(t, fmt.Sprintf("1: %s", expectVal))
}

func TestParseBeginRedelegation(t *testing.T) {
	addr := "coinex18rdsh78t4ds76p58kum34rye2pmrt3hj8z2ehg"
	subMan := &MocSubscribeManager{}
	subMan.RedelegationSubscribeInfo = make(map[string][]Subscriber)
	subMan.RedelegationSubscribeInfo[addr] = make([]Subscriber, 1)
	subMan.RedelegationSubscribeInfo[addr][0] = &PlainSubscriber{1}
	hub := getHub(t, subMan)
	defer os.RemoveAll("tmp")
	hub.oldChainID = "coinex-test"
	hub.chainID = "coinex-test"

	key := "begin_redelegation"
	val := `{"delegator":"coinex18rdsh78t4ds76p58kum34rye2pmrt3hj8z2ehg","src":"coinexvaloper1z6vr3s5nrn5d6fyxl5vmw77ehznme07w9dan6x","dst":"coinexvaloper16pr4xqlsglwu6urkyt975nxzl65hlt2fw0n58d","amount":"200000000000","completion_time":"2019-08-21T16:00:50+08:00"}`
	consumeMsg(hub, key, val)

	key2 := "complete_redelegation"
	val2 := `{"delegator":"coinex18rdsh78t4ds76p58kum34rye2pmrt3hj8z2ehg","src":"coinexvaloper1z6vr3s5nrn5d6fyxl5vmw77ehznme07w9dan6x","dst":"coinexvaloper16pr4xqlsglwu6urkyt975nxzl65hlt2fw0n58d"}`
	hub.lastBlockTime = time.Unix(1566374450, 0)
	hub.currBlockTime = time.Unix(1566374456, 0)
	consumeMsg(hub, key2, val2)
	date, _ := time.Parse(time.RFC3339, "2019-08-21T16:00:50+08:00")
	mode := `{"delegator":"coinex18rdsh78t4ds76p58kum34rye2pmrt3hj8z2ehg","src":"coinexvaloper1z6vr3s5nrn5d6fyxl5vmw77ehznme07w9dan6x","dst":"coinexvaloper16pr4xqlsglwu6urkyt975nxzl65hlt2fw0n58d","amount":"200000000000","completion_time":%d}`
	expectVal := fmt.Sprintf(mode, date.Unix())
	subMan.CompareResult(t, fmt.Sprintf("1: %s", expectVal))
	subMan.ClearPushList()

	hub.chainID = "coinex-test-new"
	date, _ = time.Parse(time.RFC3339, "2019-08-23T16:00:50+08:00")
	expectVal = fmt.Sprintf(mode, date.Unix())
	consumeMsg(hub, key, expectVal)
	hub.lastBlockTime = date
	hub.currBlockTime = time.Unix(date.Unix()+6, 0)
	consumeMsg(hub, key2, val2)
	subMan.CompareResult(t, fmt.Sprintf("1: %s", expectVal))
}
