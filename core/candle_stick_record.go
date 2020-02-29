package core

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

/*
We generate the minute-candle-sticks in a "equipping and flushing" mode.
Within a minute M, the 'Update' function can be invoked multiple times and the minute-candle-stick of M is equipped
with more and more deals.
When a new minute comes, 'newBlock' functions will be invoked. Please note, this moment is not always M+1, instead,
it can be M+1, M+2 or any time later.
When 'newBlock' is invoked, we must flush out the minute-candle-stick which is getting equipped. Is there always
such a minute-candle-stick? Not always. In some cases, there has been no deals since last flush. The
moment of last flush is recorded in LastMinuteCSTime. The last time 'Update' is invoked is recorded in LastUpdateTime.
By comparing LastMinuteCSTime and LastUpdateTime, we'll know whether there is an equipping minute-candle-stick or not.
*/

// Record the candle sticks within one day
// It can provide three granularities: minute, hour and day.
// If the client side wants other granularities, we need to merge several candle sticks into one.
type CandleStickRecord struct {
	MinuteCS       [60]baseCandleStick `json:"minute_cs"`
	HourCS         [24]baseCandleStick `json:"hour_cs"`
	LastUpdateTime time.Time           `json:"last_update"` //the last time that 'Update' was invoked
	//the last time a non-empty minute/hour/day candle stick was "flushed out"
	//A candle stick is NOT "flushed out" UNTIL a new minute/hour/day comes
	LastMinuteCSTime int64 `json:"last_minute_cs_time"`
	LastHourCSTime   int64 `json:"last_hour_cs_time"`
	LastDayCSTime    int64 `json:"last_day_cs_time"`
	//the last price of a non-empty minute/hour/day candle stick
	LastMinutePrice sdk.Dec `json:"last_minute_price"`
	LastHourPrice   sdk.Dec `json:"last_hour_price"`
	LastDayPrice    sdk.Dec `json:"last_day_price"`
	Market          string  `json:"Market"`
}

func NewCandleStickRecord(market string) *CandleStickRecord {
	res := &CandleStickRecord{Market: market, LastUpdateTime: time.Unix(0, 0)}
	for i := 0; i < len(res.MinuteCS); i++ {
		res.MinuteCS[i] = newBaseCandleStick()
	}
	for i := 0; i < len(res.HourCS); i++ {
		res.HourCS[i] = newBaseCandleStick()
	}
	res.LastMinutePrice = sdk.ZeroDec()
	res.LastHourPrice = sdk.ZeroDec()
	res.LastDayPrice = sdk.ZeroDec()
	return res
}

// When a new block comes, gets the k-line data before the block time
func (csr *CandleStickRecord) newBlock(isNewDay, isNewHour, isNewMinute bool, lastBlockTime time.Time) []*CandleStick {
	res := make([]*CandleStick, 0, 3)
	lastBlockUnixTime := lastBlockTime.UTC().Unix()
	if mcs := csr.getMinuteCandleStick(isNewMinute, lastBlockUnixTime); mcs != nil {
		res = append(res, mcs)
	}
	if hcs := csr.getHourCandleStick(isNewHour, lastBlockUnixTime); hcs != nil {
		res = append(res, hcs)
	}
	if dcs := csr.getDayCandleStick(isNewDay, lastBlockUnixTime); dcs != nil {
		res = append(res, dcs)
	}

	csr.clearRecords(isNewHour, isNewDay)
	return res
}

func (csr *CandleStickRecord) getMinuteCandleStick(isNewMinute bool, lastBlockTime int64) *CandleStick {
	lastUpdateTime := csr.LastUpdateTime.UTC().Unix()
	if isNewMinute && lastUpdateTime != 0 {
		if lastUpdateTime != csr.LastMinuteCSTime {
			lastUpdateCandleStick := csr.MinuteCS[csr.LastUpdateTime.UTC().Minute()]
			cs := csr.newCandleStick(lastUpdateCandleStick, lastUpdateTime, Minute)
			if !cs.TotalDeal.IsZero() {
				csr.LastMinuteCSTime = cs.EndingUnixTime
				csr.LastMinutePrice = cs.ClosePrice
				return cs
			}
		}
		// When the volume is 0, fill the last price
		return &CandleStick{
			OpenPrice:      csr.LastMinutePrice,
			ClosePrice:     csr.LastMinutePrice,
			HighPrice:      csr.LastMinutePrice,
			LowPrice:       csr.LastMinutePrice,
			TotalDeal:      sdk.ZeroInt(),
			EndingUnixTime: lastBlockTime,
			TimeSpan:       MinuteStr,
			Market:         csr.Market,
		}
	}
	return nil
}

