package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestCandleStickRecord_Update(t *testing.T) {
	record := NewCandleStickRecord("abc/cet")
	now := time.Now()
	record.Update(now, sdk.NewDec(125), 526)
	require.EqualValues(t, now, record.LastUpdateTime)

	baseCandlerStick := record.MinuteCS[now.UTC().Minute()]
	require.True(t, baseCandlerStick.hasDeal())
	require.EqualValues(t, 526, baseCandlerStick.TotalDeal.Int64())
	require.EqualValues(t, sdk.NewDec(125), baseCandlerStick.HighPrice)
	require.EqualValues(t, sdk.NewDec(125), baseCandlerStick.OpenPrice)
	require.EqualValues(t, sdk.NewDec(125), baseCandlerStick.LowPrice)
	require.EqualValues(t, sdk.NewDec(125), baseCandlerStick.ClosePrice)

}

func TestGetMinuteCandleStick(t *testing.T) {
	record := setMinuteRecord()

	mcs := record.getMinuteCandleStick(true, record.LastUpdateTime.UTC().Unix()-100)
	cs := record.MinuteCS[record.LastUpdateTime.UTC().Minute()]
	require.EqualValues(t, cs.OpenPrice, mcs.OpenPrice)
	require.EqualValues(t, cs.ClosePrice, mcs.ClosePrice)
	require.EqualValues(t, cs.HighPrice, mcs.HighPrice)
	require.EqualValues(t, cs.LowPrice, mcs.LowPrice)
	require.EqualValues(t, cs.TotalDeal, mcs.TotalDeal)
	require.EqualValues(t, record.LastUpdateTime.UTC().Unix(), mcs.EndingUnixTime)

	record.LastMinuteCSTime = record.LastUpdateTime.UTC().Unix()
	lastBlockTime := record.LastUpdateTime.UTC().Unix() - 1000
	mcs = record.getMinuteCandleStick(true, lastBlockTime)
	require.EqualValues(t, record.LastMinutePrice, mcs.OpenPrice)
	require.EqualValues(t, record.LastMinutePrice, mcs.ClosePrice)
	require.EqualValues(t, record.LastMinutePrice, mcs.HighPrice)
	require.EqualValues(t, record.LastMinutePrice, mcs.LowPrice)
	require.EqualValues(t, 0, mcs.TotalDeal.Int64())
	require.EqualValues(t, lastBlockTime, mcs.EndingUnixTime)

}

func setMinuteRecord() *CandleStickRecord {
	now := time.Now()
	record := NewCandleStickRecord("abc/cet")
	record.LastUpdateTime = now
	record.LastMinutePrice = sdk.NewDec(90)
	record.LastMinuteCSTime = now.UTC().Unix() - 10000
	record.MinuteCS[record.LastUpdateTime.UTC().Minute()] = baseCandleStick{
		OpenPrice:  sdk.NewDec(69),
		ClosePrice: sdk.NewDec(96),
		HighPrice:  sdk.NewDec(96),
		LowPrice:   sdk.NewDec(69),
		TotalDeal:  sdk.NewInt(9600),
	}
	return record
}

func TestGetHourCandleStick(t *testing.T) {
	record := setHourRecord()
	lastBlockTime := record.LastUpdateTime.UTC().Unix() - 2000
	hcs := record.getHourCandleStick(true, lastBlockTime)
	require.EqualValues(t, 2, hcs.OpenPrice.RoundInt64())
	require.EqualValues(t, 2*59, hcs.ClosePrice.RoundInt64())
	require.EqualValues(t, 2*59, hcs.HighPrice.RoundInt64())
	require.EqualValues(t, 2, hcs.LowPrice.RoundInt64())
	require.EqualValues(t, record.LastHourPrice, hcs.ClosePrice)

	record.LastHourCSTime = record.LastUpdateTime.UTC().Unix()
	hcs = record.getHourCandleStick(true, lastBlockTime)
	require.EqualValues(t, record.LastHourPrice, hcs.OpenPrice)
	require.EqualValues(t, record.LastHourPrice, hcs.ClosePrice)
	require.EqualValues(t, record.LastHourPrice, hcs.HighPrice)
	require.EqualValues(t, record.LastHourPrice, hcs.LowPrice)
	require.EqualValues(t, 0, hcs.TotalDeal.Int64())
}

func setHourRecord() *CandleStickRecord {
	now := time.Now()
	record := NewCandleStickRecord("abc/cet")
	for i := int64(0); i < 60; i++ {
		record.MinuteCS[i] = baseCandleStick{
			OpenPrice:  sdk.NewDec(i * 2),
			ClosePrice: sdk.NewDec(i * 2),
			HighPrice:  sdk.NewDec(i * 2),
			LowPrice:   sdk.NewDec(i * 2),
			TotalDeal:  sdk.NewInt(i * 100),
		}
	}
	record.LastUpdateTime = now
	record.LastHourCSTime = now.UTC().Unix() - 1000
	record.LastHourPrice = sdk.NewDec(999)
	return record
}

func TestGetDayCandleStick(t *testing.T) {
	record := setDayRecord()
	lastBlockTime := record.LastUpdateTime.UTC().Unix() - 2000
	dcs := record.getDayCandleStick(true, lastBlockTime)
	require.EqualValues(t, 2, dcs.OpenPrice.RoundInt64())
	require.EqualValues(t, 2*23, dcs.ClosePrice.RoundInt64())
	require.EqualValues(t, 2*23, dcs.HighPrice.RoundInt64())
	require.EqualValues(t, 2, dcs.LowPrice.RoundInt64())
	require.EqualValues(t, record.LastDayPrice, dcs.ClosePrice)

	record.LastDayCSTime = record.LastUpdateTime.UTC().Unix()
	dcs = record.getDayCandleStick(true, lastBlockTime)
	require.EqualValues(t, record.LastDayPrice, dcs.OpenPrice)
	require.EqualValues(t, record.LastDayPrice, dcs.ClosePrice)
	require.EqualValues(t, record.LastDayPrice, dcs.HighPrice)
	require.EqualValues(t, record.LastDayPrice, dcs.LowPrice)
	require.EqualValues(t, 0, dcs.TotalDeal.Int64())
}

func setDayRecord() *CandleStickRecord {
	now := time.Now()
	record := NewCandleStickRecord("abc/cet")
	for i := int64(0); i < 24; i++ {
		record.HourCS[i] = baseCandleStick{
			OpenPrice:  sdk.NewDec(i * 2),
			ClosePrice: sdk.NewDec(i * 2),
			HighPrice:  sdk.NewDec(i * 2),
			LowPrice:   sdk.NewDec(i * 2),
			TotalDeal:  sdk.NewInt(i * 100),
		}
	}
	record.LastUpdateTime = now
	record.LastDayCSTime = now.UTC().Unix() - 1000
	record.LastDayPrice = sdk.NewDec(999)
	return record
}
