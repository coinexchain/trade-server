package core

import (
	"fmt"
	"testing"
	"encoding/json"
	"strings"
	"github.com/stretchr/testify/require"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dbm "github.com/tendermint/tm-db"
)

type PlainSubscriber struct {
	ID int64
}

func (s *PlainSubscriber) Detail() interface{} {
	return nil
}

func (s *PlainSubscriber) WriteMsg([]byte) error {
	return nil
}

type TickerSubscriber struct {
	PlainSubscriber
	Markets []string
}

func (s *TickerSubscriber) Detail() interface{} {
	return s.Markets
}

func (s *TickerSubscriber) WriteMsg([]byte) error {
	return nil
}

func NewTickerSubscriber(id int64, markets []string)  *TickerSubscriber {
	return &TickerSubscriber{
		PlainSubscriber: PlainSubscriber{ID:id},
		Markets: markets,
	}
}

type CandleStickSubscriber struct {
	PlainSubscriber
	TimeSpan byte
}

func (s *CandleStickSubscriber) Detail() interface{} {
	return s.TimeSpan
}

func (s *CandleStickSubscriber) WriteMsg([]byte) error {
	return nil
}

func NewCandleStickSubscriber(id int64, timespan byte)  *CandleStickSubscriber {
	return &CandleStickSubscriber{
		PlainSubscriber: PlainSubscriber{ID:id},
		TimeSpan: timespan,
	}
}

type pushInfo struct {
	Target  Subscriber
	Payload string
}

type mocSubscribeManager struct {
	SlashSubscribeInfo        []Subscriber
	HeightSubscribeInfo       []Subscriber
	TickerSubscribeInfo       []Subscriber
	CandleStickSubscribeInfo  map[string][]Subscriber
	DepthSubscribeInfo        map[string][]Subscriber
	DealSubscribeInfo         map[string][]Subscriber
	BancorInfoSubscribeInfo   map[string][]Subscriber
	CommentSubscribeInfo      map[string][]Subscriber
	OrderSubscribeInfo        map[string][]Subscriber
	BancorTradeSubscribeInfo  map[string][]Subscriber
	IncomeSubscribeInfo       map[string][]Subscriber
	UnbondingSubscribeInfo    map[string][]Subscriber
	RedelegationSubscribeInfo map[string][]Subscriber
	UnlockSubscribeInfo       map[string][]Subscriber
	TxSubscribeInfo           map[string][]Subscriber

	PushList []pushInfo
}

func (subMan *mocSubscribeManager) showResult() {
	for _, info := range subMan.PushList {
		var id int64
		switch v := info.Target.(type) {
			case *PlainSubscriber: id=v.ID
			case *TickerSubscriber: id=v.ID
			case *CandleStickSubscriber: id=v.ID
		}
		fmt.Printf("%d: %s\n", id, info.Payload)
	}
}

func (subMan *mocSubscribeManager) clearPushList() {
	subMan.PushList = subMan.PushList[:0]
}

func (subMan *mocSubscribeManager) compareResult(t *testing.T, correct string) {
	out := make([]string,1,10)
	out[0] = "\n"
	for _, info := range subMan.PushList {
		var id int64
		switch v := info.Target.(type) {
			case *PlainSubscriber: id=v.ID
			case *TickerSubscriber: id=v.ID
			case *CandleStickSubscriber: id=v.ID
		}
		s := fmt.Sprintf("%d: %s\n", id, info.Payload)
		out=append(out,s)
	}
	require.Equal(t, correct, strings.Join(out,""))
}

var addr1 string
var addr2 string

