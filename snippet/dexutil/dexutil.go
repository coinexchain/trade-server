package dexutil

import (
	"time"

	"github.com/coinexchain/dex/modules/market/internal/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/emirpasic/gods/maps/treemap"
)

type baseCandleStick struct {
	openPrice sdk.Dec
	closePrice   sdk.Dec
	highestPrice   sdk.Dec
	lowestPrice   sdk.Dec
	totalDeal  sdk.Int
}

func (cs *baseCandleStick) hasDeal() bool {
	return cs.totalDeal.IsPositive()
}

func (cs *baseCandleStick) update(price sdk.Dec, amount int64) {
	if !cs.hasDeal() {
		cs.openPrice = price
	}
	cs.closePrice = price
	if !cs.hasDeal() {
		cs.highestPrice = price
		cs.lowestPrice = price
	} else {
		if cs.highestPrice.LT(price) {
			cs.highestPrice = price
		}
		if cs.lowestPrice.GT(price) {
			cs.lowestPrice = price
		}
	}
	cs.totalDeal = cs.totalDeal.AddRaw(amount)
}

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
			cs.highestPrice = sub.highestPrice
			cs.lowestPrice = sub.lowestPrice
		} else {
			if cs.highestPrice.LT(sub.highestPrice) {
				cs.highestPrice = sub.highestPrice
			}
			if cs.lowestPrice.GT(sub.lowestPrice) {
				cs.lowestPrice = sub.lowestPrice
			}
		}
		cs.totalDeal = cs.totalDeal.Add(sub.totalDeal)
	}
	return
}

//============================================================

type CandleStickRecord struct {
	minuteCS       [60]baseCandleStick
	hourCS         [24]baseCandleStick
	lastUpdateTime time.Time
}

type CandleStick struct {
	BeginPrice     sdk.Dec `json:"begin_price"`
	EndPrice       sdk.Dec `json:"end_price"`
	MaxPrice       sdk.Dec `json:"max_price"`
	MinPrice       sdk.Dec `json:"min_price"`
	TotalDeal      sdk.Int `json:"total_deal"`
	EndingUnixTime int64   `json:"unix_time"`
	TimeSpan       string  `json:"type"`
	MarketSymbol   string  `json:"symbol"`
}

func newCandleStick(cs baseCandleStick, t int64, span string) *CandleStick {
	return &CandleStick{
		BeginPrice:     cs.openPrice,
		EndPrice:       cs.closePrice,
		MaxPrice:       cs.highestPrice,
		MinPrice:       cs.lowestPrice,
		TotalDeal:      cs.totalDeal,
		EndingUnixTime: t,
		TimeSpan:       span,
	}
}

func (cs *CandleStickRecord) newBlock(t time.Time, isNewDay, isNewHour, isNewMinute bool) []*CandleStick {
	res := make([]*CandleStick, 0, 3)
	if isNewDay && t.Unix()-cs.lastUpdateTime.Unix() < 60*60*24 {
		dayCS := merge(cs.hourCS[:])
		res = append(res, newCandleStick(dayCS, cs.lastUpdateTime.Unix(), "day"))
	}
	if isNewHour && t.Unix()-cs.lastUpdateTime.Unix() < 60*60 {
		cs.hourCS[cs.lastUpdateTime.Hour()] = merge(cs.minuteCS[:])
		res = append(res, newCandleStick(cs.hourCS[cs.lastUpdateTime.Hour()], cs.lastUpdateTime.Unix(), "hour"))
	}
	if isNewMinute && t.Unix()-cs.lastUpdateTime.Unix() < 60 {
		res = append(res, newCandleStick(cs.minuteCS[cs.lastUpdateTime.Minute()], cs.lastUpdateTime.Unix(), "day"))
	}

	if isNewDay {
		for i := 0; i < 24; i++ {
			cs.hourCS[i] = baseCandleStick{}
		}
		for i := 0; i < 60; i++ {
			cs.minuteCS[i] = baseCandleStick{}
		}
	} else if isNewHour {
		for i := 0; i < 60; i++ {
			cs.minuteCS[i] = baseCandleStick{}
		}
	}
	return res
}

