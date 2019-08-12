package core

import (
	"time"

	"github.com/coinexchain/dex/modules/market"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/emirpasic/gods/maps/treemap"
)

type baseCandleStick struct {
	openPrice  sdk.Dec
	closePrice sdk.Dec
	highPrice  sdk.Dec
	lowPrice   sdk.Dec
	totalDeal  sdk.Int
}

func (cs *baseCandleStick) hasDeal() bool {
	return cs.totalDeal.IsPositive()
}

// When new deal comes, update candle stick accordingly
func (cs *baseCandleStick) update(price sdk.Dec, amount int64) {
	if !cs.hasDeal() {
		cs.openPrice = price
	}
	cs.closePrice = price
	if !cs.hasDeal() {
		cs.highPrice = price
		cs.lowPrice = price
	} else {
		if cs.highPrice.LT(price) {
			cs.highPrice = price
		}
		if cs.lowPrice.GT(price) {
			cs.lowPrice = price
		}
	}
	cs.totalDeal = cs.totalDeal.AddRaw(amount)
}

// merge some candle sticks with smaller time span, into one candle stick with larger time span
func merge(subList []baseCandleStick) (cs baseCandleStick) {
	for _, sub := range subList {
		if !sub.hasDeal() {
			continue
		}
		if !cs.hasDeal() {
			cs.openPrice = sub.openPrice
		}
		cs.closePrice = sub.closePrice
		if !cs.hasDeal() {
			cs.highPrice = sub.highPrice
			cs.lowPrice = sub.lowPrice
		} else {
			if cs.highPrice.LT(sub.highPrice) {
				cs.highPrice = sub.highPrice
			}
			if cs.lowPrice.GT(sub.lowPrice) {
				cs.lowPrice = sub.lowPrice
			}
		}
		cs.totalDeal = cs.totalDeal.Add(sub.totalDeal)
	}
	return
}

//============================================================

// Record the candle sticks within one day
type CandleStickRecord struct {
	minuteCS       [60]baseCandleStick
	hourCS         [24]baseCandleStick
	lastUpdateTime time.Time
	market         string
}

func (csr *CandleStickRecord) newCandleStick(cs baseCandleStick, t int64, span byte) CandleStick {
	return CandleStick{
		OpenPrice:      cs.openPrice,
		ClosePrice:     cs.closePrice,
		HighPrice:      cs.highPrice,
		LowPrice:       cs.lowPrice,
		TotalDeal:      cs.totalDeal,
		EndingUnixTime: t,
		TimeSpan:       span,
		Market:         csr.market,
	}
}

// When a new block comes, flush the pending candle sticks
func (csr *CandleStickRecord) newBlock(t time.Time, isNewDay, isNewHour, isNewMinute bool) []CandleStick {
	res := make([]CandleStick, 0, 3)
	if isNewMinute {
		cs := csr.newCandleStick(csr.minuteCS[csr.lastUpdateTime.Minute()], csr.lastUpdateTime.Unix(), Minute)
		res = append(res, cs)
	}
	if isNewHour {
		csr.hourCS[csr.lastUpdateTime.Hour()] = merge(csr.minuteCS[:])
		cs := csr.newCandleStick(csr.hourCS[csr.lastUpdateTime.Hour()], csr.lastUpdateTime.Unix(), Hour)
		res = append(res, cs)
	}
	if isNewDay {
		dayCS := merge(csr.hourCS[:])
		cs := csr.newCandleStick(dayCS, csr.lastUpdateTime.Unix(), Day)
		res = append(res, cs)
	}

	if isNewDay {
		for i := 0; i < 24; i++ {
			csr.hourCS[i] = baseCandleStick{}
		}
	}
	if isNewDay || isNewHour {
		for i := 0; i < 60; i++ {
			csr.minuteCS[i] = baseCandleStick{}
		}
	}
	return res
}

func (csr *CandleStickRecord) Update(t time.Time, price sdk.Dec, amount int64) {
	csr.minuteCS[t.Minute()].update(price, amount)
	csr.lastUpdateTime = t
}

//=====================================

// Manager of the CandleStickRecords for all markets
type CandleStickManager struct {
	csrMap        map[string]*CandleStickRecord
	lastBlockTime time.Time
}

func NewCandleStickManager(markets []string) CandleStickManager {
	res := CandleStickManager{csrMap: make(map[string]*CandleStickRecord)}
	for _, market := range markets {
		res.csrMap[market] = &CandleStickRecord{}
	}
	return res
}

func (manager *CandleStickManager) AddMarket(market string) {
	manager.csrMap[market] = &CandleStickRecord{market: market}
}

