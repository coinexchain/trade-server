package core

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/emirpasic/gods/maps/treemap"
)

type baseCandleStick struct {
	OpenPrice  sdk.Dec `json:"open"`
	ClosePrice sdk.Dec `json:"close"`
	HighPrice  sdk.Dec `json:"high"`
	LowPrice   sdk.Dec `json:"low"`
	TotalDeal  sdk.Int `json:"total"`
}

func defaultBaseCandleStick() baseCandleStick {
	return baseCandleStick{
		OpenPrice:  sdk.ZeroDec(),
		ClosePrice: sdk.ZeroDec(),
		HighPrice:  sdk.ZeroDec(),
		LowPrice:   sdk.ZeroDec(),
		TotalDeal:  sdk.ZeroInt(),
	}
}

func (cs *baseCandleStick) hasDeal() bool {
	return cs.TotalDeal.IsPositive()
}

// When new deal comes, update candle stick accordingly
func (cs *baseCandleStick) update(price sdk.Dec, amount int64) {
	if !cs.hasDeal() {
		cs.OpenPrice = price
	}
	cs.ClosePrice = price
	if !cs.hasDeal() {
		cs.HighPrice = price
		cs.LowPrice = price
	} else {
		if cs.HighPrice.LT(price) {
			cs.HighPrice = price
		}
		if cs.LowPrice.GT(price) {
			cs.LowPrice = price
		}
	}
	cs.TotalDeal = cs.TotalDeal.AddRaw(amount)
}

// merge some candle sticks with smaller time span, into one candle stick with larger time span
func merge(subList []baseCandleStick) (cs baseCandleStick) {
	cs = defaultBaseCandleStick()
	for _, sub := range subList {
		if !sub.hasDeal() {
			continue
		}
		if !cs.hasDeal() {
			cs.OpenPrice = sub.OpenPrice
		}
		cs.ClosePrice = sub.ClosePrice
		if !cs.hasDeal() {
			cs.HighPrice = sub.HighPrice
			cs.LowPrice = sub.LowPrice
		} else {
			if cs.HighPrice.LT(sub.HighPrice) {
				cs.HighPrice = sub.HighPrice
			}
			if cs.LowPrice.GT(sub.LowPrice) {
				cs.LowPrice = sub.LowPrice
			}
		}
		cs.TotalDeal = cs.TotalDeal.Add(sub.TotalDeal)
	}
	return
}

//============================================================

// Record the candle sticks within one day
type CandleStickRecord struct {
	MinuteCS       [60]baseCandleStick `json:"minute_cs"`
	HourCS         [24]baseCandleStick `json:"hour_cs"`
	LastUpdateTime time.Time           `json:"last_update"`
	Market         string              `json:"Market"`
}

func NewCandleStickRecord(market string) *CandleStickRecord {
	res := &CandleStickRecord{Market: market, LastUpdateTime: time.Unix(0, 0)}
	for i := 0; i < len(res.MinuteCS); i++ {
		res.MinuteCS[i] = defaultBaseCandleStick()
	}
	for i := 0; i < len(res.HourCS); i++ {
		res.HourCS[i] = defaultBaseCandleStick()
	}
	return res
}

func getSpanStrFromSpan(span byte) string {
	switch span {
	case Minute:
		return MinuteStr
	case Hour:
		return HourStr
	case Day:
		return DayStr
	default:
		return ""
	}
}

func getSpanFromSpanStr(spanStr string) byte {
	switch spanStr {
	case MinuteStr:
		return Minute
	case HourStr:
		return Hour
	case DayStr:
		return Day
	default:
		return 0
	}
}

func (csr *CandleStickRecord) newCandleStick(cs baseCandleStick, t int64, span byte) CandleStick {
	return CandleStick{
		OpenPrice:      cs.OpenPrice,
		ClosePrice:     cs.ClosePrice,
		HighPrice:      cs.HighPrice,
		LowPrice:       cs.LowPrice,
		TotalDeal:      cs.TotalDeal,
		EndingUnixTime: t,
		TimeSpan:       getSpanStrFromSpan(span),
		Market:         csr.Market,
	}
}

// When a new block comes, flush the pending candle sticks
func (csr *CandleStickRecord) newBlock(isNewDay, isNewHour, isNewMinute bool) []CandleStick {
	res := make([]CandleStick, 0, 3)
	lastTime := csr.LastUpdateTime.Unix()
	if isNewMinute && lastTime != 0 {
		cs := csr.newCandleStick(csr.MinuteCS[csr.LastUpdateTime.Minute()], csr.LastUpdateTime.Unix(), Minute)
		res = append(res, cs)
	}
	if isNewHour && lastTime != 0 {
		csr.HourCS[csr.LastUpdateTime.Hour()] = merge(csr.MinuteCS[:])
		cs := csr.newCandleStick(csr.HourCS[csr.LastUpdateTime.Hour()], csr.LastUpdateTime.Unix(), Hour)
		res = append(res, cs)
	}
	if isNewDay && lastTime != 0 {
		dayCS := merge(csr.HourCS[:])
		cs := csr.newCandleStick(dayCS, csr.LastUpdateTime.Unix(), Day)
		res = append(res, cs)
	}

	if isNewDay {
		for i := 0; i < 24; i++ {
			csr.HourCS[i] = defaultBaseCandleStick()
		}
	}
	if isNewDay || isNewHour {
		for i := 0; i < 60; i++ {
			csr.MinuteCS[i] = defaultBaseCandleStick()
		}
	}
	return res
}

func (csr *CandleStickRecord) Update(t time.Time, price sdk.Dec, amount int64) {
	csr.MinuteCS[t.Minute()].update(price, amount)
	csr.LastUpdateTime = t
}

