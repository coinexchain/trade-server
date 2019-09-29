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
	MinuteCS         [60]baseCandleStick `json:"minute_cs"`
	HourCS           [24]baseCandleStick `json:"hour_cs"`
	LastUpdateTime   time.Time           `json:"last_update"`
	LastMinuteCSTime int64               `json:"last_minute_cs_time"`
	LastHourCSTime   int64               `json:"last_hour_cs_time"`
	LastDayCSTime    int64               `json:"last_day_cs_time"`
	LastMinutePrice  sdk.Dec             `json:"last_minute_price"`
	LastHourPrice    sdk.Dec             `json:"last_hour_price"`
	LastDayPrice     sdk.Dec             `json:"last_day_price"`
	Market           string              `json:"Market"`
}

func NewCandleStickRecord(market string) *CandleStickRecord {
	res := &CandleStickRecord{Market: market, LastUpdateTime: time.Unix(0, 0)}
	for i := 0; i < len(res.MinuteCS); i++ {
		res.MinuteCS[i] = defaultBaseCandleStick()
	}
	for i := 0; i < len(res.HourCS); i++ {
		res.HourCS[i] = defaultBaseCandleStick()
	}
	res.LastMinutePrice = sdk.ZeroDec()
	res.LastHourPrice = sdk.ZeroDec()
	res.LastDayPrice = sdk.ZeroDec()
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

func GetSpanFromSpanStr(spanStr string) byte {
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
func (csr *CandleStickRecord) newBlock(isNewDay, isNewHour, isNewMinute bool, t time.Time) []CandleStick {
	res := make([]CandleStick, 0, 3)
	lastTime := csr.LastUpdateTime.Unix()
	if isNewMinute && lastTime != 0 {
		cs := csr.newCandleStick(csr.MinuteCS[csr.LastUpdateTime.Minute()], t.Unix(), Minute)
		if !cs.TotalDeal.IsZero() && csr.LastUpdateTime.Unix() != csr.LastMinuteCSTime {
			res = append(res, cs)
			csr.LastMinuteCSTime = csr.LastUpdateTime.Unix()
			csr.LastMinutePrice = cs.ClosePrice
		} else {
			res = append(res, CandleStick{
				OpenPrice:      csr.LastMinutePrice,
				ClosePrice:     csr.LastMinutePrice,
				HighPrice:      csr.LastMinutePrice,
				LowPrice:       csr.LastMinutePrice,
				TotalDeal:      sdk.ZeroInt(),
				EndingUnixTime: t.Unix(),
				TimeSpan:       MinuteStr,
				Market:         csr.Market,
			})
		}
	}
	if isNewHour && lastTime != 0 {
		csr.HourCS[csr.LastUpdateTime.Hour()] = merge(csr.MinuteCS[:])
		cs := csr.newCandleStick(csr.HourCS[csr.LastUpdateTime.Hour()], t.Unix(), Hour)
		if !cs.TotalDeal.IsZero() && csr.LastUpdateTime.Unix() != csr.LastHourCSTime {
			res = append(res, cs)
			csr.LastHourCSTime = csr.LastUpdateTime.Unix()
			csr.LastHourPrice = cs.ClosePrice
		} else {
			res = append(res, CandleStick{
				OpenPrice:      csr.LastHourPrice,
				ClosePrice:     csr.LastHourPrice,
				HighPrice:      csr.LastHourPrice,
				LowPrice:       csr.LastHourPrice,
				TotalDeal:      sdk.ZeroInt(),
				EndingUnixTime: t.Unix(),
				TimeSpan:       HourStr,
				Market:         csr.Market,
			})
		}
	}
	if isNewDay && lastTime != 0 {
		dayCS := merge(csr.HourCS[:])
		cs := csr.newCandleStick(dayCS, t.Unix(), Day)
		if !cs.TotalDeal.IsZero() && csr.LastUpdateTime.Unix() != csr.LastDayCSTime {
			res = append(res, cs)
			csr.LastDayCSTime = csr.LastUpdateTime.Unix()
			csr.LastDayPrice = cs.ClosePrice
		} else {
			res = append(res, CandleStick{
				OpenPrice:      csr.LastDayPrice,
				ClosePrice:     csr.LastDayPrice,
				HighPrice:      csr.LastDayPrice,
				LowPrice:       csr.LastDayPrice,
				TotalDeal:      sdk.ZeroInt(),
				EndingUnixTime: t.Unix(),
				TimeSpan:       DayStr,
				Market:         csr.Market,
			})
		}
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
		csSlice := csr.newBlock(isNewDay, isNewHour, isNewMinute, manager.LastBlockTime)
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
	ppMap      *treemap.Map //map[string]*PricePoint
	Updated    map[string]*PricePoint
	Side       string
	Levels     []string
	MulDecs    []sdk.Dec
	LevelDepth map[string]map[string]*PricePoint
}

func (dm *DepthManager) AddLevel(levelStr string) error {
	for _, lvl := range dm.Levels {
		//Return if already added
		if lvl == levelStr {
			return nil
		}
	}
	p, err := sdk.NewDecFromStr(levelStr)
	if err != nil {
		return err
	}
	dm.MulDecs = append(dm.MulDecs, sdk.OneDec().QuoTruncate(p))
	dm.Levels = append(dm.Levels, levelStr)
	dm.LevelDepth[levelStr] = make(map[string]*PricePoint)
	return nil
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

func DefaultDepthManager(side string) *DepthManager {
	dm := &DepthManager{
		ppMap:      treemap.NewWithStringComparator(),
		Updated:    make(map[string]*PricePoint),
		Side:       side,
		Levels:     make([]string, 0, 20),
		MulDecs:    make([]sdk.Dec, 0, 20),
		LevelDepth: make(map[string]map[string]*PricePoint, 20),
	}
	return dm
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
	//tmp := pp.Amount
	pp.Amount = pp.Amount.Add(amount)
	if pp.Amount.IsZero() {
		dm.ppMap.Remove(s)
	} else {
		dm.ppMap.Put(s, pp)
	}
	//if "8.800000000000000000" == price.String() && "buy" == dm.Side {
	//	fmt.Printf("== %s Price: %s amount: %s %s => %s\n", dm.Side, price.String(), amount.String(), tmp.String(), pp.Amount.String())
	//}
	dm.Updated[s] = pp

	for i, lev := range dm.Levels {
		updateAmount(dm.LevelDepth[lev], &PricePoint{Price: price, Amount: amount}, dm.MulDecs[i])
	}
}

// returns the changed PricePoints of last block. Clear dm.Updated for the next block
func (dm *DepthManager) EndBlock() (map[string]*PricePoint, map[string]map[string]*PricePoint) {
	oldUpdated, oldLevelDepth := dm.Updated, dm.LevelDepth
	dm.Updated = make(map[string]*PricePoint)
	dm.LevelDepth = make(map[string]map[string]*PricePoint, len(dm.Levels))
	for _, lev := range dm.Levels {
		dm.LevelDepth[lev] = make(map[string]*PricePoint)
	}
	return oldUpdated, oldLevelDepth
}

func updateAmount(m map[string]*PricePoint, point *PricePoint, mulDec sdk.Dec) {
	price := point.Price.Mul(mulDec).TruncateDec()
	if mulDec.GT(sdk.ZeroDec()) {
		price = price.Quo(mulDec)
	} else if mulDec.LT(sdk.ZeroDec()) {
		price = price.Mul(mulDec)
	}
	s := string(DecToBigEndianBytes(price))
	if val, ok := m[s]; ok {
		val.Amount = val.Amount.Add(point.Amount)
	} else {
		m[s] = &PricePoint{
			Price:  price,
			Amount: point.Amount,
		}
	}
}

func mergePrice(updated []*PricePoint, level string) map[string]*PricePoint {
	p, err := sdk.NewDecFromStr(level)
	if err != nil {
		return nil
	}
	mulDec := sdk.OneDec().QuoTruncate(p)
	depth := make(map[string]*PricePoint)
	for _, point := range updated {
		updateAmount(depth, point, mulDec)
	}
	return depth
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
	PriceList   [MinuteNumInDay]sdk.Dec `json:"price_list"`
	Price1st    sdk.Dec                 `json:"price_1st"`
	Minute1st   int                     `json:"minute_1st"`
	Price2nd    sdk.Dec                 `json:"price_2nd"`
	Minute2nd   int                     `json:"minute_2nd"`
	Market      string                  `json:"market"`
	Initialized bool                    `json:"initialized"`
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
		tm.Price1st = currPrice
		tm.Minute1st = currMinute
		tm.Price2nd = currPrice
		tm.Minute2nd = currMinute
		return
	}
	tm.PriceList[tm.Minute2nd] = tm.Price2nd
	for {
		tm.Minute2nd++
		if tm.Minute2nd >= MinuteNumInDay {
			tm.Minute2nd = 0
		}
		if tm.Minute2nd == tm.Minute1st {
			break
		}
		tm.PriceList[tm.Minute2nd] = tm.Price2nd
	}
	tm.Price2nd = tm.Price1st
	tm.Minute2nd = tm.Minute1st
	tm.Price1st = currPrice
	tm.Minute1st = currMinute
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
	if (tm.Minute1st == currMinute && !tm.Price1st.Equal(tm.Price2nd)) ||
		!tm.PriceList[currMinute].Equal(tm.PriceList[lastMinute]) {
		return &Ticker{
			NewPrice:          tm.Price1st,
			OldPriceOneDayAgo: tm.PriceList[currMinute],
			Market:            tm.Market,
			MinuteInDay:       currMinute,
		}
	}
	if tm.Minute1st != currMinute { //flush the price
		tm.UpdateNewestPrice(tm.Price1st, currMinute)
	}
	return nil
}

//=============================

type XTickerManager struct {
	PriceList         [MinuteNumInDay]sdk.Dec `json:"price_list"`
	AmountList        [MinuteNumInDay]sdk.Int `json:"amount_list"`
	Highest           sdk.Dec                 `json:"highest"`
	Lowest            sdk.Dec                 `json:"lowest"`
	Market            string                  `json:"market"`
	Initialized       bool                    `json:"initialized"`
	TotalAmount       sdk.Int                 `json:"amount"`
	LastUpdate        int                     `json:"amount_last_update"`
	OldPriceOneDayAgo sdk.Dec                 `json:"old_price"`
}

func DefaultXTickerManager(Market string) *XTickerManager {
	return &XTickerManager{
		Market:      Market,
		TotalAmount: sdk.ZeroInt(),
	}
}

func (tm *XTickerManager) updateHighest() {
	tm.Highest = sdk.ZeroDec()
	for i := 0; i < MinuteNumInDay; i++ {
		price := tm.PriceList[i]
		if !tm.AmountList[i].IsZero() && price.GT(tm.Highest) {
			tm.Highest = price
		}
	}
}
func (tm *XTickerManager) updateLowest() {
	tm.Lowest = sdk.ZeroDec()
	for i := 0; i < MinuteNumInDay; i++ {
		price := tm.PriceList[i]
		if !tm.AmountList[i].IsZero() && (price.LT(tm.Lowest) || tm.Lowest.IsZero()) {
			tm.Lowest = price
		}
	}
}

func (tm *XTickerManager) setPrice(i int, price sdk.Dec, amount sdk.Int, highestMayChange *bool, lowestMayChange *bool) {
	*highestMayChange = *highestMayChange || (tm.PriceList[i].Equal(tm.Highest) && !tm.AmountList[i].IsZero())
	*lowestMayChange = *lowestMayChange || (tm.PriceList[i].Equal(tm.Lowest) && !tm.AmountList[i].IsZero())
	tm.PriceList[i] = price
	if !amount.IsZero() && price.GT(tm.Highest) {
		tm.Highest = price
	}
	if !amount.IsZero() && (price.LT(tm.Lowest) || tm.Lowest.IsZero()) {
		tm.Lowest = price
	}
}

func (tm *XTickerManager) UpdateNewestPrice(currPrice sdk.Dec, currMinute int, currAmount sdk.Int) {
	if currMinute >= MinuteNumInDay || currMinute < 0 {
		panic("Minute not valid")
	}
	if !tm.Initialized {
		tm.Initialized = true
		for i := 0; i < MinuteNumInDay; i++ {
			tm.PriceList[i] = currPrice
			tm.AmountList[i] = sdk.ZeroInt()
		}
		tm.Highest = currPrice
		tm.Lowest = currPrice
		tm.TotalAmount = currAmount
		//fmt.Printf("Here! %d %s %s\n", currMinute, tm.Market, tm.TotalAmount)
		tm.AmountList[currMinute] = currAmount
		tm.LastUpdate = currMinute
		return
	}

	tm.OldPriceOneDayAgo = tm.PriceList[currMinute]
	lastPrice := tm.PriceList[tm.LastUpdate]
	highestMayChange := false
	lowestMayChange := false
	for {
		tm.LastUpdate++
		if tm.LastUpdate >= MinuteNumInDay {
			tm.LastUpdate = 0
		}
		tm.TotalAmount = tm.TotalAmount.Sub(tm.AmountList[tm.LastUpdate])
		//fmt.Printf("Here Sub! %d %s %s\n", tm.LastUpdate, tm.AmountList[tm.LastUpdate], tm.TotalAmount)
		if tm.LastUpdate == currMinute {
			tm.setPrice(tm.LastUpdate, currPrice, currAmount, &highestMayChange, &lowestMayChange)
			tm.AmountList[tm.LastUpdate] = currAmount
			break
		} else {
			tm.setPrice(tm.LastUpdate, lastPrice, currAmount, &highestMayChange, &lowestMayChange)
			tm.AmountList[tm.LastUpdate] = sdk.ZeroInt()
		}
	}
	if highestMayChange {
		tm.updateHighest()
	}
	if lowestMayChange {
		tm.updateLowest()
	}
	tm.TotalAmount = tm.TotalAmount.Add(currAmount)
}

func (tm *XTickerManager) GetXTicker(currMinute int) *XTicker {
	if !tm.Initialized {
		return nil
	}
	if currMinute != tm.LastUpdate {
		tm.UpdateNewestPrice(tm.PriceList[tm.LastUpdate], currMinute, sdk.ZeroInt())
	}
	//fmt.Printf("Here Sicker! %s %s\n", tm.Market, tm.TotalAmount)
	return &XTicker{
		Market:            tm.Market,
		MinuteInDay:       tm.LastUpdate,
		NewPrice:          tm.PriceList[tm.LastUpdate],
		OldPriceOneDayAgo: tm.OldPriceOneDayAgo,
		HighPrice:         tm.Highest,
		LowPrice:          tm.Lowest,
		TotalDeal:         tm.TotalAmount,
	}
}
