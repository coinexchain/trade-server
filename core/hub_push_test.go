package core

import (
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tm-db"
)

func getHub(t *testing.T, subMan *MocSubscribeManager) *Hub {
	db, err := dbm.NewGoLevelDB("test", "tmp")
	require.Nil(t, err)

	hub := NewHub(db, subMan, 1, 0, 1, 1, "", 0)
	hub.currBlockHeight = 5
	hub.StoreLeastHeight()
	hub.skipHeight = false
	hub.upgradeHeight = math.MaxInt64
	return hub
}

func fillCommitInfo(hub *Hub) {
	hub.ConsumeMessage("commit", []byte("{}"))
}

func consumeMsgAndCompareRet(t *testing.T, hub *Hub, subMan *MocSubscribeManager, key, val string) {
	hub.ConsumeMessage(key, []byte(val))
	fillCommitInfo(hub)

	time.Sleep(time.Millisecond)
	subMan.CompareResult(t, fmt.Sprintf("1: %s", val))
	subMan.ClearPushList()
}

func TestHub_PushHeightInfoMsg(t *testing.T) {
	subMan := &MocSubscribeManager{}
	subMan.HeightSubscribeInfo = make([]Subscriber, 1)
	subMan.HeightSubscribeInfo[0] = &PlainSubscriber{1}
	hub := getHub(t, subMan)
	hub.upgradeHeight = 5
	defer os.RemoveAll("tmp")

	// consumer height msg
	key := "height_info"
	val := `{"chain_id":"coinexdex-test1","height":6,"timestamp":283947,"last_block_hash":"1AEE872130EEA53168AD546A453BB343B4ABAE075949AF7AB995EF855790F5A4"}`
	consumeMsgAndCompareRet(t, hub, subMan, key, val)
}

func TestHub_PushBancorMsg(t *testing.T) {
	subMan := &MocSubscribeManager{}
	subMan.BancorInfoSubscribeInfo = make(map[string][]Subscriber)
	subMan.BancorInfoSubscribeInfo["abc/cet"] = make([]Subscriber, 1)
	subMan.BancorInfoSubscribeInfo["abc/cet"][0] = &PlainSubscriber{ID: 1}
	hub := getHub(t, subMan)
	hub.oldChainID = "chain"
	defer os.RemoveAll("tmp")

	key := "bancor_create"
	val := `{"owner":"coinex1yj66ancalgk7dz3383s6cyvdd0nd93q0tk4x0c","stock":"abc","money":"cet","init_price":"1.000000000000000000","max_supply":"10000000000000","max_price":"500.000000000000000000","current_price":"1.000000005988000000","stock_in_pool":"9999999999880","money_in_pool":"120","earliest_cancel_time":1917014400}`
	consumeMsgAndCompareRet(t, hub, subMan, key, val)

	key = "bancor_info"
	consumeMsgAndCompareRet(t, hub, subMan, key, val)
}

func TestHub_PushBancorTradeInfoMsg(t *testing.T) {
	addr1 := "coinex1x6rhu5m53fw8qgpwuljauaptvxyur57zym4jly"
	subMan := &MocSubscribeManager{}
	subMan.BancorTradeSubscribeInfo = make(map[string][]Subscriber)
	subMan.BancorTradeSubscribeInfo[addr1] = make([]Subscriber, 1)
	subMan.BancorTradeSubscribeInfo[addr1][0] = &PlainSubscriber{ID: 1}
	hub := getHub(t, subMan)
	defer os.RemoveAll("tmp")

	key := "bancor_trade"
	val := `{"sender":"coinex1x6rhu5m53fw8qgpwuljauaptvxyur57zym4jly","stock":"abc","money":"cet","amount":60,"side":1,"money_limit":100,"transaction_price":"5.300000000000000000","block_height":290}`
	consumeMsgAndCompareRet(t, hub, subMan, key, val)
}