//=====================================

// Manager of the CandleStickRecords for all markets
type CandleStickManager struct {
	CsrMap        map[string]*CandleStickRecord `json:"csr_map"`
	LastBlockTime time.Time                     `json:"last_block_time"`
}

func NewCandleStickManager(markets []string) CandleStickManager {
	res := CandleStickManager{CsrMap: make(map[string]*CandleStickRecord)}
	for _, market := range markets {
		res.CsrMap[market] = NewCandleStickRecord(market)
	}
	return res
}

func (manager *CandleStickManager) AddMarket(market string) {
	manager.CsrMap[market] = NewCandleStickRecord(market)
}

func (manager *CandleStickManager) NewBlock(t time.Time) []CandleStick {
	res := make([]CandleStick, 0, 100)
	isNewDay := t.Day() != manager.LastBlockTime.Day() || t.Unix()-manager.LastBlockTime.Unix() > 60*60*24
	isNewHour := t.Hour() != manager.LastBlockTime.Hour() || t.Unix()-manager.LastBlockTime.Unix() > 60*60
	isNewMinute := t.Minute() != manager.LastBlockTime.Minute() || t.Unix()-manager.LastBlockTime.Unix() > 60
	for _, csr := range manager.CsrMap {
		csSlice := csr.newBlock(isNewDay, isNewHour, isNewMinute)
		res = append(res, csSlice...)
	}
	manager.LastBlockTime = t
	return res
}

func (manager *CandleStickManager) GetRecord(Market string) *CandleStickRecord {
	csr, ok := manager.CsrMap[Market]
	if ok {
		return csr
	}
	return nil
}

//=====================================

// Manager for the depth information of one side of the order book: sell or buy
type DepthManager struct {
	ppMap   *treemap.Map //map[string]*PricePoint
	Updated map[string]*PricePoint
}

func (dm *DepthManager) Size() int {
	return dm.ppMap.Size()
}

func (dm *DepthManager) DumpPricePoints() []*PricePoint {
	size := dm.ppMap.Size()
	pps := make([]*PricePoint, size)
	iter := dm.ppMap.Iterator()
	iter.Begin()
	for i := 0; i < size; i++ {
		iter.Next()
		pps[i] = iter.Value().(*PricePoint)
	}
	return pps
}

func DefaultDepthManager() *DepthManager {
	return &DepthManager{
		ppMap:   treemap.NewWithStringComparator(),
		Updated: make(map[string]*PricePoint),
	}
}

// positive amount for increment, negative amount for decrement
func (dm *DepthManager) DeltaChange(price sdk.Dec, amount sdk.Int) {
	s := string(DecToBigEndianBytes(price))
	ptr, ok := dm.ppMap.Get(s)
	var pp *PricePoint
	if !ok {
		pp = &PricePoint{Price: price, Amount: sdk.ZeroInt()}
	} else {
		pp = ptr.(*PricePoint)
	}
	pp.Amount = pp.Amount.Add(amount)
	if pp.Amount.IsZero() {
		dm.ppMap.Remove(s)
	} else {
		dm.ppMap.Put(s, pp)
	}
	dm.Updated[s] = pp
}

// returns the changed PricePoints of last block. Clear dm.Updated for the next block
func (dm *DepthManager) EndBlock() map[string]*PricePoint {
	ret := dm.Updated
	dm.Updated = make(map[string]*PricePoint)
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

const MinuteNumInDay = 24 * 60

type TickerManager struct {
	PriceList    [MinuteNumInDay]sdk.Dec `json:"price_list"`
	NewestPrice  sdk.Dec                 `json:"new_price"`
	NewestMinute int                     `json:"new_minute"`
	Market       string                  `json:"market"`
	Initialized  bool                    `json:"initialized"`
}

func DefaultTickerManager(Market string) *TickerManager {
	return &TickerManager{
		Market: Market,
	}
}

// Flush the cached NewestPrice and NewestMinute to PriceList,
// and assign currPrice to NewestPrice, currMinute to NewestMinute
func (tm *TickerManager) UpdateNewestPrice(currPrice sdk.Dec, currMinute int) {
	if currMinute >= MinuteNumInDay || currMinute < 0 {
		panic("Minute too large")
	}
	if !tm.Initialized {
		tm.Initialized = true
		for i := 0; i < MinuteNumInDay; i++ {
			tm.PriceList[i] = currPrice
		}
		tm.NewestPrice = currPrice
		tm.NewestMinute = currMinute
		return
	}
	tm.PriceList[tm.NewestMinute] = tm.NewestPrice
	for {
		tm.NewestMinute++
		if tm.NewestMinute >= MinuteNumInDay {
			tm.NewestMinute = 0
		}
		if tm.NewestMinute == currMinute {
			break
		}
		tm.PriceList[tm.NewestMinute] = tm.NewestPrice
	}
	tm.NewestPrice = currPrice
}

// Return a Ticker if NewPrice or OldPriceOneDayAgo is different from its previous minute
func (tm *TickerManager) GetTicker(currMinute int) *Ticker {
	if !tm.Initialized {
		return nil
	}
	if currMinute >= MinuteNumInDay || currMinute < 0 {
		panic("Minute too large")
	}
	lastMinute := currMinute - 1
	if lastMinute < 0 {
		lastMinute = MinuteNumInDay - 1
	}
	if tm.NewestMinute == currMinute || !tm.PriceList[currMinute].Equal(tm.PriceList[lastMinute]) {
		return &Ticker{
			NewPrice:          tm.NewestPrice,
			OldPriceOneDayAgo: tm.PriceList[currMinute],
			Market:            tm.Market,
		}
	}
	return nil
}
