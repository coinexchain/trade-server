package core

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	CreateOrderEndByte = byte('c')
	FillOrderEndByte   = byte('f')
	CancelOrderEndByte = byte('d')
	Minute             = byte(0x10)
	Hour               = byte(0x20)
	Day                = byte(0x30)
)

type CandleStick struct {
	BeginPrice     sdk.Dec `json:"begin"`
	EndPrice       sdk.Dec `json:"end"`
	MaxPrice       sdk.Dec `json:"max"`
	MinPrice       sdk.Dec `json:"min"`
	TotalDeal      sdk.Int `json:"total"`
	EndingUnixTime int64   `json:"unix_time"`
	TimeSpan       byte    `json:"time_span"`
	MarketSymbol   string  `json:"market"`
}

type Ticker struct {
	Market            string  `json:"market"`
	NewPrice          sdk.Dec `json:"new"`
	OldPriceOneDayAgo sdk.Dec `json:"old"`
}

type PricePoint struct {
	Price  sdk.Dec `json:"p"`
	Amount sdk.Int `json:"a"`
}

type Subscriber interface {
	Detail() interface{}
}

type SubscribeManager interface {
	GetSlashSubscribeInfo() []Subscriber
	GetHeightSubscribeInfo() []Subscriber
	GetTickerSubscribeInfo() []Subscriber
	GetCandleStickSubscribeInfo() map[string][]Subscriber

	GetDepthSubscribeInfo() map[string][]Subscriber
	GetDealSubscribeInfo() map[string][]Subscriber
	GetOrderSubscribeInfo() map[string][]Subscriber
	GetBancorInfoSubscribeInfo() map[string][]Subscriber
	GetBancorTradeSubscribeInfo() map[string][]Subscriber
	GetIncomeSubscribeInfo() map[string][]Subscriber
	GetUnbondingSubscribeInfo() map[string][]Subscriber
	GetRedelegationSubscribeInfo() map[string][]Subscriber
	GetUnlockSubscribeInfo() map[string][]Subscriber
	GetTxSubscribeInfo() map[string][]Subscriber
	GetCommentSubscribeInfo() map[string][]Subscriber

	PushSlash(subscriber Subscriber, info []byte)
	PushHeight(subscriber Subscriber, info []byte)
	PushTicker(subscriber Subscriber, t []*Ticker)
	PushDepthSell(subscriber Subscriber, info []byte)
	PushDepthBuy(subscriber Subscriber, info []byte)
	PushCandleStick(subscriber Subscriber, cs *CandleStick)
	PushDeal(subscriber Subscriber, info []byte)
	PushCreateOrder(subscriber Subscriber, info []byte)
	PushFillOrder(subscriber Subscriber, info []byte)
	PushCancelOrder(subscriber Subscriber, info []byte)
	PushBancorInfo(subscriber Subscriber, info []byte)
	PushBancorTrade(subscriber Subscriber, info []byte)
	PushIncome(subscriber Subscriber, info []byte)
	PushUnbonding(subscriber Subscriber, info []byte)
	PushRedelegation(subscriber Subscriber, info []byte)
	PushUnlock(subscriber Subscriber, info []byte)
	PushTx(subscriber Subscriber, info []byte)
	PushComment(subscriber Subscriber, info []byte)
}

type Consumer interface {
	ConsumeMessage(msgType string, bz []byte)
}

type Querier interface {
	QueryTikers(marketList []string) []*Ticker
	QueryBlockTime(height int64, count int) []int64
	QueryDepth(market string, count int) (sell []*PricePoint, buy []*PricePoint)
	QueryCandleStick(market string, timespan byte, unixSec int64, count int) [][]byte
	QueryDeal(market string, unixSec int64, count int) [][]byte
	QueryOrder(account string, unixSec int64, count int) (data [][]byte, tags []byte)
	QueryBancorInfo(market string, unixSec int64, count int) [][]byte
	QueryBancorTrade(account string, unixSec int64, count int) [][]byte
	QueryRedelegation(account string, unixSec int64, count int) [][]byte
	QueryUnbonding(account string, unixSec int64, count int) [][]byte
	QueryUnlock(account string, unixSec int64, count int) [][]byte
	QueryIncome(account string, unixSec int64, count int) [][]byte
	QueryTx(account string, unixSec int64, count int) [][]byte
	QueryComment(token string, unixSec int64, count int) [][]byte
}
