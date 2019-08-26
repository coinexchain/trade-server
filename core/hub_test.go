package core

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tm-db"
	"strings"
	"testing"
)

func simpleAddr(s string) (sdk.AccAddress, error) {
	return sdk.AccAddressFromHex("01234567890123456789012345678901234" + s)
}

func toStr(payload [][]byte) string {
	out := make([]string, len(payload))
	for i := 0; i < len(out); i++ {
		out[i] = string(payload[i])
	}
	return strings.Join(out, "\n")
}

func Test1(t *testing.T) {
	acc1, _ := simpleAddr("00001")
	acc2, _ := simpleAddr("00002")
	addr1 := acc1.String()
	addr2 := acc2.String()

	db := dbm.NewMemDB()
	subMan := GetSubscribeManager(addr1, addr2)
	hub := NewHub(db, subMan)

	newHeightInfo := &NewHeightInfo{
		Height:        1000,
		TimeStamp:     T("2019-07-15T08:07:10Z"),
		LastBlockHash: []byte("01234567890123456789"),
	}
	bytes, _ := json.Marshal(newHeightInfo)
	hub.ConsumeMessage("height_info", bytes)

	correct := `
3: {"height":1000,"timestamp":"2019-07-15T08:07:10Z","last_block_hash":"3031323334353637383930313233343536373839"}
4: {"height":1000,"timestamp":"2019-07-15T08:07:10Z","last_block_hash":"3031323334353637383930313233343536373839"}
`
	subMan.compareResult(t, correct)
	subMan.clearPushList()

	notificationSlash := &NotificationSlash{
		Validator: "Val1",
		Power:     "30%",
		Reason:    "double_sign",
		Jailed:    true,
	}
	bytes, _ = json.Marshal(notificationSlash)
	hub.ConsumeMessage("notify_slash", bytes)

	transRec := TransferRecord{
		Sender:    addr1,
		Recipient: addr2,
		Amount:    "1cet",
	}
	notificationTx := &NotificationTx{
		Signers:      []string{addr1},
		Transfers:    []TransferRecord{transRec},
		SerialNumber: 20000,
		MsgTypes:     []string{"MsgType1"},
		TxJSON:       "blabla",
		Height:       1001,
	}
	bytes, _ = json.Marshal(notificationTx)
	hub.ConsumeMessage("notify_tx", bytes)
	correct = `
25: {"signers":["cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca"],"transfers":[{"sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","recipient":"cosmos1qy352eufqy352eufqy352eufqy35qqqz9ayrkz","amount":"1cet"}],"serial_number":20000,"msg_types":["MsgType1"],"tx_json":"blabla","height":1001}
20: {"signers":["cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca"],"transfers":[{"sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","recipient":"cosmos1qy352eufqy352eufqy352eufqy35qqqz9ayrkz","amount":"1cet"}],"serial_number":20000,"msg_types":["MsgType1"],"tx_json":"blabla","height":1001}
`
	subMan.compareResult(t, correct)
	subMan.clearPushList()

	lockedCoins := LockedCoins{
		{
			Coin:       sdk.Coin{Denom: "cet", Amount: sdk.NewInt(5000)},
			UnlockTime: T("2019-07-15T08:18:10Z").Unix(),
		},
	}
	notificationUnlock := &NotificationUnlock{
		Address: addr2,
		Unlocked: sdk.Coins{
			{Denom: "abc", Amount: sdk.NewInt(15000)},
		},
		LockedCoins: lockedCoins,
		FrozenCoins: sdk.Coins{},
		Coins:       sdk.Coins{},
		Height:      1001,
	}
	bytes, _ = json.Marshal(notificationUnlock)
	hub.ConsumeMessage("notify_unlock", bytes)
	correct = `
23: {"address":"cosmos1qy352eufqy352eufqy352eufqy35qqqz9ayrkz","unlocked":[{"denom":"abc","amount":"15000"}],"locked_coins":[{"coin":{"denom":"cet","amount":"5000"},"unlock_time":1563178690}],"frozen_coins":[],"coins":[],"height":1001}
24: {"address":"cosmos1qy352eufqy352eufqy352eufqy35qqqz9ayrkz","unlocked":[{"denom":"abc","amount":"15000"}],"locked_coins":[{"coin":{"denom":"cet","amount":"5000"},"unlock_time":1563178690}],"frozen_coins":[],"coins":[],"height":1001}
`
	subMan.compareResult(t, correct)
	subMan.clearPushList()

	notificationBeginRedelegation := &NotificationBeginRedelegation{
		Delegator:      addr2,
		ValidatorSrc:   "Val1",
		ValidatorDst:   "Val2",
		Amount:         "500",
		CompletionTime: "2019-07-15T08:18:10Z",
	}
	bytes, _ = json.Marshal(notificationBeginRedelegation)
	hub.ConsumeMessage("begin_redelegation", bytes)

	notificationBeginUnbonding := &NotificationBeginUnbonding{
		Delegator:      addr1,
		Validator:      "Val1",
		Amount:         "300",
		CompletionTime: "2019-07-15T08:18:10Z",
	}
	bytes, _ = json.Marshal(notificationBeginUnbonding)
	hub.ConsumeMessage("begin_unbonding", bytes)

	createOrderInfo := &CreateOrderInfo{
		OrderID:     addr1 + "-1",
		Sender:      addr1,
		TradingPair: "abc/cet",
		OrderType:   LIMIT,
		Price:       sdk.NewDec(100),
		Quantity:    300,
		Side:        SELL,
		TimeInForce: GTE,
		FeatureFee:  1,
		Height:      1001,
		FrozenFee:   1,
		Freeze:      10,
	}
	bytes, _ = json.Marshal(createOrderInfo)
	hub.ConsumeMessage("create_order_info", bytes)

	createOrderInfo = &CreateOrderInfo{
		OrderID:     addr1 + "-2",
		Sender:      addr1,
		TradingPair: "abc/cet",
		OrderType:   LIMIT,
		Price:       sdk.NewDec(100),
		Quantity:    300,
		Side:        BUY,
		TimeInForce: GTE,
		FeatureFee:  1,
		Height:      1001,
		FrozenFee:   1,
		Freeze:      10,
	}
	bytes, _ = json.Marshal(createOrderInfo)
	hub.ConsumeMessage("create_order_info", bytes)
	correct = `
15: {"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-1","sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","trading_pair":"abc/cet","order_type":2,"price":"100.000000000000000000","quantity":300,"side":2,"time_in_force":3,"feature_fee":1,"height":1001,"frozen_fee":1,"freeze":10}
16: {"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-1","sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","trading_pair":"abc/cet","order_type":2,"price":"100.000000000000000000","quantity":300,"side":2,"time_in_force":3,"feature_fee":1,"height":1001,"frozen_fee":1,"freeze":10}
15: {"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-2","sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","trading_pair":"abc/cet","order_type":2,"price":"100.000000000000000000","quantity":300,"side":1,"time_in_force":3,"feature_fee":1,"height":1001,"frozen_fee":1,"freeze":10}
16: {"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-2","sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","trading_pair":"abc/cet","order_type":2,"price":"100.000000000000000000","quantity":300,"side":1,"time_in_force":3,"feature_fee":1,"height":1001,"frozen_fee":1,"freeze":10}
`
	subMan.compareResult(t, correct)
	subMan.clearPushList()

	hub.ConsumeMessage("commit", nil)
	correct = `
0: {"validator":"Val1","power":"30%","reason":"double_sign","jailed":true}
1: {"validator":"Val1","power":"30%","reason":"double_sign","jailed":true}
8: {"type":"depth","payload":{"trading_pair":"abc/cet","bids":null,"asks":[{"p":"100.000000000000000000","a":"300"}]}}
8: {"type":"depth","payload":{"trading_pair":"abc/cet","bids":[{"p":"100.000000000000000000","a":"300"}],"asks":null}}
9: {"type":"depth","payload":{"trading_pair":"abc/cet","bids":null,"asks":[{"p":"100.000000000000000000","a":"300"}]}}
9: {"type":"depth","payload":{"trading_pair":"abc/cet","bids":[{"p":"100.000000000000000000","a":"300"}],"asks":null}}
`
	subMan.compareResult(t, correct)
	subMan.clearPushList()

	sellDepth, buyDepth := hub.QueryDepth("abc/cet", 20)
	correct = `[{"p":"100.000000000000000000","a":"300"}]`
	bytes, _ = json.Marshal(sellDepth)
	require.Equal(t, correct, string(bytes))
	bytes, _ = json.Marshal(buyDepth)
	require.Equal(t, correct, string(bytes))

	newHeightInfo = &NewHeightInfo{
		Height:        1001,
		TimeStamp:     T("2019-07-15T08:19:10Z"),
		LastBlockHash: []byte("12345678901234567890"),
	}
	bytes, _ = json.Marshal(newHeightInfo)
	hub.ConsumeMessage("height_info", bytes)
	correct = `
3: {"height":1001,"timestamp":"2019-07-15T08:19:10Z","last_block_hash":"3132333435363738393031323334353637383930"}
4: {"height":1001,"timestamp":"2019-07-15T08:19:10Z","last_block_hash":"3132333435363738393031323334353637383930"}
`
	subMan.compareResult(t, correct)
	subMan.clearPushList()

	notificationCompleteRedelegation := &NotificationCompleteRedelegation{
		Delegator:    addr2,
		ValidatorSrc: "Val1",
		ValidatorDst: "Val2",
	}
	bytes, _ = json.Marshal(notificationCompleteRedelegation)
	hub.ConsumeMessage("complete_redelegation", bytes)

	notificationCompleteUnbonding := &NotificationCompleteUnbonding{
		Delegator: addr1,
		Validator: "Val1",
	}
	bytes, _ = json.Marshal(notificationCompleteUnbonding)
	hub.ConsumeMessage("complete_unbonding", bytes)
	correct = `
22: {"delegator":"cosmos1qy352eufqy352eufqy352eufqy35qqqz9ayrkz","src":"Val1","dst":"Val2","amount":"500","completion_time":"2019-07-15T08:18:10Z"}
21: {"delegator":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","validator":"Val1","amount":"300","completion_time":"2019-07-15T08:18:10Z"}
`
	subMan.compareResult(t, correct)
	subMan.clearPushList()

	commentRef := CommentRef{
		ID:           180,
		RewardTarget: addr1,
		RewardToken:  "cet",
		RewardAmount: 500000,
		Attitudes:    []int32{},
	}
	tokenComment := &TokenComment{
		ID:          181,
		Height:      1001,
		Sender:      addr2,
		Token:       "cet",
		Donation:    0,
		Title:       "I love CET",
		Content:     "I love CET so much.",
		ContentType: 3,
		References:  []CommentRef{commentRef},
	}
	bytes, _ = json.Marshal(tokenComment)
	hub.ConsumeMessage("token_comment", bytes)
	correct = `
13: {"id":181,"height":1001,"sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqz9ayrkz","token":"cet","donation":0,"title":"I love CET","content":"I love CET so much.","content_type":3,"references":[{"id":180,"reward_target":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","reward_token":"cet","reward_amount":500000,"attitudes":[]}]}
14: {"id":181,"height":1001,"sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqz9ayrkz","token":"cet","donation":0,"title":"I love CET","content":"I love CET so much.","content_type":3,"references":[{"id":180,"reward_target":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","reward_token":"cet","reward_amount":500000,"attitudes":[]}]}
`
	subMan.compareResult(t, correct)
	subMan.clearPushList()

	fillOrderInfo := &FillOrderInfo{
		OrderID:     addr1 + "-1",
		TradingPair: "abc/cet",
		Height:      1001,
		Side:        SELL,
		Price:       sdk.NewDec(100),
		LeftStock:   0,
		Freeze:      0,
		DealStock:   100,
		DealMoney:   10,
		CurrStock:   0,
		CurrMoney:   0,
	}
	bytes, _ = json.Marshal(fillOrderInfo)
	hub.ConsumeMessage("fill_order_info", bytes)
	correct = `
15: {"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-1","trading_pair":"abc/cet","height":1001,"side":2,"price":"100.000000000000000000","left_stock":0,"freeze":0,"deal_stock":100,"deal_money":10,"curr_stock":0,"curr_money":0}
16: {"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-1","trading_pair":"abc/cet","height":1001,"side":2,"price":"100.000000000000000000","left_stock":0,"freeze":0,"deal_stock":100,"deal_money":10,"curr_stock":0,"curr_money":0}
`
	subMan.compareResult(t, correct)
	subMan.clearPushList()

	sellDepth, buyDepth = hub.QueryDepth("abc/cet", 20)
	correct = `[{"p":"100.000000000000000000","a":"200"}]`
	bytes, _ = json.Marshal(sellDepth)
	require.Equal(t, correct, string(bytes))
	correct = `[{"p":"100.000000000000000000","a":"300"}]`
	bytes, _ = json.Marshal(buyDepth)
	require.Equal(t, correct, string(bytes))

	cancelOrderInfo := &CancelOrderInfo{
		OrderID:        addr1 + "-1",
		TradingPair:    "abc/cet",
		Height:         1001,
		Side:           BUY,
		Price:          sdk.NewDec(100),
		DelReason:      "Manually cancel the order",
		UsedCommission: 0,
		LeftStock:      50,
		RemainAmount:   0,
		DealStock:      100,
		DealMoney:      10,
	}
	bytes, _ = json.Marshal(cancelOrderInfo)
	hub.ConsumeMessage("del_order_info", bytes)
	correct = `
15: {"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-1","trading_pair":"abc/cet","height":1001,"side":1,"price":"100.000000000000000000","del_reason":"Manually cancel the order","used_commission":0,"left_stock":50,"remain_amount":0,"deal_stock":100,"deal_money":10}
16: {"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-1","trading_pair":"abc/cet","height":1001,"side":1,"price":"100.000000000000000000","del_reason":"Manually cancel the order","used_commission":0,"left_stock":50,"remain_amount":0,"deal_stock":100,"deal_money":10}
`
	subMan.compareResult(t, correct)
	subMan.clearPushList()

	hub.ConsumeMessage("commit", nil)
	correct = `
8: {"type":"depth","payload":{"trading_pair":"abc/cet","bids":null,"asks":[{"p":"100.000000000000000000","a":"200"}]}}
8: {"type":"depth","payload":{"trading_pair":"abc/cet","bids":[{"p":"100.000000000000000000","a":"250"}],"asks":null}}
9: {"type":"depth","payload":{"trading_pair":"abc/cet","bids":null,"asks":[{"p":"100.000000000000000000","a":"200"}]}}
9: {"type":"depth","payload":{"trading_pair":"abc/cet","bids":[{"p":"100.000000000000000000","a":"250"}],"asks":null}}
`
	subMan.compareResult(t, correct)
	subMan.clearPushList()

	newHeightInfo = &NewHeightInfo{
		Height:        1002,
		TimeStamp:     T("2019-07-15T08:29:10Z"),
		LastBlockHash: []byte("23456789012345678901"),
	}
	bytes, _ = json.Marshal(newHeightInfo)
	hub.ConsumeMessage("height_info", bytes)
	correct = `
3: {"height":1002,"timestamp":"2019-07-15T08:29:10Z","last_block_hash":"3233343536373839303132333435363738393031"}
4: {"height":1002,"timestamp":"2019-07-15T08:29:10Z","last_block_hash":"3233343536373839303132333435363738393031"}
6: {"open":"0.100000000000000000","close":"0.100000000000000000","high":"0.100000000000000000","low":"0.100000000000000000","total":"100","unix_time":1563178750,"time_span":16,"market":"abc/cet"}
`
	subMan.compareResult(t, correct)
	subMan.clearPushList()

	msgBancorInfoForKafka := &MsgBancorInfoForKafka{
		Owner:              addr1,
		Stock:              "xyz",
		Money:              "cet",
		InitPrice:          sdk.NewDec(10),
		MaxSupply:          sdk.NewInt(10000),
		MaxPrice:           sdk.NewDec(100),
		Price:              sdk.NewDec(20),
		StockInPool:        sdk.NewInt(50),
		MoneyInPool:        sdk.NewInt(5000),
		EarliestCancelTime: 0,
		BlockHeight:        1001,
	}
	bytes, _ = json.Marshal(msgBancorInfoForKafka)
	hub.ConsumeMessage("bancor_info", bytes)
	correct = `
12: {"sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","stock":"xyz","money":"cet","init_price":"10.000000000000000000","max_supply":"10000","max_price":"100.000000000000000000","price":"20.000000000000000000","stock_in_pool":"50","money_in_pool":"5000","earliest_cancel_time":0,"block_height":1001}
`
	subMan.compareResult(t, correct)
	subMan.clearPushList()

	msgBancorTradeInfoForKafka := MsgBancorTradeInfoForKafka{
		Sender:      addr2,
		Stock:       "xyz",
		Money:       "cet",
		Amount:      1,
		Side:        SELL,
		MoneyLimit:  10,
		TxPrice:     sdk.NewDec(2),
		BlockHeight: 1001,
	}
	bytes, _ = json.Marshal(msgBancorTradeInfoForKafka)
	hub.ConsumeMessage("bancor_trade", bytes)
	correct = `
17: {"sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqz9ayrkz","stock":"xyz","money":"cet","amount":1,"side":2,"money_limit":10,"transaction_price":"2.000000000000000000","block_height":1001}
18: {"sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqz9ayrkz","stock":"xyz","money":"cet","amount":1,"side":2,"money_limit":10,"transaction_price":"2.000000000000000000","block_height":1001}
`
	subMan.compareResult(t, correct)
	subMan.clearPushList()

	hub.ConsumeMessage("commit", nil)
	correct = `
0: [{"market":"abc/cet","new":"0.100000000000000000","old":"0.100000000000000000"}]
`
	subMan.compareResult(t, correct)
	subMan.clearPushList()

	tickers := hub.QueryTickers([]string{"abc/cet"})
	bytes, _ = json.Marshal(tickers)
	correct = `[{"market":"abc/cet","new":"0.100000000000000000","old":"0.100000000000000000"}]`
	require.Equal(t, correct, string(bytes))

	blockTimes := hub.QueryBlockTime(1100, 100)
	bytes, _ = json.Marshal(blockTimes)
	correct = `[1563179350,1563178750,1563178030]`
	require.Equal(t, correct, string(bytes))

	//subMan.showResult()
	correct = `{"open":"0.100000000000000000","close":"0.100000000000000000","high":"0.100000000000000000","low":"0.100000000000000000","total":"100","unix_time":1563178750,"time_span":16,"market":"abc/cet"}`
	unixTime := T("2019-07-15T08:39:10Z").Unix()
	data := hub.QueryCandleStick("abc/cet", Minute, unixTime, 0, 20)
	require.Equal(t, correct, toStr(data))

	data, tags, timesid := hub.QueryOrder(addr1, unixTime, 0, 20)
	correct = `{"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-1","trading_pair":"abc/cet","height":1001,"side":1,"price":"100.000000000000000000","del_reason":"Manually cancel the order","used_commission":0,"left_stock":50,"remain_amount":0,"deal_stock":100,"deal_money":10}
{"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-1","trading_pair":"abc/cet","height":1001,"side":2,"price":"100.000000000000000000","left_stock":0,"freeze":0,"deal_stock":100,"deal_money":10,"curr_stock":0,"curr_money":0}
{"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-2","sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","trading_pair":"abc/cet","order_type":2,"price":"100.000000000000000000","quantity":300,"side":1,"time_in_force":3,"feature_fee":1,"height":1001,"frozen_fee":1,"freeze":10}
{"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-1","sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","trading_pair":"abc/cet","order_type":2,"price":"100.000000000000000000","quantity":300,"side":2,"time_in_force":3,"feature_fee":1,"height":1001,"frozen_fee":1,"freeze":10}`
	require.Equal(t, correct, toStr(data))
	require.Equal(t, "dfcc", string(tags))
	bytes, _ = json.Marshal(timesid)
	require.Equal(t, "[1563178750,12,1563178750,10,1563178030,7,1563178030,6]", string(bytes))

	data, timesid = hub.QueryDeal("abc/cet", unixTime, 0, 20)
	correct = `{"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-1","trading_pair":"abc/cet","height":1001,"side":2,"price":"100.000000000000000000","left_stock":0,"freeze":0,"deal_stock":100,"deal_money":10,"curr_stock":0,"curr_money":0}`
	require.Equal(t, correct, toStr(data))
	bytes, _ = json.Marshal(timesid)
	require.Equal(t, "[1563178750,11]", string(bytes))

	data, timesid = hub.QueryBancorInfo("xyz/cet", unixTime, 0, 20)
	correct = `{"sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","stock":"xyz","money":"cet","init_price":"10.000000000000000000","max_supply":"10000","max_price":"100.000000000000000000","price":"20.000000000000000000","stock_in_pool":"50","money_in_pool":"5000","earliest_cancel_time":0,"block_height":1001}`
	require.Equal(t, correct, toStr(data))
	bytes, _ = json.Marshal(timesid)
	require.Equal(t, "[1563179350,14]", string(bytes))

	data, timesid = hub.QueryBancorTrade(addr2, unixTime, 0, 20)
	correct = `{"sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqz9ayrkz","stock":"xyz","money":"cet","amount":1,"side":2,"money_limit":10,"transaction_price":"2.000000000000000000","block_height":1001}`
	require.Equal(t, correct, toStr(data))
	bytes, _ = json.Marshal(timesid)
	require.Equal(t, "[1563179350,15]", string(bytes))

	data, timesid = hub.QueryRedelegation(addr2, unixTime, 0, 20)
	correct = `{"delegator":"cosmos1qy352eufqy352eufqy352eufqy35qqqz9ayrkz","src":"Val1","dst":"Val2","amount":"500","completion_time":"2019-07-15T08:18:10Z"}`
	require.Equal(t, correct, toStr(data))
	bytes, _ = json.Marshal(timesid)
	require.Equal(t, "[1563178690,4]", string(bytes))

	data, timesid = hub.QueryUnbonding(addr1, unixTime, 0, 20)
	correct = `{"delegator":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","validator":"Val1","amount":"300","completion_time":"2019-07-15T08:18:10Z"}`
	require.Equal(t, correct, toStr(data))
	bytes, _ = json.Marshal(timesid)
	require.Equal(t, "[1563178690,5]", string(bytes))

	data, timesid = hub.QueryUnlock(addr2, unixTime, 0, 20)
	correct = `{"address":"cosmos1qy352eufqy352eufqy352eufqy35qqqz9ayrkz","unlocked":[{"denom":"abc","amount":"15000"}],"locked_coins":[{"coin":{"denom":"cet","amount":"5000"},"unlock_time":1563178690}],"frozen_coins":[],"coins":[],"height":1001}`
	require.Equal(t, correct, toStr(data))
	bytes, _ = json.Marshal(timesid)
	require.Equal(t, "[1563178030,3]", string(bytes))

	data, timesid = hub.QueryIncome(addr2, unixTime, 0, 20)
	correct = `{"signers":["cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca"],"transfers":[{"sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","recipient":"cosmos1qy352eufqy352eufqy352eufqy35qqqz9ayrkz","amount":"1cet"}],"serial_number":20000,"msg_types":["MsgType1"],"tx_json":"blabla","height":1001}`
	require.Equal(t, correct, toStr(data))
	bytes, _ = json.Marshal(timesid)
	require.Equal(t, "[1563178030,2]", string(bytes))

	data, timesid = hub.QueryTx(addr1, unixTime, 0, 20)
	correct = `{"signers":["cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca"],"transfers":[{"sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","recipient":"cosmos1qy352eufqy352eufqy352eufqy35qqqz9ayrkz","amount":"1cet"}],"serial_number":20000,"msg_types":["MsgType1"],"tx_json":"blabla","height":1001}`
	require.Equal(t, correct, toStr(data))
	bytes, _ = json.Marshal(timesid)
	require.Equal(t, "[1563178030,1]", string(bytes))

	data, timesid = hub.QueryComment("cet", unixTime, 0, 20)
	correct = `{"id":181,"height":1001,"sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqz9ayrkz","token":"cet","donation":0,"title":"I love CET","content":"I love CET so much.","content_type":3,"references":[{"id":180,"reward_target":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","reward_token":"cet","reward_amount":500000,"attitudes":[]}]}`
	require.Equal(t, correct, toStr(data))
	bytes, _ = json.Marshal(timesid)
	require.Equal(t, "[1563178750,9]", string(bytes))

	data, timesid = hub.QuerySlash(unixTime, 0, 20)
	correct = `{"validator":"Val1","power":"30%","reason":"double_sign","jailed":true}`
	require.Equal(t, correct, toStr(data))
	bytes, _ = json.Marshal(timesid)
	require.Equal(t, "[1563178030,8]", string(bytes))

	newHeightInfo = &NewHeightInfo{
		Height:        1002,
		TimeStamp:     T("2019-07-15T08:31:10Z"),
		LastBlockHash: []byte("23456789012345678901"),
	}
	bytes, _ = json.Marshal(newHeightInfo)
	hub.ConsumeMessage("height_info", bytes)
	correct = `
3: {"height":1002,"timestamp":"2019-07-15T08:31:10Z","last_block_hash":"3233343536373839303132333435363738393031"}
4: {"height":1002,"timestamp":"2019-07-15T08:31:10Z","last_block_hash":"3233343536373839303132333435363738393031"}
6: {"open":"0.100000000000000000","close":"0.100000000000000000","high":"0.100000000000000000","low":"0.100000000000000000","total":"100","unix_time":1563178750,"time_span":16,"market":"abc/cet"}
`
	subMan.compareResult(t, correct)
	subMan.clearPushList()

	fillOrderInfo = &FillOrderInfo{
		OrderID:     addr1 + "-1",
		TradingPair: "abc/cet",
		Height:      1003,
		Side:        SELL,
		Price:       sdk.NewDec(100),
		LeftStock:   0,
		Freeze:      0,
		DealStock:   200,
		DealMoney:   25,
		CurrStock:   0,
		CurrMoney:   0,
	}
	bytes, _ = json.Marshal(fillOrderInfo)
	hub.ConsumeMessage("fill_order_info", bytes)
	correct = `
15: {"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-1","trading_pair":"abc/cet","height":1003,"side":2,"price":"100.000000000000000000","left_stock":0,"freeze":0,"deal_stock":200,"deal_money":25,"curr_stock":0,"curr_money":0}
16: {"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-1","trading_pair":"abc/cet","height":1003,"side":2,"price":"100.000000000000000000","left_stock":0,"freeze":0,"deal_stock":200,"deal_money":25,"curr_stock":0,"curr_money":0}
`
	subMan.compareResult(t, correct)
	subMan.clearPushList()

	hub.ConsumeMessage("commit", nil)
	correct = `
0: [{"market":"abc/cet","new":"0.100000000000000000","old":"0.100000000000000000"}]
8: {"type":"depth","payload":{"trading_pair":"abc/cet","bids":null,"asks":[{"p":"100.000000000000000000","a":"0"}]}}
9: {"type":"depth","payload":{"trading_pair":"abc/cet","bids":null,"asks":[{"p":"100.000000000000000000","a":"0"}]}}
`
	subMan.compareResult(t, correct)
	subMan.clearPushList()

	newHeightInfo = &NewHeightInfo{
		Height:        1003,
		TimeStamp:     T("2019-07-16T00:01:10Z"),
		LastBlockHash: []byte("23456789012345678901"),
	}
	bytes, _ = json.Marshal(newHeightInfo)
	hub.ConsumeMessage("height_info", bytes)
	correct = `
3: {"height":1003,"timestamp":"2019-07-16T00:01:10Z","last_block_hash":"3233343536373839303132333435363738393031"}
4: {"height":1003,"timestamp":"2019-07-16T00:01:10Z","last_block_hash":"3233343536373839303132333435363738393031"}
6: {"open":"0.125000000000000000","close":"0.125000000000000000","high":"0.125000000000000000","low":"0.125000000000000000","total":"200","unix_time":1563179470,"time_span":16,"market":"abc/cet"}
7: {"open":"0.100000000000000000","close":"0.125000000000000000","high":"0.125000000000000000","low":"0.100000000000000000","total":"300","unix_time":1563179470,"time_span":32,"market":"abc/cet"}
5: {"open":"0.100000000000000000","close":"0.125000000000000000","high":"0.125000000000000000","low":"0.100000000000000000","total":"300","unix_time":1563179470,"time_span":48,"market":"abc/cet"}
`
	subMan.compareResult(t, correct)
	subMan.clearPushList()

	fillOrderInfo = &FillOrderInfo{
		OrderID:     addr1 + "-1",
		TradingPair: "abc/cet",
		Height:      1003,
		Side:        SELL,
		Price:       sdk.NewDec(110),
		LeftStock:   0,
		Freeze:      0,
		DealStock:   200,
		DealMoney:   25,
		CurrStock:   0,
		CurrMoney:   0,
	}
	bytes, _ = json.Marshal(fillOrderInfo)
	hub.ConsumeMessage("fill_order_info", bytes)
	correct = `
15: {"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-1","trading_pair":"abc/cet","height":1003,"side":2,"price":"110.000000000000000000","left_stock":0,"freeze":0,"deal_stock":200,"deal_money":25,"curr_stock":0,"curr_money":0}
16: {"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-1","trading_pair":"abc/cet","height":1003,"side":2,"price":"110.000000000000000000","left_stock":0,"freeze":0,"deal_stock":200,"deal_money":25,"curr_stock":0,"curr_money":0}
`
	subMan.compareResult(t, correct)
	subMan.clearPushList()

	hub.ConsumeMessage("commit", nil)
	correct = `
0: [{"market":"abc/cet","new":"0.125000000000000000","old":"0.100000000000000000"}]
8: {"type":"depth","payload":{"trading_pair":"abc/cet","bids":null,"asks":[{"p":"110.000000000000000000","a":"-200"}]}}
9: {"type":"depth","payload":{"trading_pair":"abc/cet","bids":null,"asks":[{"p":"110.000000000000000000","a":"-200"}]}}
`
	subMan.compareResult(t, correct)
	subMan.clearPushList()

	unixTime = T("2019-07-25T08:39:10Z").Unix()
	data = hub.QueryCandleStick("abc/cet", Hour, unixTime, 0, 20)
	correct = `{"open":"0.100000000000000000","close":"0.125000000000000000","high":"0.125000000000000000","low":"0.100000000000000000","total":"300","unix_time":1563179470,"time_span":32,"market":"abc/cet"}`
	require.Equal(t, correct, toStr(data))

	data = hub.QueryCandleStick("abc/cet", Day, unixTime, 0, 20)
	correct = `{"open":"0.100000000000000000","close":"0.125000000000000000","high":"0.125000000000000000","low":"0.100000000000000000","total":"300","unix_time":1563179470,"time_span":48,"market":"abc/cet"}`
	require.Equal(t, correct, toStr(data))
}