func getSubscribeManager() *mocSubscribeManager {
	res := &mocSubscribeManager{}
	res.PushList = make([]pushInfo, 0, 100)
	res.SlashSubscribeInfo = make([]Subscriber, 2)
	res.SlashSubscribeInfo[0] = &PlainSubscriber{ID: 0}
	res.SlashSubscribeInfo[1] = &PlainSubscriber{ID: 1}
	res.HeightSubscribeInfo = make([]Subscriber, 2)
	res.HeightSubscribeInfo[0] = &PlainSubscriber{ID: 3}
	res.HeightSubscribeInfo[1] = &PlainSubscriber{ID: 4}
	res.TickerSubscribeInfo = make([]Subscriber, 1)
	res.TickerSubscribeInfo[0] = NewTickerSubscriber(0, []string{"abc/cet", "xyz/cet"})
	res.CandleStickSubscribeInfo = make(map[string][]Subscriber)
	res.CandleStickSubscribeInfo["abc/cet"] = make([]Subscriber, 2)
	res.CandleStickSubscribeInfo["abc/cet"][0] = NewCandleStickSubscriber(6, Minute)
	res.CandleStickSubscribeInfo["abc/cet"][1] = NewCandleStickSubscriber(7, Hour)
	res.DepthSubscribeInfo = make(map[string][]Subscriber)
	res.DepthSubscribeInfo["abc/cet"] = make([]Subscriber, 2)
	res.DepthSubscribeInfo["abc/cet"][0] = &PlainSubscriber{ID: 8}
	res.DepthSubscribeInfo["abc/cet"][1] = &PlainSubscriber{ID: 9}
	res.DepthSubscribeInfo["xyz/cet"] = make([]Subscriber, 1)
	res.DepthSubscribeInfo["xyz/cet"][0] = &PlainSubscriber{ID: 10}
	res.DealSubscribeInfo = make(map[string][]Subscriber)
	res.DealSubscribeInfo["xyz/cet"] = make([]Subscriber, 1)
	res.DealSubscribeInfo["xyz/cet"][0] = &PlainSubscriber{ID: 11}
	res.BancorInfoSubscribeInfo = make(map[string][]Subscriber)
	res.BancorInfoSubscribeInfo["xyz/cet"] = make([]Subscriber, 1)
	res.BancorInfoSubscribeInfo["xyz/cet"][0] = &PlainSubscriber{ID: 12}
	res.CommentSubscribeInfo = make(map[string][]Subscriber)
	res.CommentSubscribeInfo["cet"] = make([]Subscriber, 2)
	res.CommentSubscribeInfo["cet"][0] = &PlainSubscriber{ID: 13}
	res.CommentSubscribeInfo["cet"][1] = &PlainSubscriber{ID: 14}
	res.OrderSubscribeInfo = make(map[string][]Subscriber)
	res.OrderSubscribeInfo[addr1] = make([]Subscriber, 2)
	res.OrderSubscribeInfo[addr1][0] = &PlainSubscriber{ID: 15}
	res.OrderSubscribeInfo[addr1][1] = &PlainSubscriber{ID: 16}
	res.BancorTradeSubscribeInfo = make(map[string][]Subscriber)
	res.BancorTradeSubscribeInfo[addr2] = make([]Subscriber, 2)
	res.BancorTradeSubscribeInfo[addr2][0] = &PlainSubscriber{ID: 17}
	res.BancorTradeSubscribeInfo[addr2][1] = &PlainSubscriber{ID: 18}
	res.IncomeSubscribeInfo = make(map[string][]Subscriber)
	res.IncomeSubscribeInfo[addr1] = make([]Subscriber, 1)
	res.IncomeSubscribeInfo[addr1][0] = &PlainSubscriber{ID: 19}
	res.IncomeSubscribeInfo[addr2] = make([]Subscriber, 1)
	res.IncomeSubscribeInfo[addr2][0] = &PlainSubscriber{ID: 20}
	res.UnbondingSubscribeInfo = make(map[string][]Subscriber)
	res.UnbondingSubscribeInfo[addr1] = make([]Subscriber, 1)
	res.UnbondingSubscribeInfo[addr1][0] = &PlainSubscriber{ID: 21}
	res.RedelegationSubscribeInfo = make(map[string][]Subscriber)
	res.RedelegationSubscribeInfo[addr2] = make([]Subscriber, 1)
	res.RedelegationSubscribeInfo[addr2][0] = &PlainSubscriber{ID: 22}
	res.UnlockSubscribeInfo = make(map[string][]Subscriber)
	res.UnlockSubscribeInfo[addr2] = make([]Subscriber, 2)
	res.UnlockSubscribeInfo[addr2][0] = &PlainSubscriber{ID: 23}
	res.UnlockSubscribeInfo[addr2][1] = &PlainSubscriber{ID: 24}
	res.TxSubscribeInfo = make(map[string][]Subscriber)
	res.TxSubscribeInfo[addr1] = make([]Subscriber, 1)
	res.TxSubscribeInfo[addr1][0] = &PlainSubscriber{ID: 25}
	return res
}

