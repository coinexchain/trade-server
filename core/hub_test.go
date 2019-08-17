package core

//import (
//	"testing"
//	"github.com/stretchr/testify/require"
//	sdk "github.com/cosmos/cosmos-sdk/types"
//)

type PlainSubscriber struct {
	ID int
}

func (s *PlainSubscriber) Detail() interface{} {
	return nil
}

func (s *PlainSubscriber) WriteMsg([]byte) error {
	return nil
}

type TickerSubscriber struct {
	ID      int
	Markets []string
}

func (s *TickerSubscriber) Detail() interface{} {
	return s.Markets
}

func (s *TickerSubscriber) WriteMsg([]byte) error {
	return nil
}

type CandleStickSubscriber struct {
	ID       int
	TimeSpan byte
}

func (s *CandleStickSubscriber) Detail() interface{} {
	return s.TimeSpan
}

func (s *CandleStickSubscriber) WriteMsg([]byte) error {
	return nil
}

type pushInfo struct {
	Target  Subscriber
	Payload interface{}
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

var addr1 = "addr1"
var addr2 = "addr2"

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
	res.TickerSubscribeInfo[0] = &TickerSubscriber{ID: 0, Markets: []string{"abc/cet", "xyz/cet"}}
	res.CandleStickSubscribeInfo = make(map[string][]Subscriber)
	res.CandleStickSubscribeInfo["abc/cet"] = make([]Subscriber, 2)
	res.CandleStickSubscribeInfo["abc/cet"][0] = &CandleStickSubscriber{ID: 5, TimeSpan: Minute}
	res.CandleStickSubscribeInfo["abc/cet"][1] = &CandleStickSubscriber{ID: 6, TimeSpan: Hour}
	res.DepthSubscribeInfo = make(map[string][]Subscriber)
	res.DepthSubscribeInfo["abc/cet"] = make([]Subscriber, 2)
	res.DepthSubscribeInfo["abc/cet"][0] = &PlainSubscriber{ID: 6}
	res.DepthSubscribeInfo["abc/cet"][1] = &PlainSubscriber{ID: 7}
	res.DepthSubscribeInfo["xyz/cet"] = make([]Subscriber, 1)
	res.DepthSubscribeInfo["xyz/cet"][0] = &PlainSubscriber{ID: 7}
	res.DealSubscribeInfo = make(map[string][]Subscriber)
	res.DealSubscribeInfo["xyz/cet"] = make([]Subscriber, 1)
	res.DealSubscribeInfo["xyz/cet"][0] = &PlainSubscriber{ID: 8}
	res.BancorInfoSubscribeInfo = make(map[string][]Subscriber)
	res.BancorInfoSubscribeInfo["xyz/cet"] = make([]Subscriber, 1)
	res.BancorInfoSubscribeInfo["xyz/cet"][0] = &PlainSubscriber{ID: 9}
	res.CommentSubscribeInfo = make(map[string][]Subscriber)
	res.CommentSubscribeInfo["cet"] = make([]Subscriber, 2)
	res.CommentSubscribeInfo["cet"][0] = &PlainSubscriber{ID: 10}
	res.CommentSubscribeInfo["cet"][1] = &PlainSubscriber{ID: 11}
	res.OrderSubscribeInfo = make(map[string][]Subscriber)
	res.OrderSubscribeInfo[addr1] = make([]Subscriber, 2)
	res.OrderSubscribeInfo[addr1][0] = &PlainSubscriber{ID: 12}
	res.OrderSubscribeInfo[addr1][1] = &PlainSubscriber{ID: 13}
	res.BancorTradeSubscribeInfo = make(map[string][]Subscriber)
	res.BancorTradeSubscribeInfo[addr2] = make([]Subscriber, 2)
	res.BancorTradeSubscribeInfo[addr2][0] = &PlainSubscriber{ID: 13}
	res.BancorTradeSubscribeInfo[addr2][1] = &PlainSubscriber{ID: 14}
	res.IncomeSubscribeInfo = make(map[string][]Subscriber)
	res.IncomeSubscribeInfo[addr1] = make([]Subscriber, 1)
	res.IncomeSubscribeInfo[addr1][0] = &PlainSubscriber{ID: 15}
	res.IncomeSubscribeInfo[addr2] = make([]Subscriber, 1)
	res.IncomeSubscribeInfo[addr2][0] = &PlainSubscriber{ID: 16}
	res.UnbondingSubscribeInfo = make(map[string][]Subscriber)
	res.UnbondingSubscribeInfo[addr1] = make([]Subscriber, 1)
	res.UnbondingSubscribeInfo[addr1][0] = &PlainSubscriber{ID: 17}
	res.RedelegationSubscribeInfo = make(map[string][]Subscriber)
	res.RedelegationSubscribeInfo[addr2] = make([]Subscriber, 1)
	res.RedelegationSubscribeInfo[addr2][0] = &PlainSubscriber{ID: 18}
	res.UnlockSubscribeInfo = make(map[string][]Subscriber)
	res.UnlockSubscribeInfo[addr2] = make([]Subscriber, 2)
	res.UnlockSubscribeInfo[addr2][0] = &PlainSubscriber{ID: 20}
	res.UnlockSubscribeInfo[addr2][1] = &PlainSubscriber{ID: 19}
	res.TxSubscribeInfo = make(map[string][]Subscriber)
	res.TxSubscribeInfo[addr1] = make([]Subscriber, 1)
	res.TxSubscribeInfo[addr1][0] = &PlainSubscriber{ID: 21}
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
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: info})
}
func (sm *mocSubscribeManager) PushHeight(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: info})
}
func (sm *mocSubscribeManager) PushTicker(subscriber Subscriber, t []*Ticker) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: t})
}
func (sm *mocSubscribeManager) PushDepthSell(subscriber Subscriber, delta map[*PricePoint]bool) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: delta})
}
func (sm *mocSubscribeManager) PushDepthBuy(subscriber Subscriber, delta map[*PricePoint]bool) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: delta})
}
func (sm *mocSubscribeManager) PushCandleStick(subscriber Subscriber, cs *CandleStick) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: cs})
}
func (sm *mocSubscribeManager) PushDeal(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: info})
}
func (sm *mocSubscribeManager) PushCreateOrder(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: info})
}
func (sm *mocSubscribeManager) PushFillOrder(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: info})
}
func (sm *mocSubscribeManager) PushCancelOrder(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: info})
}
func (sm *mocSubscribeManager) PushBancorInfo(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: info})
}
func (sm *mocSubscribeManager) PushBancorTrade(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: info})
}
func (sm *mocSubscribeManager) PushIncome(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: info})
}
func (sm *mocSubscribeManager) PushUnbonding(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: info})
}
func (sm *mocSubscribeManager) PushRedelegation(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: info})
}
func (sm *mocSubscribeManager) PushUnlock(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: info})
}
func (sm *mocSubscribeManager) PushTx(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: info})
}
func (sm *mocSubscribeManager) PushComment(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: info})
}