func (cs *CandleStickRecord) Update(t time.Time, price sdk.Dec, amount int64) {
	cs.minuteCS[t.Minute()].update(price, amount)
	cs.lastUpdateTime = t
}

//=====================================

type CandleStickManager struct {
	csrMap        map[string]*CandleStickRecord
	lastBlockTime time.Time
}

func NewCandleStickManager(marketSymbols []string) CandleStickManager {
	res := CandleStickManager{csrMap: make(map[string]*CandleStickRecord)}
	for _, sym := range marketSymbols {
		res.csrMap[sym] = &CandleStickRecord{}
	}
	return res
}

func (manager *CandleStickManager) Add(sym string) {
	manager.csrMap[sym] = &CandleStickRecord{}
}

func (manager *CandleStickManager) NewBlock(t time.Time) []CandleStick {
	res := make([]CandleStick, 0, 100)
	isNewDay := t.Day() != manager.lastBlockTime.Day() || t.Unix()-manager.lastBlockTime.Unix() > 60*60*24
	isNewHour := t.Hour() != manager.lastBlockTime.Hour() || t.Unix()-manager.lastBlockTime.Unix() > 60*60
	isNewMinute := t.Minute() != manager.lastBlockTime.Minute() || t.Unix()-manager.lastBlockTime.Unix() > 60
	for sym, csr := range manager.csrMap {
		for _, tcs := range csr.newBlock(t, isNewDay, isNewHour, isNewMinute) {
			tcs.MarketSymbol = sym
			res = append(res, *tcs)
		}
	}
	manager.lastBlockTime = t
	return res
}

func (manager *CandleStickManager) GetRecord(marketSymbol string) *CandleStickRecord {
	csr, ok := manager.csrMap[marketSymbol]
	if ok {
		return csr
	}
	return nil
}

//=====================================

type DepthManager struct {
	ppMap *treemap.Map //map[string]*types.PricePoint
	updated map[*types.PricePoint]bool
}

func DefaultDepthManager() *DepthManager {
	return &DepthManager{
		ppMap: treemap.NewWithStringComparator(),
		updated: make(map[*types.PricePoint]bool),
	}
}

func NewDepthManager(ppList []*types.PricePoint) *DepthManager {
	res:=DefaultDepthManager()
	for _, pp := range ppList {
		res.Change(pp.Price, pp.LeftStock)
	}
	return res
}

func (dm *DepthManager) Change(price sdk.Dec, amount sdk.Int) {
	s := string(types.DecToBigEndianBytes(price))
	ptr, ok := dm.ppMap.Get(s)
	pp := ptr.(*types.PricePoint)
	if !ok {
		pp = &types.PricePoint{Price:price, LeftStock:sdk.ZeroInt()}
	}
	pp.LeftStock = pp.LeftStock.Add(amount)
	if pp.LeftStock.IsZero() {
		dm.ppMap.Remove(s)
	} else {
		dm.ppMap.Put(s, pp)
	}
	dm.updated[pp]=true
}

func (dm *DepthManager) EndBlock() map[*types.PricePoint]bool {
	ret := dm.updated
	dm.updated=make(map[*types.PricePoint]bool)
	return ret
}

func (dm *DepthManager) GetLowest(n int) []*types.PricePoint {
	res:=make([]*types.PricePoint,0,n)
	iter:=dm.ppMap.Iterator()
	iter.Begin()
	for i:=0; i<n; i++ {
		if ok:=iter.Next(); !ok {
			break
		}
		res=append(res, iter.Value().(*types.PricePoint))
	}
	return res
}

func (dm *DepthManager) GetHighest(n int) []*types.PricePoint {
	res:=make([]*types.PricePoint,0,n)
	iter:=dm.ppMap.Iterator()
	iter.End()
	for i:=0; i<n; i++ {
		if ok:=iter.Prev(); !ok {
			break
		}
		res=append(res, iter.Value().(*types.PricePoint))
	}
	return res
}