func (sm *mocSubscribeManager) ClearPushInfoList() {
	sm.PushList = sm.PushList[:0]
}

func (sm *mocSubscribeManager) GetSlashSubscribeInfo() []Subscriber {
	return sm.SlashSubscribeInfo
}
func (sm *mocSubscribeManager) GetHeightSubscribeInfo() []Subscriber {
	return sm.HeightSubscribeInfo
}
func (sm *mocSubscribeManager) GetTickerSubscribeInfo() []Subscriber {
	return sm.TickerSubscribeInfo
}
func (sm *mocSubscribeManager) GetCandleStickSubscribeInfo() map[string][]Subscriber {
	return sm.CandleStickSubscribeInfo
}
func (sm *mocSubscribeManager) GetDepthSubscribeInfo() map[string][]Subscriber {
	return sm.DepthSubscribeInfo
}
func (sm *mocSubscribeManager) GetDealSubscribeInfo() map[string][]Subscriber {
	return sm.DealSubscribeInfo
}
func (sm *mocSubscribeManager) GetBancorInfoSubscribeInfo() map[string][]Subscriber {
	return sm.BancorInfoSubscribeInfo
}
func (sm *mocSubscribeManager) GetCommentSubscribeInfo() map[string][]Subscriber {
	return sm.CommentSubscribeInfo
}
func (sm *mocSubscribeManager) GetOrderSubscribeInfo() map[string][]Subscriber {
	return sm.OrderSubscribeInfo
}
func (sm *mocSubscribeManager) GetBancorTradeSubscribeInfo() map[string][]Subscriber {
	return sm.BancorTradeSubscribeInfo
}
func (sm *mocSubscribeManager) GetIncomeSubscribeInfo() map[string][]Subscriber {
	return sm.IncomeSubscribeInfo
}
func (sm *mocSubscribeManager) GetUnbondingSubscribeInfo() map[string][]Subscriber {
	return sm.UnbondingSubscribeInfo
}
func (sm *mocSubscribeManager) GetRedelegationSubscribeInfo() map[string][]Subscriber {
	return sm.RedelegationSubscribeInfo
}
func (sm *mocSubscribeManager) GetUnlockSubscribeInfo() map[string][]Subscriber {
	return sm.UnlockSubscribeInfo
}
func (sm *mocSubscribeManager) GetTxSubscribeInfo() map[string][]Subscriber {
	return sm.TxSubscribeInfo
}
func (sm *mocSubscribeManager) PushSlash(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *mocSubscribeManager) PushHeight(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *mocSubscribeManager) PushTicker(subscriber Subscriber, t []*Ticker) {
	info, _ := json.Marshal(t)
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *mocSubscribeManager) PushDepthSell(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *mocSubscribeManager) PushDepthBuy(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *mocSubscribeManager) PushCandleStick(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *mocSubscribeManager) PushDeal(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *mocSubscribeManager) PushCreateOrder(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *mocSubscribeManager) PushFillOrder(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *mocSubscribeManager) PushCancelOrder(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *mocSubscribeManager) PushBancorInfo(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *mocSubscribeManager) PushBancorTrade(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *mocSubscribeManager) PushIncome(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *mocSubscribeManager) PushUnbonding(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *mocSubscribeManager) PushRedelegation(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *mocSubscribeManager) PushUnlock(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *mocSubscribeManager) PushTx(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *mocSubscribeManager) PushComment(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}

func simpleAddr(s string) (sdk.AccAddress, error) {
	return sdk.AccAddressFromHex("01234567890123456789012345678901234" + s)
}

func toStr(payload [][]byte) string {
	out := make([]string, len(payload))
	for i:=0; i<len(out); i++ {
		out[i]=string(payload[i])
	}
	return strings.Join(out,"\n")
}

