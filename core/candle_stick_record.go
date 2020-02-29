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
	MinuteCS         [60]baseCandleStick `json:"minute_cs"`
	HourCS           [24]baseCandleStick `json:"hour_cs"`
	LastUpdateTime   time.Time           `json:"last_update"`         //the last time that 'Update' was invoked
	//the last time a non-empty minute/hour/day candle stick was "flushed out"
	//A candle stick is NOT "flushed out" UNTIL a new minute/hour/day comes
	LastMinuteCSTime int64               `json:"last_minute_cs_time"`
	LastHourCSTime   int64               `json:"last_hour_cs_time"`
	LastDayCSTime    int64               `json:"last_day_cs_time"`
	//the last price of a non-empty minute/hour/day candle stick
	LastMinutePrice  sdk.Dec             `json:"last_minute_price"`
	LastHourPrice    sdk.Dec             `json:"last_hour_price"`
	LastDayPrice     sdk.Dec             `json:"last_day_price"`
	Market           string              `json:"Market"`
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

// When a new block comes, flush the pending candle sticks
func (csr *CandleStickRecord) newBlock(isNewDay, isNewHour, isNewMinute bool, endTime time.Time) []CandleStick {
	res := make([]CandleStick, 0, 3)
	lastTime := csr.LastUpdateTime.Unix()
	if isNewMinute && lastTime != 0 {
		cs := csr.newCandleStick(csr.MinuteCS[csr.LastUpdateTime.UTC().Minute()], csr.LastUpdateTime.Unix(), Minute)
		if !cs.TotalDeal.IsZero() && /* the function Update has changed it, so the deal is not zero */
			csr.LastUpdateTime.Unix() != csr.LastMinuteCSTime /*Has new updates after last minute-candle-stick*/ {
			// generate new candle stick and record its time
			res = append(res, cs)
			csr.LastMinuteCSTime = csr.LastUpdateTime.Unix()
			csr.LastMinutePrice = cs.ClosePrice
		} else {
			res = append(res, CandleStick{ //generate an empty candle stick
				OpenPrice:      csr.LastMinutePrice,
				ClosePrice:     csr.LastMinutePrice,
				HighPrice:      csr.LastMinutePrice,
				LowPrice:       csr.LastMinutePrice,
				TotalDeal:      sdk.ZeroInt(),
				EndingUnixTime: endTime.Unix(),
				TimeSpan:       MinuteStr,
				Market:         csr.Market,
			})
		}
	}
	if isNewHour && lastTime != 0 {
		gotResult := false
		if csr.LastUpdateTime.Unix() != csr.LastHourCSTime { //Has new updates after last hour-candle-stick
			// merge minute-candle-sticks to hour-candle-sticks
			csr.HourCS[csr.LastUpdateTime.UTC().Hour()] = merge(csr.MinuteCS[:])
			cs := csr.newCandleStick(csr.HourCS[csr.LastUpdateTime.UTC().Hour()], csr.LastUpdateTime.Unix(), Hour)
			if !cs.TotalDeal.IsZero() {
				res = append(res, cs)
				csr.LastHourCSTime = csr.LastUpdateTime.Unix()
				csr.LastHourPrice = cs.ClosePrice
				gotResult = true
			}
		}
		if !gotResult {
			res = append(res, CandleStick{
				OpenPrice:      csr.LastHourPrice,
				ClosePrice:     csr.LastHourPrice,
				HighPrice:      csr.LastHourPrice,
				LowPrice:       csr.LastHourPrice,
				TotalDeal:      sdk.ZeroInt(),
				EndingUnixTime: endTime.Unix(),
				TimeSpan:       HourStr,
				Market:         csr.Market,
			})
		}
	}
	if isNewDay && lastTime != 0 {
		gotResult := false
		if csr.LastUpdateTime.Unix() != csr.LastDayCSTime { //Has new updates after last day-candle-stick
			// merge hour-candle-sticks to day-candle-sticks
			dayCS := merge(csr.HourCS[:])
			cs := csr.newCandleStick(dayCS, csr.LastUpdateTime.Unix(), Day)
			if !cs.TotalDeal.IsZero() {
				res = append(res, cs)
				csr.LastDayCSTime = csr.LastUpdateTime.Unix()
				csr.LastDayPrice = cs.ClosePrice
				gotResult = true
			}
		}
		if !gotResult {
			res = append(res, CandleStick{
				OpenPrice:      csr.LastDayPrice,
				ClosePrice:     csr.LastDayPrice,
				HighPrice:      csr.LastDayPrice,
				LowPrice:       csr.LastDayPrice,
				TotalDeal:      sdk.ZeroInt(),
				EndingUnixTime: endTime.Unix(),
				TimeSpan:       DayStr,
				Market:         csr.Market,
			})
		}
	}

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
	return res
}

func (csr *CandleStickRecord) newCandleStick(cs baseCandleStick, endTime int64, span byte) CandleStick {
	return CandleStick{
		OpenPrice:      cs.OpenPrice,
		ClosePrice:     cs.ClosePrice,
		HighPrice:      cs.HighPrice,
		LowPrice:       cs.LowPrice,
		TotalDeal:      cs.TotalDeal,
		EndingUnixTime: endTime,
		TimeSpan:       getSpanStrFromSpan(span),
		Market:         csr.Market,
	}
}

// Update when there is a new deal
func (csr *CandleStickRecord) Update(t time.Time, price sdk.Dec, amount int64) {
	csr.MinuteCS[t.UTC().Minute()].update(price, amount)
	csr.LastUpdateTime = t
}
