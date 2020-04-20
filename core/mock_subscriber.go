package core

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
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

func (s *PlainSubscriber) GetConn() *Conn {
	return nil
}

type TickerSubscriber struct {
	PlainSubscriber
	Markets map[string]struct{}
}

func (s *TickerSubscriber) Detail() interface{} {
	return s.Markets
}

func (s *TickerSubscriber) WriteMsg([]byte) error {
	return nil
}

func (s *TickerSubscriber) GetConn() *Conn {
	return nil
}

func NewTickerSubscriber(id int64, markets map[string]struct{}) *TickerSubscriber {
	return &TickerSubscriber{
		PlainSubscriber: PlainSubscriber{ID: id},
		Markets:         markets,
	}
}

type CandleStickSubscriber struct {
	PlainSubscriber
	TimeSpan string
}

func (s *CandleStickSubscriber) Detail() interface{} {
	return s.TimeSpan
}

func (s *CandleStickSubscriber) WriteMsg([]byte) error {
	return nil
}

func (s *CandleStickSubscriber) GetConn() *Conn {
	return nil
}

func NewCandleStickSubscriber(id int64, timespan string) *CandleStickSubscriber {
	return &CandleStickSubscriber{
		PlainSubscriber: PlainSubscriber{ID: id},
		TimeSpan:        timespan,
	}
}

type DepthSubscriber struct {
	PlainSubscriber
	level    string
	payloads []string

	sync.Mutex
}

func (s *DepthSubscriber) Detail() interface{} {
	return s.level
}

func (s *DepthSubscriber) WriteMsg(msg []byte) error {
	s.Lock()
	defer s.Unlock()
	s.payloads = append(s.payloads, string(msg))
	return nil
}

func (s *DepthSubscriber) ClearMsg() {
	s.Lock()
	defer s.Unlock()
	s.payloads = s.payloads[:0]
}

func (s *DepthSubscriber) CompareRet(t *testing.T, actual []string) {
	s.Lock()
	defer s.Unlock()
	assert.EqualValues(t, s.payloads, actual)
}

func (s *DepthSubscriber) GetConn() *Conn {
	return nil
}

