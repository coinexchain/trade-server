package core

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
		cs := csr.newCandleStick(csr.MinuteCS[csr.LastUpdateTime.UTC().Minute()], t.Unix(), Minute)
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
		gotResult := false
		if csr.LastUpdateTime.Unix() != csr.LastHourCSTime {
			csr.HourCS[csr.LastUpdateTime.UTC().Hour()] = merge(csr.MinuteCS[:])
			cs := csr.newCandleStick(csr.HourCS[csr.LastUpdateTime.UTC().Hour()], t.Unix(), Hour)
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
				EndingUnixTime: t.Unix(),
				TimeSpan:       HourStr,
				Market:         csr.Market,
			})
		}
	}
	if isNewDay && lastTime != 0 {
		gotResult := false
		if csr.LastUpdateTime.Unix() != csr.LastDayCSTime {
			dayCS := merge(csr.HourCS[:])
			cs := csr.newCandleStick(dayCS, t.Unix(), Day)
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
				EndingUnixTime: t.Unix(),
				TimeSpan:       DayStr,
				Market:         csr.Market,
			})
		}
	}

	if isNewDay {
		for i := 0; i < 24; i++ {
			csr.HourCS[i] = newBaseCandleStick()
		}
	}
	if isNewDay || isNewHour {
		for i := 0; i < 60; i++ {
			csr.MinuteCS[i] = newBaseCandleStick()
		}
	}
	return res
}

func (csr *CandleStickRecord) Update(t time.Time, price sdk.Dec, amount int64) {
	csr.MinuteCS[t.UTC().Minute()].update(price, amount)
	csr.LastUpdateTime = t
}
