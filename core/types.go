package core

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	CreateOrderEndByte = byte('c')
	FillOrderEndByte   = byte('f')
	CancelOrderEndByte = byte('d')
	Minute             = byte(0x10)
	Hour               = byte(0x20)
	Day                = byte(0x30)
	MinuteStr          = "1min"
	HourStr            = "1hour"
	DayStr             = "1day"
	CreateOrderStr     = "create"
	FillOrderStr       = "fill"
	CancelOrderStr     = "cancel"
)

// push candle stick msg to ws
type CandleStick struct {
	OpenPrice      sdk.Dec `json:"open"`
	ClosePrice     sdk.Dec `json:"close"`
	HighPrice      sdk.Dec `json:"high"`
	LowPrice       sdk.Dec `json:"low"`
	TotalDeal      sdk.Int `json:"total"`
	EndingUnixTime int64   `json:"unix_time"`
	TimeSpan       string  `json:"time_span"`
	Market         string  `json:"market"`
}

type Ticker struct {
	Market            string  `json:"market"`
	NewPrice          sdk.Dec `json:"new"`
	OldPriceOneDayAgo sdk.Dec `json:"old"`
	MinuteInDay       int     `json:"minute_in_day"`
}

type XTicker struct {
	Market            string  `json:"market"`
	NewPrice          sdk.Dec `json:"new"`
	OldPriceOneDayAgo sdk.Dec `json:"old"`
	MinuteInDay       int     `json:"minute_in_day"`
	HighPrice         sdk.Dec `json:"high"`
	LowPrice          sdk.Dec `json:"low"`
	TotalDeal         sdk.Int `json:"volume"`
}

type Donation struct {
	Sender string `json:"sender"`
	Amount string `json:"amount"`
}

type PricePoint struct {
	Price  sdk.Dec `json:"p"`
	Amount sdk.Int `json:"a"`
}

type Subscriber interface {
	Detail() interface{}
	WriteMsg([]byte) error
	GetConn() *Conn
}

type SubscribeManager interface {
	GetSlashSubscribeInfo() []Subscriber
	GetHeightSubscribeInfo() []Subscriber

	//The returned subscribers have detailed information of markets
	//one subscriber can subscribe tickers from no more than 100 markets
	GetTickerSubscribeInfo() []Subscriber

	//The returned subscribers have detailed information of timespan
	GetCandleStickSubscribeInfo() map[string][]Subscriber

	//the map keys are markets' names
	GetDepthSubscribeInfo() map[string][]Subscriber
	GetDealSubscribeInfo() map[string][]Subscriber

	//the map keys are bancor contracts' names
	GetBancorInfoSubscribeInfo() map[string][]Subscriber

	//the map keys are tokens' names
	GetCommentSubscribeInfo() map[string][]Subscriber

	//the map keys are accounts' bech32 addresses
	GetMarketSubscribeInfo() map[string][]Subscriber
	GetOrderSubscribeInfo() map[string][]Subscriber
	GetBancorTradeSubscribeInfo() map[string][]Subscriber
	GetBancorDealSubscribeInfo() map[string][]Subscriber
	GetIncomeSubscribeInfo() map[string][]Subscriber
	GetUnbondingSubscribeInfo() map[string][]Subscriber
	GetRedelegationSubscribeInfo() map[string][]Subscriber
	GetUnlockSubscribeInfo() map[string][]Subscriber
	GetTxSubscribeInfo() map[string][]Subscriber
	GetLockedSubscribeInfo() map[string][]Subscriber
	GetValidatorCommissionInfo() map[string][]Subscriber
	GetDelegationRewards() map[string][]Subscriber

	PushLockedSendMsg(subscriber Subscriber, info []byte)
	PushSlash(subscriber Subscriber, info []byte)
	PushHeight(subscriber Subscriber, info []byte)
	PushTicker(subscriber Subscriber, t []*Ticker)
	PushDepthFullMsg(subscriber Subscriber, info []byte)
	PushDepthWithChange(subscriber Subscriber, info []byte)
	PushDepthWithDelta(subscriber Subscriber, delta []byte)
	PushCandleStick(subscriber Subscriber, info []byte)
	PushDeal(subscriber Subscriber, info []byte)
	PushBancorDeal(subscriber Subscriber, info []byte)
	PushCreateMarket(subscriber Subscriber, info []byte)
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
	PushValidatorCommissionInfo(subscriber Subscriber, info []byte)
	PushDelegationRewards(subscriber Subscriber, info []byte)

	SetSkipOption(isSkip bool)
}

type Consumer interface {
	ConsumeMessage(msgType string, bz []byte)
}

type Querier interface {
	// All these functions can be safely called in goroutines.
	QueryTickers(marketList []string) []*Ticker
	QueryBlockTime(height int64, count int) []int64
	QueryDepth(market string, count int) (sell []*PricePoint, buy []*PricePoint)
	QueryCandleStick(market string, timespan byte, time int64, sid int64, count int) []json.RawMessage

	QueryLocked(account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64)
	QueryDeal(market string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64)
	QueryBancorDeal(market string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64)
	QueryBancorInfo(market string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64)
	QueryBancorTrade(account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64)
	QueryRedelegation(account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64)
	QueryUnbonding(account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64)
	QueryUnlock(account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64)
	QueryIncome(account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64)
	QueryTx(account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64)
	QueryComment(token string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64)
	QuerySlash(time int64, sid int64, count int) (data []json.RawMessage, timesid []int64)
	QueryDonation(time int64, sid int64, count int) (data []json.RawMessage, timesid []int64)
	QueryDelist(market string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64)

	QueryOrderAboutToken(tag, token, account string, time int64, sid int64, count int) (data []json.RawMessage, tags []byte, timesid []int64)
	QueryLockedAboutToken(token, account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64)
	QueryBancorTradeAboutToken(token, account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64)
	QueryUnlockAboutToken(token, account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64)
	QueryIncomeAboutToken(token, account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64)
	QueryTxAboutToken(token, account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64)

	QueryTxByHashID(hexHashID string) json.RawMessage
}

type Pruneable interface {
	SetPruneTimestamp(t uint64)
	GetPruneTimestamp() uint64
}