func Test1(t *testing.T) {
	acc1, _ := simpleAddr("00001")
	acc2, _ := simpleAddr("00002")
	addr1 = acc1.String()
	addr2 = acc2.String()

	db := dbm.NewMemDB()
	subMan := getSubscribeManager()
	hub := NewHub(db, subMan)

	newHeightInfo := &NewHeightInfo {
		Height:  1000,
		TimeStamp:  T("2019-07-15T08:07:10Z"),
		LastBlockHash: []byte("01234567890123456789"),
	}
	bytes, _ := json.Marshal(newHeightInfo)
	hub.ConsumeMessage("height_info", bytes)

	correct := `
3: {"height":1000,"timestamp":"2019-07-15T08:07:10Z","last_block_hash":"3031323334353637383930313233343536373839"}
4: {"height":1000,"timestamp":"2019-07-15T08:07:10Z","last_block_hash":"3031323334353637383930313233343536373839"}
`
	subMan.compareResult(t,correct)
	subMan.clearPushList()

	notificationSlash := &NotificationSlash{
		Validator: "Val1",
		Power: "30%",
		Reason: "double_sign",
		Jailed: true,
	}
	bytes, _ = json.Marshal(notificationSlash)
	hub.ConsumeMessage("notify_slash", bytes)

	transRec := TransferRecord{
		Sender: addr1,
		Recipient: addr2,
		Amount: "1cet",
	}
	notificationTx := &NotificationTx{
		Signers:  []string{addr1},
		Transfers:   []TransferRecord{transRec},
		SerialNumber: 20000,
		MsgTypes: []string{"MsgType1"},
		TxJSON: "blabla",
		Height: 1001,
	}
	bytes, _ = json.Marshal(notificationTx)
	hub.ConsumeMessage("notify_tx", bytes)
	correct = `
25: {"signers":["cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca"],"transfers":[{"sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","recipient":"cosmos1qy352eufqy352eufqy352eufqy35qqqz9ayrkz","amount":"1cet"}],"serial_number":20000,"msg_types":["MsgType1"],"tx_json":"blabla","height":1001}
20: {"signers":["cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca"],"transfers":[{"sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","recipient":"cosmos1qy352eufqy352eufqy352eufqy35qqqz9ayrkz","amount":"1cet"}],"serial_number":20000,"msg_types":["MsgType1"],"tx_json":"blabla","height":1001}
`
	subMan.compareResult(t,correct)
	subMan.clearPushList()

	lockedCoins := LockedCoins{
		{
			Coin: sdk.Coin{Denom:"cet", Amount:sdk.NewInt(5000)},
			UnlockTime: T("2019-07-15T08:18:10Z").Unix(),
		},
	}
	notificationUnlock := &NotificationUnlock {
		Address: addr2,
		Unlocked:    sdk.Coins {
			{Denom:"abc", Amount:sdk.NewInt(15000)},
		},
		LockedCoins:lockedCoins,
		FrozenCoins: sdk.Coins{},
		Coins:       sdk.Coins{},
		Height: 1001,
	}
	bytes, _ = json.Marshal(notificationUnlock)
	hub.ConsumeMessage("notify_unlock", bytes)
	correct = `
23: {"address":"cosmos1qy352eufqy352eufqy352eufqy35qqqz9ayrkz","unlocked":[{"denom":"abc","amount":"15000"}],"locked_coins":[{"coin":{"denom":"cet","amount":"5000"},"unlock_time":1563178690}],"frozen_coins":[],"coins":[],"height":1001}
24: {"address":"cosmos1qy352eufqy352eufqy352eufqy35qqqz9ayrkz","unlocked":[{"denom":"abc","amount":"15000"}],"locked_coins":[{"coin":{"denom":"cet","amount":"5000"},"unlock_time":1563178690}],"frozen_coins":[],"coins":[],"height":1001}
`
	subMan.compareResult(t,correct)
	subMan.clearPushList()

	notificationBeginRedelegation := &NotificationBeginRedelegation {
		Delegator: addr2,
		ValidatorSrc: "Val1",
		ValidatorDst: "Val2",
		Amount: "500",
		CompletionTime: "2019-07-15T08:18:10Z",
	}
	bytes, _ = json.Marshal(notificationBeginRedelegation)
	hub.ConsumeMessage("begin_redelegation", bytes)

	notificationBeginUnbonding := &NotificationBeginUnbonding {
		Delegator: addr1,
		Validator: "Val1",
		Amount: "300",
		CompletionTime: "2019-07-15T08:18:10Z",
	}
	bytes, _ = json.Marshal(notificationBeginUnbonding)
	hub.ConsumeMessage("begin_unbonding", bytes)

	createOrderInfo := &CreateOrderInfo{
		OrderID:    addr1+"-1",
		Sender:     addr1,
		TradingPair:"abc/cet",
		OrderType:  LIMIT,
		Price:      sdk.NewDec(100),
		Quantity:   300,
		Side:       SELL,
		TimeInForce:GTE,
		FeatureFee: 1,
		Height:     1001,
		FrozenFee:  1,
		Freeze:     10,
	}
	bytes, _ = json.Marshal(createOrderInfo)
	hub.ConsumeMessage("create_order_info", bytes)

	createOrderInfo = &CreateOrderInfo{
		OrderID:    addr1+"-2",
		Sender:     addr1,
		TradingPair:"abc/cet",
		OrderType:  LIMIT,
		Price:      sdk.NewDec(100),
		Quantity:   300,
		Side:       BUY,
		TimeInForce:GTE,
		FeatureFee: 1,
		Height:     1001,
		FrozenFee:  1,
		Freeze:     10,
	}
	bytes, _ = json.Marshal(createOrderInfo)
	hub.ConsumeMessage("create_order_info", bytes)
	correct = `
15: {"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-1","sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","trading_pair":"abc/cet","order_type":2,"price":"100.000000000000000000","quantity":300,"side":2,"time_in_force":3,"feature_fee":1,"height":1001,"frozen_fee":1,"freeze":10}
16: {"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-1","sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","trading_pair":"abc/cet","order_type":2,"price":"100.000000000000000000","quantity":300,"side":2,"time_in_force":3,"feature_fee":1,"height":1001,"frozen_fee":1,"freeze":10}
15: {"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-2","sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","trading_pair":"abc/cet","order_type":2,"price":"100.000000000000000000","quantity":300,"side":1,"time_in_force":3,"feature_fee":1,"height":1001,"frozen_fee":1,"freeze":10}
16: {"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-2","sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","trading_pair":"abc/cet","order_type":2,"price":"100.000000000000000000","quantity":300,"side":1,"time_in_force":3,"feature_fee":1,"height":1001,"frozen_fee":1,"freeze":10}
`
	subMan.compareResult(t,correct)
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
	subMan.compareResult(t,correct)
	subMan.clearPushList()

	sellDepth, buyDepth := hub.QueryDepth("abc/cet", 20)
	correct = `[{"p":"100.000000000000000000","a":"300"}]`
	bytes, _ = json.Marshal(sellDepth)
	require.Equal(t, correct, string(bytes))
	bytes, _ = json.Marshal(buyDepth)
	require.Equal(t, correct, string(bytes))

	newHeightInfo = &NewHeightInfo {
		Height:  1001,
		TimeStamp:  T("2019-07-15T08:19:10Z"),
		LastBlockHash: []byte("12345678901234567890"),
	}
	bytes, _ = json.Marshal(newHeightInfo)
	hub.ConsumeMessage("height_info", bytes)
	correct = `
3: {"height":1001,"timestamp":"2019-07-15T08:19:10Z","last_block_hash":"3132333435363738393031323334353637383930"}
4: {"height":1001,"timestamp":"2019-07-15T08:19:10Z","last_block_hash":"3132333435363738393031323334353637383930"}
`
	subMan.compareResult(t,correct)
	subMan.clearPushList()

	notificationCompleteRedelegation := &NotificationCompleteRedelegation {
		Delegator: addr2,
		ValidatorSrc: "Val1",
		ValidatorDst: "Val2",
	}
	bytes, _ = json.Marshal(notificationCompleteRedelegation)
	hub.ConsumeMessage("complete_redelegation", bytes)

	notificationCompleteUnbonding := &NotificationCompleteUnbonding {
		Delegator: addr1,
		Validator: "Val1",
	}
	bytes, _ = json.Marshal(notificationCompleteUnbonding)
	hub.ConsumeMessage("complete_unbonding", bytes)
	correct = `
22: {"delegator":"cosmos1qy352eufqy352eufqy352eufqy35qqqz9ayrkz","src":"Val1","dst":"Val2","amount":"500","completion_time":"2019-07-15T08:18:10Z"}
21: {"delegator":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","validator":"Val1","amount":"300","completion_time":"2019-07-15T08:18:10Z"}
`
	subMan.compareResult(t,correct)
	subMan.clearPushList()

	commentRef := CommentRef {
		ID: 180,
		RewardTarget: addr1,
		RewardToken: "cet",
		RewardAmount: 500000,
		Attitudes:    []int32{},
	}
	tokenComment := &TokenComment {
		ID:         181,
		Height:     1001,
		Sender:     addr2,
		Token:      "cet",
		Donation:   0,
		Title:      "I love CET",
		Content:    "I love CET so much.",
		ContentType: 3,
		References: []CommentRef{commentRef},
	}
	bytes, _ = json.Marshal(tokenComment)
	hub.ConsumeMessage("token_comment", bytes)
	correct = `
13: {"id":181,"height":1001,"sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqz9ayrkz","token":"cet","donation":0,"title":"I love CET","content":"I love CET so much.","content_type":3,"references":[{"id":180,"reward_target":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","reward_token":"cet","reward_amount":500000,"attitudes":[]}]}
14: {"id":181,"height":1001,"sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqz9ayrkz","token":"cet","donation":0,"title":"I love CET","content":"I love CET so much.","content_type":3,"references":[{"id":180,"reward_target":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","reward_token":"cet","reward_amount":500000,"attitudes":[]}]}
`
	subMan.compareResult(t,correct)
	subMan.clearPushList()

	fillOrderInfo := &FillOrderInfo {
		OrderID:    addr1+"-1",
		TradingPair:"abc/cet",
		Height:   1001,
		Side:     SELL,
		Price:    sdk.NewDec(100),
		LeftStock:0,
		Freeze:   0,
		DealStock:100,
		DealMoney:10,
		CurrStock:0,
		CurrMoney:0,
	}
	bytes, _ = json.Marshal(fillOrderInfo)
	hub.ConsumeMessage("fill_order_info", bytes)
	correct = `
15: {"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-1","trading_pair":"abc/cet","height":1001,"side":2,"price":"100.000000000000000000","left_stock":0,"freeze":0,"deal_stock":100,"deal_money":10,"curr_stock":0,"curr_money":0}
16: {"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-1","trading_pair":"abc/cet","height":1001,"side":2,"price":"100.000000000000000000","left_stock":0,"freeze":0,"deal_stock":100,"deal_money":10,"curr_stock":0,"curr_money":0}
`
	subMan.compareResult(t,correct)
	subMan.clearPushList()

	sellDepth, buyDepth = hub.QueryDepth("abc/cet", 20)
	correct = `[{"p":"100.000000000000000000","a":"200"}]`
	bytes, _ = json.Marshal(sellDepth)
	require.Equal(t, correct, string(bytes))
	correct = `[{"p":"100.000000000000000000","a":"300"}]`
	bytes, _ = json.Marshal(buyDepth)
	require.Equal(t, correct, string(bytes))

	cancelOrderInfo := &CancelOrderInfo{
		OrderID:    addr1+"-1",
		TradingPair:"abc/cet",
		Height:   1001,
		Side:     BUY,
		Price:    sdk.NewDec(100),
		DelReason:"Manually cancel the order",
		UsedCommission:0,
		LeftStock:     50,
		RemainAmount:  0,
		DealStock:     100,
		DealMoney:     10,
	}
	bytes, _ = json.Marshal(cancelOrderInfo)
	hub.ConsumeMessage("del_order_info", bytes)
	correct = `
15: {"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-1","trading_pair":"abc/cet","height":1001,"side":1,"price":"100.000000000000000000","del_reason":"Manually cancel the order","used_commission":0,"left_stock":50,"remain_amount":0,"deal_stock":100,"deal_money":10}
16: {"order_id":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca-1","trading_pair":"abc/cet","height":1001,"side":1,"price":"100.000000000000000000","del_reason":"Manually cancel the order","used_commission":0,"left_stock":50,"remain_amount":0,"deal_stock":100,"deal_money":10}
`
	subMan.compareResult(t,correct)
	subMan.clearPushList()

	hub.ConsumeMessage("commit", nil)
	correct = `
8: {"type":"depth","payload":{"trading_pair":"abc/cet","bids":null,"asks":[{"p":"100.000000000000000000","a":"200"}]}}
8: {"type":"depth","payload":{"trading_pair":"abc/cet","bids":[{"p":"100.000000000000000000","a":"250"}],"asks":null}}
9: {"type":"depth","payload":{"trading_pair":"abc/cet","bids":null,"asks":[{"p":"100.000000000000000000","a":"200"}]}}
9: {"type":"depth","payload":{"trading_pair":"abc/cet","bids":[{"p":"100.000000000000000000","a":"250"}],"asks":null}}
`
	subMan.compareResult(t,correct)
	subMan.clearPushList()

	newHeightInfo = &NewHeightInfo{
		Height:  1002,
		TimeStamp:  T("2019-07-15T08:29:10Z"),
		LastBlockHash: []byte("23456789012345678901"),
	}
	bytes, _ = json.Marshal(newHeightInfo)
	hub.ConsumeMessage("height_info", bytes)
	correct = `
3: {"height":1002,"timestamp":"2019-07-15T08:29:10Z","last_block_hash":"3233343536373839303132333435363738393031"}
4: {"height":1002,"timestamp":"2019-07-15T08:29:10Z","last_block_hash":"3233343536373839303132333435363738393031"}
6: {"open":"0.100000000000000000","close":"0.100000000000000000","high":"0.100000000000000000","low":"0.100000000000000000","total":"100","unix_time":1563178750,"time_span":16,"market":"abc/cet"}
`
	subMan.compareResult(t,correct)
	subMan.clearPushList()

	msgBancorInfoForKafka := &MsgBancorInfoForKafka{
		Owner:             addr1,
		Stock:             "xyz",
		Money:             "cet",
		InitPrice:         sdk.NewDec(10),
		MaxSupply:         sdk.NewInt(10000),
		MaxPrice:          sdk.NewDec(100),
		Price:             sdk.NewDec(20),
		StockInPool:       sdk.NewInt(50),
		MoneyInPool:       sdk.NewInt(5000),
		EarliestCancelTime:0,
		BlockHeight:       1001,
	}
	bytes, _ = json.Marshal(msgBancorInfoForKafka)
	hub.ConsumeMessage("bancor_info", bytes)
	correct = `
12: {"sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqptw34ca","stock":"xyz","money":"cet","init_price":"10.000000000000000000","max_supply":"10000","max_price":"100.000000000000000000","price":"20.000000000000000000","stock_in_pool":"50","money_in_pool":"5000","earliest_cancel_time":0,"block_height":1001}
`
	subMan.compareResult(t,correct)
	subMan.clearPushList()

	msgBancorTradeInfoForKafka := MsgBancorTradeInfoForKafka{
		Sender:     addr2,
		Stock:      "xyz",
		Money:      "cet",
		Amount:     1,
		Side:       SELL,
		MoneyLimit: 10,
		TxPrice:    sdk.NewDec(2),
		BlockHeight:1001,
	}
	bytes, _ = json.Marshal(msgBancorTradeInfoForKafka)
	hub.ConsumeMessage("bancor_trade", bytes)
	correct = `
17: {"sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqz9ayrkz","stock":"xyz","money":"cet","amount":1,"side":2,"money_limit":10,"transaction_price":"2.000000000000000000","block_height":1001}
18: {"sender":"cosmos1qy352eufqy352eufqy352eufqy35qqqz9ayrkz","stock":"xyz","money":"cet","amount":1,"side":2,"money_limit":10,"transaction_price":"2.000000000000000000","block_height":1001}
`
	subMan.compareResult(t,correct)
	subMan.clearPushList()

	hub.ConsumeMessage("commit", nil)
	correct = `
0: [{"market":"abc/cet","new":"0.100000000000000000","old":"0.100000000000000000"}]
`
	subMan.compareResult(t,correct)
	subMan.clearPushList()

	tickers := hub.QueryTickers([]string{"abc/cet"})
	bytes, _ = json.Marshal(tickers)
	correct = `[{"market":"abc/cet","new":"0.100000000000000000","old":"0.100000000000000000"}]`
	require.Equal(t, correct, string(bytes))

	blockTimes := hub.QueryBlockTime(1100,100)
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
}