func TestHub_PushBancorDealMsg(t *testing.T) {
	subMan := &MocSubscribeManager{}
	subMan.BancorDealSubscribeInfo = make(map[string][]Subscriber)
	subMan.BancorDealSubscribeInfo["abc/cet"] = make([]Subscriber, 1)
	subMan.BancorDealSubscribeInfo["abc/cet"][0] = &PlainSubscriber{1}
	hub := getHub(t, subMan)
	defer os.RemoveAll("tmp")

	key := "bancor_trade"
	val := `{"sender":"coinex1x6rhu5m53fw8qgpwuljauaptvxyur57zym4jly","stock":"abc","money":"cet","amount":60,"side":1,"money_limit":100,"transaction_price":"5.300000000000000000","block_height":290}`
	consumeMsgAndCompareRet(t, hub, subMan, key, val)
}

func TestHub_PushMarketInfoMsg(t *testing.T) {
	subMan := &MocSubscribeManager{}
	subMan.MarketSubscribeInfo = make(map[string][]Subscriber)
	subMan.MarketSubscribeInfo["abc/cet"] = make([]Subscriber, 1)
	subMan.MarketSubscribeInfo["abc/cet"][0] = &PlainSubscriber{1}
	hub := getHub(t, subMan)
	defer os.RemoveAll("tmp")

	key := "create_market_info"
	val := `{"stock":"abc","money":"cet","creator":"cettest1kwzuytmjmkd7x045s5r5k92hwrfgn2vhfn77wx","price_precision":3,"order_precision":0}`
	consumeMsgAndCompareRet(t, hub, subMan, key, val)
}

func TestHub_PushCreateOrderInfoMsg(t *testing.T) {
	addr1 := "coinex1x6rhu5m53fw8qgpwuljauaptvxyur57zym4jly"
	subMan := &MocSubscribeManager{}
	subMan.OrderSubscribeInfo = make(map[string][]Subscriber)
	subMan.OrderSubscribeInfo[addr1] = make([]Subscriber, 1)
	subMan.OrderSubscribeInfo[addr1][0] = &PlainSubscriber{1}
	hub := getHub(t, subMan)
	defer os.RemoveAll("tmp")

	// create order
	key := "create_order_info"
	val := `{"order_id":"coinex1x6rhu5m53fw8qgpwuljauaptvxyur57zym4jly-5203","sender":"coinex1x6rhu5m53fw8qgpwuljauaptvxyur57zym4jly","trading_pair":"abc/cet","order_type":2,"price":"8.200000000000000000","quantity":519727887,"side":1,"time_in_force":3,"feature_fee":0,"height":201,"frozen_fee":2754558,"freeze":4261768674}`
	consumeMsgAndCompareRet(t, hub, subMan, key, val)

	// del order
	key = "del_order_info"
	val = `{"order_id":"coinex1x6rhu5m53fw8qgpwuljauaptvxyur57zym4jly-5203","trading_pair":"abcd/cet","height":1199,"side":1,"price":"0.005200000000000000","del_reason":"IOC order cancel ","used_commission":1000000,"left_stock":10000000,"remain_amount":52000,"deal_stock":0,"deal_money":0}`
	consumeMsgAndCompareRet(t, hub, subMan, key, val)

	// fill order
	key = "fill_order_info"
	val = `{"order_id":"coinex1x6rhu5m53fw8qgpwuljauaptvxyur57zym4jly-5203","trading_pair":"abc/cet","height":201,"side":2,"price":"0.300000000000000000","left_stock":0,"freeze":0,"deal_stock":550729718,"deal_money":2918867505,"curr_stock":550729718,"curr_money":2918867505}`
	consumeMsgAndCompareRet(t, hub, subMan, key, val)
}

func TestHub_PushDealInfoMsg(t *testing.T) {
	subMan := &MocSubscribeManager{}
	subMan.DealSubscribeInfo = make(map[string][]Subscriber)
	subMan.DealSubscribeInfo["abc/cet"] = make([]Subscriber, 1)
	subMan.DealSubscribeInfo["abc/cet"][0] = &PlainSubscriber{1}
	hub := getHub(t, subMan)
	defer os.RemoveAll("tmp")

	// fill order
	key := "fill_order_info"
	val := `{"order_id":"coinex1x6rhu5m53fw8qgpwuljauaptvxyur57zym4jly-5203","trading_pair":"abc/cet","height":201,"side":2,"price":"0.300000000000000000","left_stock":0,"freeze":0,"deal_stock":550729718,"deal_money":2918867505,"curr_stock":550729718,"curr_money":2918867505}`
	consumeMsgAndCompareRet(t, hub, subMan, key, val)
}

