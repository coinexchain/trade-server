package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
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

func NewTickerSubscriber(id int64, markets []string) *TickerSubscriber {
	return &TickerSubscriber{
		PlainSubscriber: PlainSubscriber{ID: id},
		Markets:         markets,
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

func NewCandleStickSubscriber(id int64, timespan byte) *CandleStickSubscriber {
	return &CandleStickSubscriber{
		PlainSubscriber: PlainSubscriber{ID: id},
		TimeSpan:        timespan,
	}
}

type pushInfo struct {
	Target  Subscriber
	Payload string
}

type MocSubscribeManager struct {
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
	LockedSubcribeInfo        map[string][]Subscriber

	PushList []pushInfo
}

func (sm *MocSubscribeManager) showResult() {
	for _, info := range sm.PushList {
		var id int64
		switch v := info.Target.(type) {
		case *PlainSubscriber:
			id = v.ID
		case *TickerSubscriber:
			id = v.ID
		case *CandleStickSubscriber:
			id = v.ID
		}
		fmt.Printf("%d: %s\n", id, info.Payload)
	}
}

func (sm *MocSubscribeManager) clearPushList() {
	sm.PushList = sm.PushList[:0]
}

func (sm *MocSubscribeManager) compareResult(t *testing.T, correct string) {
	out := make([]string, 1, 10)
	out[0] = "\n"
	for _, info := range sm.PushList {
		var id int64
		switch v := info.Target.(type) {
		case *PlainSubscriber:
			id = v.ID
		case *TickerSubscriber:
			id = v.ID
		case *CandleStickSubscriber:
			id = v.ID
		}
		s := fmt.Sprintf("%d: %s\n", id, info.Payload)
		out = append(out, s)
	}
	require.Equal(t, correct, strings.Join(out, ""))
}

func GetSubscribeManager(addr1, addr2 string) *MocSubscribeManager {
	res := &MocSubscribeManager{}
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
	res.CandleStickSubscribeInfo["abc/cet"] = make([]Subscriber, 3)
	res.CandleStickSubscribeInfo["abc/cet"][0] = NewCandleStickSubscriber(6, Minute)
	res.CandleStickSubscribeInfo["abc/cet"][1] = NewCandleStickSubscriber(7, Hour)
	res.CandleStickSubscribeInfo["abc/cet"][2] = NewCandleStickSubscriber(5, Day)
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

func (sm *MocSubscribeManager) ClearPushInfoList() {
	sm.PushList = sm.PushList[:0]
}

func (sm *MocSubscribeManager) GetSlashSubscribeInfo() []Subscriber {
	return sm.SlashSubscribeInfo
}
func (sm *MocSubscribeManager) GetHeightSubscribeInfo() []Subscriber {
	return sm.HeightSubscribeInfo
}
func (sm *MocSubscribeManager) GetTickerSubscribeInfo() []Subscriber {
	return sm.TickerSubscribeInfo
}
func (sm *MocSubscribeManager) GetCandleStickSubscribeInfo() map[string][]Subscriber {
	return sm.CandleStickSubscribeInfo
}
func (sm *MocSubscribeManager) GetDepthSubscribeInfo() map[string][]Subscriber {
	return sm.DepthSubscribeInfo
}
func (sm *MocSubscribeManager) GetLockedSubscribeInfo() map[string][]Subscriber {
	return sm.LockedSubcribeInfo
}
func (sm *MocSubscribeManager) GetDealSubscribeInfo() map[string][]Subscriber {
	return sm.DealSubscribeInfo
}
func (sm *MocSubscribeManager) GetBancorInfoSubscribeInfo() map[string][]Subscriber {
	return sm.BancorInfoSubscribeInfo
}
func (sm *MocSubscribeManager) GetCommentSubscribeInfo() map[string][]Subscriber {
	return sm.CommentSubscribeInfo
}
func (sm *MocSubscribeManager) GetOrderSubscribeInfo() map[string][]Subscriber {
	return sm.OrderSubscribeInfo
}
func (sm *MocSubscribeManager) GetBancorTradeSubscribeInfo() map[string][]Subscriber {
	return sm.BancorTradeSubscribeInfo
}
func (sm *MocSubscribeManager) GetIncomeSubscribeInfo() map[string][]Subscriber {
	return sm.IncomeSubscribeInfo
}
func (sm *MocSubscribeManager) GetUnbondingSubscribeInfo() map[string][]Subscriber {
	return sm.UnbondingSubscribeInfo
}
func (sm *MocSubscribeManager) GetRedelegationSubscribeInfo() map[string][]Subscriber {
	return sm.RedelegationSubscribeInfo
}
func (sm *MocSubscribeManager) GetUnlockSubscribeInfo() map[string][]Subscriber {
	return sm.UnlockSubscribeInfo
}
func (sm *MocSubscribeManager) GetTxSubscribeInfo() map[string][]Subscriber {
	return sm.TxSubscribeInfo
}
func (sm *MocSubscribeManager) PushLockedSendMsg(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushSlash(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushHeight(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushTicker(subscriber Subscriber, t []*Ticker) {
	info, _ := json.Marshal(t)
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushDepthSell(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushDepthBuy(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushCandleStick(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushDeal(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushCreateOrder(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushFillOrder(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushCancelOrder(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushBancorInfo(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushBancorTrade(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushIncome(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushUnbonding(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushRedelegation(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushUnlock(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushTx(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushComment(subscriber Subscriber, info []byte) {
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
