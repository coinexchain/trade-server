package core

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestBaseCandleStick(t *testing.T) {
	cs := defaultBaseCandleStick()
	//test update
	cs.update(sdk.NewDec(50), 2)
	cs.update(sdk.NewDec(10), 2)
	require.Equal(t, baseCandleStick{
		OpenPrice:  sdk.NewDec(50),
		ClosePrice: sdk.NewDec(10),
		HighPrice:  sdk.NewDec(50),
		LowPrice:   sdk.NewDec(10),
		TotalDeal:  sdk.NewInt(4),
	}, cs)
	cs.update(sdk.NewDec(60), 1)
	cs.update(sdk.NewDec(20), 1)
	require.Equal(t, baseCandleStick{
		OpenPrice:  sdk.NewDec(50),
		ClosePrice: sdk.NewDec(20),
		HighPrice:  sdk.NewDec(60),
		LowPrice:   sdk.NewDec(10),
		TotalDeal:  sdk.NewInt(6),
	}, cs)
	cs.update(sdk.NewDec(5), 10)
	require.Equal(t, baseCandleStick{
		OpenPrice:  sdk.NewDec(50),
		ClosePrice: sdk.NewDec(5),
		HighPrice:  sdk.NewDec(60),
		LowPrice:   sdk.NewDec(5),
		TotalDeal:  sdk.NewInt(16),
	}, cs)
	//test merge
	cs1 := baseCandleStick{
		OpenPrice:  sdk.NewDec(10),
		ClosePrice: sdk.NewDec(50),
		HighPrice:  sdk.NewDec(50),
		LowPrice:   sdk.NewDec(10),
		TotalDeal:  sdk.NewInt(4),
	}
	cs2 := baseCandleStick{
		OpenPrice:  sdk.NewDec(50),
		ClosePrice: sdk.NewDec(100),
		HighPrice:  sdk.NewDec(200),
		LowPrice:   sdk.NewDec(5),
		TotalDeal:  sdk.NewInt(4),
	}
	cs3 := baseCandleStick{
		OpenPrice:  sdk.NewDec(50),
		ClosePrice: sdk.NewDec(100),
		HighPrice:  sdk.NewDec(200),
		LowPrice:   sdk.NewDec(5),
		TotalDeal:  sdk.NewInt(0),
	}
	cs = merge([]baseCandleStick{cs1, cs2, cs3})
	require.Equal(t, baseCandleStick{
		OpenPrice:  sdk.NewDec(10),
		ClosePrice: sdk.NewDec(100),
		HighPrice:  sdk.NewDec(200),
		LowPrice:   sdk.NewDec(5),
		TotalDeal:  sdk.NewInt(8),
	}, cs)
}

var lastTime time.Time

func T(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	lastTime = t
	return t
}