func TestHub_PushSlashMsg(t *testing.T) {
	subMan := &MocSubscribeManager{}
	subMan.SlashSubscribeInfo = make([]Subscriber, 1)
	subMan.SlashSubscribeInfo[0] = &PlainSubscriber{1}
	hub := getHub(t, subMan)
	defer os.RemoveAll("tmp")

	key := "slash"
	val := `{"validator":"coinexvalcons1qwztwxzzndpdc94tujv8fux9phfenqmvx296zw","power":"1000000","reason":"double_sign","jailed":false}`
	consumeMsgAndCompareRet(t, hub, subMan, key, val)
}

func TestHub_PushTxMsg(t *testing.T) {
	addr := "coinex18rdsh78t4ds76p58kum34rye2pmrt3hj8z2ehg"
	subMan := &MocSubscribeManager{}
	subMan.TxSubscribeInfo = make(map[string][]Subscriber)
	subMan.TxSubscribeInfo[addr] = make([]Subscriber, 1)
	subMan.TxSubscribeInfo[addr][0] = &PlainSubscriber{1}
	hub := getHub(t, subMan)
	defer os.RemoveAll("tmp")

	key := "notify_tx"
	val := `{"signers":["coinex18rdsh78t4ds76p58kum34rye2pmrt3hj8z2ehg"],"transfers":[{"sender":"coinex1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8vc4efa","recipient":"coinex18rdsh78t4ds76p58kum34rye2pmrt3hj8z2ehg","amount":"46591208cet"}],"serial_number":5,"msg_types":["MsgBeginRedelegate"],"tx_json":"{\"msg\":[{\"delegator_address\":\"coinex18rdsh78t4ds76p58kum34rye2pmrt3hj8z2ehg\",\"validator_src_address\":\"coinexvaloper1z6vr3s5nrn5d6fyxl5vmw77ehznme07w9dan6x\",\"validator_dst_address\":\"coinexvaloper16pr4xqlsglwu6urkyt975nxzl65hlt2fw0n58d\",\"amount\":{\"denom\":\"cet\",\"amount\":\"200000000000\"}}],\"fee\":{\"amount\":[{\"denom\":\"cet\",\"amount\":\"100000000\"}],\"gas\":300000},\"signatures\":[{\"pub_key\":[3,186,173,114,10,114,60,69,148,17,209,55,10,57,127,90,126,132,109,131,225,144,113,107,12,103,145,187,141,175,133,172,235],\"signature\":\"QdoVUjjRSxI+ob/DBCGYlFhysh4w2acDRDwrH28CAUBNFapbQlXKAPB9nG5aIzR9dLIhWhOaTmhfsCve/ak0SA==\"}],\"memo\":\"用户user0发起从validator1到validator2的2000_0000_0000sato.CET的redelegation\"}","height":35,"hash":""}`
	consumeMsgAndCompareRet(t, hub, subMan, key, val)
}