func (csr *CandleStickRecord) getHourCandleStick(isNewHour bool, lastBlockTime int64) *CandleStick {
	lastUpdateTime := csr.LastUpdateTime.UTC().Unix()
	if isNewHour && lastUpdateTime != 0 {
		if lastUpdateTime != csr.LastHourCSTime { //Has new updates after last hour-candle-stick
			// merge minute-candle-sticks to hour-candle-sticks
			hour := csr.LastUpdateTime.UTC().Hour()
			csr.HourCS[hour] = merge(csr.MinuteCS[:])
			cs := csr.newCandleStick(csr.HourCS[hour], lastUpdateTime, Hour)
			if !cs.TotalDeal.IsZero() {
				csr.LastHourCSTime = cs.EndingUnixTime
				csr.LastHourPrice = cs.ClosePrice
				return cs
			}

		}
		// When the volume is 0, fill the last price
		return &CandleStick{
			OpenPrice:      csr.LastHourPrice,
			ClosePrice:     csr.LastHourPrice,
			HighPrice:      csr.LastHourPrice,
			LowPrice:       csr.LastHourPrice,
			TotalDeal:      sdk.ZeroInt(),
			EndingUnixTime: lastBlockTime,
			TimeSpan:       HourStr,
			Market:         csr.Market,
		}
	}
	return nil
}

func (csr *CandleStickRecord) getDayCandleStick(isNewDay bool, lastBlockTime int64) *CandleStick {
	lastUpdateTime := csr.LastUpdateTime.UTC().Unix()
	if isNewDay && lastUpdateTime != 0 {
		if csr.LastUpdateTime.Unix() != csr.LastDayCSTime { //Has new updates after last day-candle-stick
			// merge hour-candle-sticks to day-candle-sticks
			dayCS := merge(csr.HourCS[:])
			cs := csr.newCandleStick(dayCS, lastUpdateTime, Day)
			if !cs.TotalDeal.IsZero() {
				csr.LastDayCSTime = cs.EndingUnixTime
				csr.LastDayPrice = cs.ClosePrice
				return cs
			}
		}
		// When the volume is 0, fill the last price
		return &CandleStick{
			OpenPrice:      csr.LastDayPrice,
			ClosePrice:     csr.LastDayPrice,
			HighPrice:      csr.LastDayPrice,
			LowPrice:       csr.LastDayPrice,
			TotalDeal:      sdk.ZeroInt(),
			EndingUnixTime: lastBlockTime,
			TimeSpan:       DayStr,
			Market:         csr.Market,
		}
	}
	return nil
}

func (csr *CandleStickRecord) clearRecords(isNewHour, isNewDay bool) {
	if isNewDay {
		for i := 0; i < 24; i++ {
			// clear the hour-candle-sticks of last day
			csr.HourCS[i] = newBaseCandleStick()
		}
	}
	if isNewDay || isNewHour {
		for i := 0; i < 60; i++ {
			// clear the minute-candle-sticks of last hour
			csr.MinuteCS[i] = newBaseCandleStick()
		}
	}
}

func (csr *CandleStickRecord) newCandleStick(updateCS baseCandleStick, updateTime int64, span byte) *CandleStick {
	return &CandleStick{
		OpenPrice:      updateCS.OpenPrice,
		ClosePrice:     updateCS.ClosePrice,
		HighPrice:      updateCS.HighPrice,
		LowPrice:       updateCS.LowPrice,
		TotalDeal:      updateCS.TotalDeal,
		EndingUnixTime: updateTime,
		TimeSpan:       getSpanStrFromSpan(span),
		Market:         csr.Market,
	}
}

// Update when there is a new deal
func (csr *CandleStickRecord) Update(t time.Time, price sdk.Dec, amount int64) {
	csr.MinuteCS[t.UTC().Minute()].update(price, amount)
	csr.LastUpdateTime = t
}