func TestCandleStickRecord(t *testing.T) {
	market1 := "abc/cet"
	csrMan := NewCandleStickManager([]string{"cet/usdt"})
	csrMan.AddMarket(market1)
	csSlice := csrMan.NewBlock(T("2019-07-15T08:07:10Z"))
	require.Equal(t, 0, len(csSlice))

	csSlice = csrMan.NewBlock(T("2019-07-15T08:08:10Z"))
	require.Equal(t, 0, len(csSlice))
	csrMan.GetRecord(market1).Update(lastTime, sdk.NewDec(50), 2)

	csSlice = csrMan.NewBlock(T("2019-07-15T08:08:40Z"))
	require.Equal(t, 0, len(csSlice))
	csrMan.GetRecord(market1).Update(lastTime, sdk.NewDec(10), 2)
	lastTimeOld := lastTime

	csSlice = csrMan.NewBlock(T("2019-07-15T08:09:10Z"))
	require.Equal(t, []CandleStick{
		{
			OpenPrice:      sdk.NewDec(50),
			ClosePrice:     sdk.NewDec(10),
			HighPrice:      sdk.NewDec(50),
			LowPrice:       sdk.NewDec(10),
			TotalDeal:      sdk.NewInt(4),
			EndingUnixTime: lastTimeOld.Unix(),
			TimeSpan:       Minute,
			Market:         market1,
		},
	}, csSlice)

	csSlice = csrMan.NewBlock(T("2019-07-15T08:09:10Z"))
	require.Equal(t, 0, len(csSlice))
	csrMan.GetRecord(market1).Update(lastTime, sdk.NewDec(10), 2)

	csSlice = csrMan.NewBlock(T("2019-07-15T08:09:20Z"))
	require.Equal(t, 0, len(csSlice))
	csrMan.GetRecord(market1).Update(lastTime, sdk.NewDec(40), 2)

	csSlice = csrMan.NewBlock(T("2019-07-15T08:09:40Z"))
	require.Equal(t, 0, len(csSlice))
	csrMan.GetRecord(market1).Update(lastTime, sdk.NewDec(20), 2)
	lastTimeOld = lastTime

	csSlice = csrMan.NewBlock(T("2019-07-15T08:10:40Z"))
	require.Equal(t, []CandleStick{
		{
			OpenPrice:      sdk.NewDec(10),
			ClosePrice:     sdk.NewDec(20),
			HighPrice:      sdk.NewDec(40),
			LowPrice:       sdk.NewDec(10),
			TotalDeal:      sdk.NewInt(6),
			EndingUnixTime: lastTimeOld.Unix(),
			TimeSpan:       Minute,
			Market:         market1,
		},
	}, csSlice)

	csSlice = csrMan.NewBlock(T("2019-07-15T08:10:40Z"))
	require.Equal(t, 0, len(csSlice))
	csrMan.GetRecord(market1).Update(lastTime, sdk.NewDec(80), 10)
	lastTimeOld = lastTime

	csSlice = csrMan.NewBlock(T("2019-07-15T09:00:40Z"))
	require.Equal(t, []CandleStick{
		{
			OpenPrice:      sdk.NewDec(80),
			ClosePrice:     sdk.NewDec(80),
			HighPrice:      sdk.NewDec(80),
			LowPrice:       sdk.NewDec(80),
			TotalDeal:      sdk.NewInt(10),
			EndingUnixTime: lastTimeOld.Unix(),
			TimeSpan:       Minute,
			Market:         market1,
		},
		{
			OpenPrice:      sdk.NewDec(50),
			ClosePrice:     sdk.NewDec(80),
			HighPrice:      sdk.NewDec(80),
			LowPrice:       sdk.NewDec(10),
			TotalDeal:      sdk.NewInt(20),
			EndingUnixTime: lastTimeOld.Unix(),
			TimeSpan:       Hour,
			Market:         market1,
		},
	}, csSlice)
}

func TestDepthManager(t *testing.T) {
	dm := DefaultDepthManager()
	dm.DeltaChange(sdk.NewDec(90), sdk.NewInt(90))
	dm.DeltaChange(sdk.NewDec(100), sdk.NewInt(100))
	dm.DeltaChange(sdk.NewDec(80), sdk.NewInt(80))
	dm.DeltaChange(sdk.NewDec(10), sdk.NewInt(10))
	points := dm.DumpPricePoints()
	require.Equal(t, []*PricePoint{
		{Price: sdk.NewDec(10), Amount: sdk.NewInt(10)},
		{Price: sdk.NewDec(80), Amount: sdk.NewInt(80)},
		{Price: sdk.NewDec(90), Amount: sdk.NewInt(90)},
		{Price: sdk.NewDec(100), Amount: sdk.NewInt(100)},
	}, points)
	points = dm.GetLowest(2)
	require.Equal(t, []*PricePoint{
		{Price: sdk.NewDec(10), Amount: sdk.NewInt(10)},
		{Price: sdk.NewDec(80), Amount: sdk.NewInt(80)},
	}, points)
	points = dm.GetHighest(2)
	require.Equal(t, []*PricePoint{
		{Price: sdk.NewDec(100), Amount: sdk.NewInt(100)},
		{Price: sdk.NewDec(90), Amount: sdk.NewInt(90)},
	}, points)
	dm.EndBlock()
	dm.DeltaChange(sdk.NewDec(80), sdk.NewInt(-10))

	ppMap := dm.EndBlock()
	for pp, _ := range ppMap {
		require.Equal(t, &PricePoint{sdk.NewDec(80), sdk.NewInt(70)}, pp)
	}
}