func TestHub_PushIncomeMsg(t *testing.T) {
	addr := "coinex18rdsh78t4ds76p58kum34rye2pmrt3hj8z2ehg"
	subMan := &MocSubscribeManager{}
	subMan.IncomeSubscribeInfo = make(map[string][]Subscriber)
	subMan.IncomeSubscribeInfo[addr] = make([]Subscriber, 1)
	subMan.IncomeSubscribeInfo[addr][0] = &PlainSubscriber{1}
	hub := getHub(t, subMan)
	defer os.RemoveAll("tmp")

	key := "notify_tx"
	val := `{"signers":["coinex18rdsh78t4ds76p58kum34rye2pmrt3hj8z2ehg"],"transfers":[{"sender":"coinex1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8vc4efa","recipient":"coinex18rdsh78t4ds76p58kum34rye2pmrt3hj8z2ehg","amount":"46591208cet"}],"serial_number":5,"msg_types":["MsgBeginRedelegate"],"tx_json":"{\"msg\":[{\"delegator_address\":\"coinex18rdsh78t4ds76p58kum34rye2pmrt3hj8z2ehg\",\"validator_src_address\":\"coinexvaloper1z6vr3s5nrn5d6fyxl5vmw77ehznme07w9dan6x\",\"validator_dst_address\":\"coinexvaloper16pr4xqlsglwu6urkyt975nxzl65hlt2fw0n58d\",\"amount\":{\"denom\":\"cet\",\"amount\":\"200000000000\"}}],\"fee\":{\"amount\":[{\"denom\":\"cet\",\"amount\":\"100000000\"}],\"gas\":300000},\"signatures\":[{\"pub_key\":[3,186,173,114,10,114,60,69,148,17,209,55,10,57,127,90,126,132,109,131,225,144,113,107,12,103,145,187,141,175,133,172,235],\"signature\":\"QdoVUjjRSxI+ob/DBCGYlFhysh4w2acDRDwrH28CAUBNFapbQlXKAPB9nG5aIzR9dLIhWhOaTmhfsCve/ak0SA==\"}],\"memo\":\"用户user0发起从validator1到validator2的2000_0000_0000sato.CET的redelegation\"}","height":35,"hash":""}`
	consumeMsgAndCompareRet(t, hub, subMan, key, val)
}

func TestHub_PushLockedCoinsMsg(t *testing.T) {
	addr := "coinex1rafnyd9j9gc9cwu5q5uflefpdn62awyl7rvh8t"
	subMan := &MocSubscribeManager{}
	subMan.LockedSubcribeInfo = make(map[string][]Subscriber)
	subMan.LockedSubcribeInfo[addr] = make([]Subscriber, 1)
	subMan.LockedSubcribeInfo[addr][0] = &PlainSubscriber{1}
	hub := getHub(t, subMan)
	defer os.RemoveAll("tmp")

	key := "send_lock_coins"
	val := `{"from_address":"coinex1avmxlmztzxap20hawpc85h3uzj3277ja88wec2","to_address":"coinex1rafnyd9j9gc9cwu5q5uflefpdn62awyl7rvh8t","amount":[{"denom":"cet","amount":"1000000"}],"unlock_time":1567080885}`
	consumeMsgAndCompareRet(t, hub, subMan, key, val)
}

func TestHub_PushRedelegationMsg(t *testing.T) {
	addr := "coinex18rdsh78t4ds76p58kum34rye2pmrt3hj8z2ehg"
	subMan := &MocSubscribeManager{}
	subMan.RedelegationSubscribeInfo = make(map[string][]Subscriber)
	subMan.RedelegationSubscribeInfo[addr] = make([]Subscriber, 1)
	subMan.RedelegationSubscribeInfo[addr][0] = &PlainSubscriber{1}
	hub := getHub(t, subMan)
	hub.oldChainID = "chain"
	defer os.RemoveAll("tmp")

	key := "begin_redelegation"
	val := `{"delegator":"coinex18rdsh78t4ds76p58kum34rye2pmrt3hj8z2ehg","src":"coinexvaloper1z6vr3s5nrn5d6fyxl5vmw77ehznme07w9dan6x","dst":"coinexvaloper16pr4xqlsglwu6urkyt975nxzl65hlt2fw0n58d","amount":"200000000000","completion_time":1566374450}`
	hub.ConsumeMessage(key, []byte(val))
	fillCommitInfo(hub)

	key = "complete_redelegation"
	val2 := `{"delegator":"coinex18rdsh78t4ds76p58kum34rye2pmrt3hj8z2ehg","src":"coinexvaloper1z6vr3s5nrn5d6fyxl5vmw77ehznme07w9dan6x","dst":"coinexvaloper16pr4xqlsglwu6urkyt975nxzl65hlt2fw0n58d"}`
	hub.lastBlockTime = time.Unix(1566374450, 0)
	hub.currBlockTime = time.Unix(1566374456, 0)
	hub.ConsumeMessage(key, []byte(val2))
	fillCommitInfo(hub)

	time.Sleep(time.Millisecond)
	subMan.CompareResult(t, fmt.Sprintf("1: %s", val))
}