func (manager *CandleStickManager) NewBlock(t time.Time) []CandleStick {
	res := make([]CandleStick, 0, 100)
	isNewDay := t.Day() != manager.lastBlockTime.Day() || t.Unix()-manager.lastBlockTime.Unix() > 60*60*24
	isNewHour := t.Hour() != manager.lastBlockTime.Hour() || t.Unix()-manager.lastBlockTime.Unix() > 60*60
	isNewMinute := t.Minute() != manager.lastBlockTime.Minute() || t.Unix()-manager.lastBlockTime.Unix() > 60
	for _, csr := range manager.csrMap {
		csSlice := csr.newBlock(t, isNewDay, isNewHour, isNewMinute)
		res = append(res, csSlice...)
	}
	manager.lastBlockTime = t
	return res
}

func (manager *CandleStickManager) GetRecord(market string) *CandleStickRecord {
	csr, ok := manager.csrMap[market]
	if ok {
		return csr
	}
	return nil
}

//=====================================

// Manager for the depth information of one side of the order book: sell or buy
type DepthManager struct {
	ppMap   *treemap.Map //map[string]*PricePoint
	updated map[*PricePoint]bool
}

func DefaultDepthManager() *DepthManager {
	return &DepthManager{
		ppMap:   treemap.NewWithStringComparator(),
		updated: make(map[*PricePoint]bool),
	}
}

// positive amount for increment, negative amount for decrement
func (dm *DepthManager) DeltaChange(price sdk.Dec, amount sdk.Int) {
	s := string(market.DecToBigEndianBytes(price))
	ptr, ok := dm.ppMap.Get(s)
	pp := ptr.(*PricePoint)
	if !ok {
		pp = &PricePoint{Price: price, Amount: sdk.ZeroInt()}
	}
	pp.Amount = pp.Amount.Add(amount)
	if pp.Amount.IsZero() {
		dm.ppMap.Remove(s)
	} else {
		dm.ppMap.Put(s, pp)
	}
	dm.updated[pp] = true
}

// returns the changed PricePoints of last block. Clear dm.updated for the next block
func (dm *DepthManager) EndBlock() map[*PricePoint]bool {
	ret := dm.updated
	dm.updated = make(map[*PricePoint]bool)
	return ret
}

// Returns the lowest n PricePoints
func (dm *DepthManager) GetLowest(n int) []*PricePoint {
	res := make([]*PricePoint, 0, n)
	iter := dm.ppMap.Iterator()
	iter.Begin()
	for i := 0; i < n; i++ {
		if ok := iter.Next(); !ok {
			break
		}
		res = append(res, iter.Value().(*PricePoint))
	}
	return res
}

// Returns the highest n PricePoints
func (dm *DepthManager) GetHighest(n int) []*PricePoint {
	res := make([]*PricePoint, 0, n)
	iter := dm.ppMap.Iterator()
	iter.End()
	for i := 0; i < n; i++ {
		if ok := iter.Prev(); !ok {
			break
		}
		res = append(res, iter.Value().(*PricePoint))
	}
	return res
}

//=====================================

type priceWithUpdate struct {
	price   sdk.Dec
	updated bool
}

const MinuteNumInDay = 24 * 60

type TickerManager struct {
	priceList    [MinuteNumInDay]priceWithUpdate
	newestPrice  sdk.Dec
	newestMinute int
	market       string
	initialized  bool
}

func DefaultTickerManager(market string) *TickerManager {
	return &TickerManager{
		market: market,
	}
}

// Flush the cached newestPrice and newestMinute to priceList,
// and assign currPrice to newestPrice, currMinute to newestMinute
func (tm *TickerManager) UpdateNewestPrice(currPrice sdk.Dec, currMinute int) {
	if currMinute >= MinuteNumInDay || currMinute < 0 {
		panic("Minute too large")
	}
	if !tm.initialized {
		tm.initialized = true
		for i := 0; i < MinuteNumInDay; i++ {
			tm.priceList[i].price = currPrice
		}
		tm.newestPrice = currPrice
		tm.newestMinute = currMinute
		return
	}
	tm.priceList[tm.newestMinute].price = tm.newestPrice
	tm.priceList[tm.newestMinute].updated = true
	for tm.newestMinute++; tm.newestMinute != currMinute; tm.newestMinute++ {
		if tm.newestMinute == MinuteNumInDay {
			tm.newestMinute = 0
		}
		tm.priceList[tm.newestMinute].price = tm.newestPrice
		tm.priceList[tm.newestMinute].updated = false
	}
	tm.newestPrice = currPrice
}

// Return a Ticker if NewPrice or OldPriceOneDayAgo is different from its previous minute
func (tm *TickerManager) GetTiker(currMinute int) *Ticker {
	if currMinute >= MinuteNumInDay || currMinute < 0 {
		panic("Minute too large")
	}
	if tm.newestMinute != currMinute && !tm.priceList[currMinute].updated {
		return nil
	}
	return &Ticker{
		NewPrice:          tm.newestPrice,
		OldPriceOneDayAgo: tm.priceList[currMinute].price,
		Market:            tm.market,
	}
}