func NewDepthSubscriber(id int64, level string) *DepthSubscriber {
	return &DepthSubscriber{
		PlainSubscriber: PlainSubscriber{ID: id},
		level:           level,
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
	BancorDealSubscribeInfo   map[string][]Subscriber
	MarketSubscribeInfo       map[string][]Subscriber

	sync.Mutex
	PushList []pushInfo
}

func (sm *MocSubscribeManager) GetValidatorCommissionInfo() map[string][]Subscriber {
	panic("implement me")
}

func (sm *MocSubscribeManager) GetDelegationRewards() map[string][]Subscriber {
	panic("implement me")
}

func (sm *MocSubscribeManager) PushCreateMarket(subscriber Subscriber, info []byte) {
	sm.Lock()
	defer sm.Unlock()
	sm.PushList = append(sm.PushList, pushInfo{subscriber, string(info)})
}

func (sm *MocSubscribeManager) PushValidatorCommissionInfo(subscriber Subscriber, info []byte) {
	panic("implement me")
}

func (sm *MocSubscribeManager) PushDelegationRewards(subscriber Subscriber, info []byte) {
	panic("implement me")
}

func (sm *MocSubscribeManager) GetMarketSubscribeInfo() map[string][]Subscriber {
	return sm.MarketSubscribeInfo
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

func (sm *MocSubscribeManager) ClearPushList() {
	sm.Lock()
	defer sm.Unlock()
	sm.PushList = sm.PushList[:0]
}

func (sm *MocSubscribeManager) CompareResult(t *testing.T, correct string) {
	sm.Lock()
	defer sm.Unlock()
	out := make([]string, 0, 10)
	for _, info := range sm.PushList {
		var id int64
		switch v := info.Target.(type) {
		case *PlainSubscriber:
			id = v.ID
		case *TickerSubscriber:
			id = v.ID
		case *CandleStickSubscriber:
			id = v.ID
		case *DepthSubscriber:
			id = v.ID
		}
		s := fmt.Sprintf("%d: %s", id, info.Payload)
		out = append(out, s)
	}
	correctList := strings.Split(strings.TrimSpace(correct), "\n")
	sort.Strings(correctList)
	sort.Strings(out)
	assert.Equal(t, correctList, out)
}

func GetDepthSubscribeManeger() *MocSubscribeManager {
	res := &MocSubscribeManager{}
	res.PushList = make([]pushInfo, 0, 100)
	res.DepthSubscribeInfo = make(map[string][]Subscriber)
	res.DepthSubscribeInfo["abc/cet"] = make([]Subscriber, 2)
	res.DepthSubscribeInfo["abc/cet"][0] = NewDepthSubscriber(8, "all")
	res.DepthSubscribeInfo["abc/cet"][1] = NewDepthSubscriber(9, "all")

	return res
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
	res.TickerSubscribeInfo = make([]Subscriber, 2)
	m := make(map[string]struct{})
	m["abc/cet"] = struct{}{}
	m["xyz/cet"] = struct{}{}
	res.TickerSubscribeInfo[0] = NewTickerSubscriber(0, m)
	res.CandleStickSubscribeInfo = make(map[string][]Subscriber)
	res.CandleStickSubscribeInfo["abc/cet"] = make([]Subscriber, 3)
	res.CandleStickSubscribeInfo["abc/cet"][0] = NewCandleStickSubscriber(6, MinuteStr)
	res.CandleStickSubscribeInfo["abc/cet"][1] = NewCandleStickSubscriber(7, HourStr)
	res.CandleStickSubscribeInfo["abc/cet"][2] = NewCandleStickSubscriber(5, DayStr)
	res.DepthSubscribeInfo = make(map[string][]Subscriber)
	res.DepthSubscribeInfo["abc/cet"] = make([]Subscriber, 2)
	res.DepthSubscribeInfo["abc/cet"][0] = NewDepthSubscriber(8, "all")
	res.DepthSubscribeInfo["abc/cet"][1] = NewDepthSubscriber(9, "all")
	res.DepthSubscribeInfo["xyz/cet"] = make([]Subscriber, 1)
	res.DepthSubscribeInfo["xyz/cet"][0] = NewDepthSubscriber(10, "all")
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
	res.LockedSubcribeInfo = make(map[string][]Subscriber)
	res.LockedSubcribeInfo[addr1] = make([]Subscriber, 1)
	res.LockedSubcribeInfo[addr1][0] = &PlainSubscriber{ID: 26}

	m = make(map[string]struct{})
	m["B:xyz/cet"] = struct{}{}
	res.TickerSubscribeInfo[1] = NewTickerSubscriber(27, m)
	res.CandleStickSubscribeInfo["B:xyz/cet"] = make([]Subscriber, 3)
	res.CandleStickSubscribeInfo["B:xyz/cet"][0] = NewCandleStickSubscriber(28, MinuteStr)
	res.CandleStickSubscribeInfo["B:xyz/cet"][1] = NewCandleStickSubscriber(29, HourStr)
	res.CandleStickSubscribeInfo["B:xyz/cet"][2] = NewCandleStickSubscriber(30, DayStr)
	return res
}

func (sm *MocSubscribeManager) SetSkipOption(isSkip bool) {
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
func (sm *MocSubscribeManager) GetBancorDealSubscribeInfo() map[string][]Subscriber {
	return sm.BancorDealSubscribeInfo
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
	sm.Lock()
	defer sm.Unlock()
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushSlash(subscriber Subscriber, info []byte) {
	sm.Lock()
	defer sm.Unlock()
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushHeight(subscriber Subscriber, info []byte) {
	sm.Lock()
	defer sm.Unlock()
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushTicker(subscriber Subscriber, t []*Ticker) {
	sm.Lock()
	defer sm.Unlock()
	info, _ := json.Marshal(t)
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushDepthWithChange(subscriber Subscriber, info []byte) {
	sm.Lock()
	defer sm.Unlock()
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushDepthWithDelta(subscriber Subscriber, info []byte) {
	sm.Lock()
	defer sm.Unlock()
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushCandleStick(subscriber Subscriber, info []byte) {
	sm.Lock()
	defer sm.Unlock()
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushDeal(subscriber Subscriber, info []byte) {
	sm.Lock()
	defer sm.Unlock()
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushCreateOrder(subscriber Subscriber, info []byte) {
	sm.Lock()
	defer sm.Unlock()
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushFillOrder(subscriber Subscriber, info []byte) {
	sm.Lock()
	defer sm.Unlock()
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushCancelOrder(subscriber Subscriber, info []byte) {
	sm.Lock()
	defer sm.Unlock()
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushBancorInfo(subscriber Subscriber, info []byte) {
	sm.Lock()
	defer sm.Unlock()
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushBancorTrade(subscriber Subscriber, info []byte) {
	sm.Lock()
	defer sm.Unlock()
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushBancorDeal(subscriber Subscriber, info []byte) {
	sm.Lock()
	defer sm.Unlock()
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushIncome(subscriber Subscriber, info []byte) {
	sm.Lock()
	defer sm.Unlock()
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushUnbonding(subscriber Subscriber, info []byte) {
	sm.Lock()
	defer sm.Unlock()
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushRedelegation(subscriber Subscriber, info []byte) {
	sm.Lock()
	defer sm.Unlock()
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushUnlock(subscriber Subscriber, info []byte) {
	sm.Lock()
	defer sm.Unlock()
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushTx(subscriber Subscriber, info []byte) {
	sm.Lock()
	defer sm.Unlock()
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushComment(subscriber Subscriber, info []byte) {
	sm.Lock()
	defer sm.Unlock()
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
func (sm *MocSubscribeManager) PushDepthFullMsg(subscriber Subscriber, info []byte) {
	sm.Lock()
	defer sm.Unlock()
	sm.PushList = append(sm.PushList, pushInfo{Target: subscriber, Payload: string(info)})
}