func TestHub_PushUnbondingMsg(t *testing.T) {
	addr := "coinex1tlegt4y40m3qu3dd4zddmjf6u3rswdqk8xxvzw"
	subMan := &MocSubscribeManager{}
	subMan.UnbondingSubscribeInfo = make(map[string][]Subscriber)
	subMan.UnbondingSubscribeInfo[addr] = make([]Subscriber, 1)
	subMan.UnbondingSubscribeInfo[addr][0] = &PlainSubscriber{1}
	hub := getHub(t, subMan)
	hub.chainID = "chain"
	defer os.RemoveAll("tmp")

	key := "begin_unbonding"
	val := `{"delegator":"coinex1tlegt4y40m3qu3dd4zddmjf6u3rswdqk8xxvzw","validator":"coinexvaloper1yj66ancalgk7dz3383s6cyvdd0nd93q0sekwpv","amount":"100000","completion_time":1566374450}`
	hub.ConsumeMessage(key, []byte(val))
	fillCommitInfo(hub)

	key = "complete_unbonding"
	val2 := `{"delegator":"coinex1tlegt4y40m3qu3dd4zddmjf6u3rswdqk8xxvzw","validator":"coinexvaloper1yj66ancalgk7dz3383s6cyvdd0nd93q0sekwpv"}`
	hub.lastBlockTime = time.Unix(1566374450, 0)
	hub.currBlockTime = time.Unix(1566374456, 0)
	hub.ConsumeMessage(key, []byte(val2))
	fillCommitInfo(hub)

	time.Sleep(time.Millisecond)
	subMan.CompareResult(t, fmt.Sprintf("1: %s", val))
}

func TestHub_PushUnlockMsg(t *testing.T) {
	addr := "coinex1tlegt4y40m3qu3dd4zddmjf6u3rswdqk8xxvzw"
	subMan := &MocSubscribeManager{}
	subMan.UnlockSubscribeInfo = make(map[string][]Subscriber)
	subMan.UnlockSubscribeInfo[addr] = make([]Subscriber, 1)
	subMan.UnlockSubscribeInfo[addr][0] = &PlainSubscriber{1}
	hub := getHub(t, subMan)
	defer os.RemoveAll("tmp")

	key := "notify_unlock"
	val := `{"address":"coinex1tlegt4y40m3qu3dd4zddmjf6u3rswdqk8xxvzw","unlocked":[{"denom":"cet","amount":"1000000"}],"locked_coins":null,"frozen_coins":[{"denom":"abc","amount":"796912961248"},{"denom":"cet","amount":"1896049635319"}],"coins":[{"denom":"abc","amount":"24999230553270264"},{"denom":"cet","amount":"24971271794985539"}],"height":669}`
	consumeMsgAndCompareRet(t, hub, subMan, key, val)
}

func TestHub_PushCommentMsg(t *testing.T) {
	subMan := &MocSubscribeManager{}
	subMan.CommentSubscribeInfo = make(map[string][]Subscriber)
	subMan.CommentSubscribeInfo["cet"] = make([]Subscriber, 1)
	subMan.CommentSubscribeInfo["cet"][0] = &PlainSubscriber{1}
	hub := getHub(t, subMan)
	defer os.RemoveAll("tmp")

	key := "token_comment"
	val := `{"id":0,"height":10,"sender":"coinex1py9lss4nr0lm6ep4uwk3tclacw42a5nx0ra92r","token":"cet","donation":200000000,"title":"I-Love-CET","content":"cet-to-the-moon","content_type":3,"references":null}`
	consumeMsgAndCompareRet(t, hub, subMan, key, val)
}
