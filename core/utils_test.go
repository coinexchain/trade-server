package core

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestBaseCandleStick(t *testing.T) {
	cs := newBaseCandleStick()
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
	require.Equal(t, CandleStick{
		OpenPrice:      sdk.NewDec(50),
		ClosePrice:     sdk.NewDec(10),
		HighPrice:      sdk.NewDec(50),
		LowPrice:       sdk.NewDec(10),
		TotalDeal:      sdk.NewInt(4),
		EndingUnixTime: lastTimeOld.Unix(),
		TimeSpan:       getSpanStrFromSpan(Minute),
		Market:         market1,
	}, *(csSlice[0]))

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
	require.Equal(t, CandleStick{
		OpenPrice:      sdk.NewDec(10),
		ClosePrice:     sdk.NewDec(20),
		HighPrice:      sdk.NewDec(40),
		LowPrice:       sdk.NewDec(10),
		TotalDeal:      sdk.NewInt(6),
		EndingUnixTime: lastTimeOld.Unix(),
		TimeSpan:       getSpanStrFromSpan(Minute),
		Market:         market1,
	}, *(csSlice[0]))

	csSlice = csrMan.NewBlock(T("2019-07-15T08:10:40Z"))
	require.Equal(t, 0, len(csSlice))
	csrMan.GetRecord(market1).Update(lastTime, sdk.NewDec(80), 10)
	lastTimeOld = lastTime

	csSlice = csrMan.NewBlock(T("2019-07-15T09:00:40Z"))
	require.Equal(t, CandleStick{
		OpenPrice:      sdk.NewDec(80),
		ClosePrice:     sdk.NewDec(80),
		HighPrice:      sdk.NewDec(80),
		LowPrice:       sdk.NewDec(80),
		TotalDeal:      sdk.NewInt(10),
		EndingUnixTime: lastTimeOld.Unix(),
		TimeSpan:       getSpanStrFromSpan(Minute),
		Market:         market1,
	}, *(csSlice[0]))

	require.Equal(t, CandleStick{
		OpenPrice:      sdk.NewDec(50),
		ClosePrice:     sdk.NewDec(80),
		HighPrice:      sdk.NewDec(80),
		LowPrice:       sdk.NewDec(10),
		TotalDeal:      sdk.NewInt(20),
		EndingUnixTime: lastTimeOld.Unix(),
		TimeSpan:       getSpanStrFromSpan(Hour),
		Market:         market1,
	}, *(csSlice[1]))
}

func TestDepthManager(t *testing.T) {
	dm := NewDepthManager("")
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

	ppMap, _ := dm.EndBlock()
	for _, pp := range ppMap {
		require.Equal(t, &PricePoint{sdk.NewDec(80), sdk.NewInt(70)}, pp)
	}
}

func TestMergePrice(t *testing.T) {
	update := make([]*PricePoint, 2)
	price, _ := sdk.NewDecFromStr("100.815623247626481946")
	update[0] = &PricePoint{
		Price:  price,
		Amount: sdk.NewInt(300),
	}
	price, _ = sdk.NewDecFromStr("112.812866273482850328")
	update[1] = &PricePoint{
		Price:  price,
		Amount: sdk.NewInt(300),
	}

	levels := []string{"1", "10", "100", "1000", "0.1", "0.01", "0.001", "0.0001", "0.00001", "0.000001",
		"0.0000001", "0.00000001", "0.000000001", "0.0000000001",
		"0.00000000001", "0.000000000001", "0.0000000000001",
		"0.00000000000001", "0.000000000000001", "0.0000000000000001",
		"0.00000000000000001", "0.000000000000000001"}
	//levels := []string{"1", "10", "100", "1000", "0.1", "0.01", "0.001", "0.0001", "0.00001", "0.000001", "0.0000001", "0.00000001"}
	correctStr := [][]string{
		{"price : 100.000000000000000000, amount : 300", "price : 112.000000000000000000, amount : 300"}, // 1
		{"price : 100.000000000000000000, amount : 300", "price : 110.000000000000000000, amount : 300"}, // 10
		{"price : 100.000000000000000000, amount : 600"},                                                 // 100
		{"price : 0.000000000000000000, amount : 600"},                                                   // 1000
		{"price : 100.800000000000000000, amount : 300", "price : 112.800000000000000000, amount : 300"}, // 0.1
		{"price : 100.810000000000000000, amount : 300", "price : 112.810000000000000000, amount : 300"}, // 0.01
		{"price : 100.815000000000000000, amount : 300", "price : 112.812000000000000000, amount : 300"}, // 0.001
		{"price : 100.815600000000000000, amount : 300", "price : 112.812800000000000000, amount : 300"}, // 0.0001
		{"price : 100.815620000000000000, amount : 300", "price : 112.812860000000000000, amount : 300"}, // 0.00001
		{"price : 100.815623000000000000, amount : 300", "price : 112.812866000000000000, amount : 300"}, // 0.000001
		{"price : 100.815623200000000000, amount : 300", "price : 112.812866200000000000, amount : 300"}, // 0.0000001
		{"price : 100.815623240000000000, amount : 300", "price : 112.812866270000000000, amount : 300"}, // 0.00000001
		{"price : 100.815623247000000000, amount : 300", "price : 112.812866273000000000, amount : 300"}, // 0.000000001
		{"price : 100.815623247600000000, amount : 300", "price : 112.812866273400000000, amount : 300"}, // 0.0000000001
		{"price : 100.815623247620000000, amount : 300", "price : 112.812866273480000000, amount : 300"}, // 0.00000000001
		{"price : 100.815623247626000000, amount : 300", "price : 112.812866273482000000, amount : 300"}, // 0.000000000001
		{"price : 100.815623247626400000, amount : 300", "price : 112.812866273482800000, amount : 300"}, // 0.0000000000001
		{"price : 100.815623247626480000, amount : 300", "price : 112.812866273482850000, amount : 300"}, // 0.00000000000001
		{"price : 100.815623247626481000, amount : 300", "price : 112.812866273482850000, amount : 300"}, // 0.000000000000001
		{"price : 100.815623247626481900, amount : 300", "price : 112.812866273482850300, amount : 300"}, // 0.0000000000000001
		{"price : 100.815623247626481940, amount : 300", "price : 112.812866273482850320, amount : 300"}, // 0.00000000000000001
		{"price : 100.815623247626481946, amount : 300", "price : 112.812866273482850328, amount : 300"}, // 0.000000000000000001
	}

	for i, lev := range levels {
		ret := mergePrice(update, lev, true)
		for _, point := range ret {
			report := fmt.Sprintf("price : %s, amount : %s", point.Price, point.Amount)
			if len(correctStr[i]) == 2 {
				if report != correctStr[i][0] && report != correctStr[i][1] {
					t.Errorf("actual : %s, expect one : %s, expect two : %s\n", report, correctStr[i][0], correctStr[i][1])
					return
				}
			} else {
				if report != correctStr[i][0] {
					t.Errorf("actual : %s, expect : %s\n", report, correctStr[i][0])
					return
				}
			}
		}
	}

	correctStr = [][]string{
		{"price : 101.000000000000000000, amount : 300", "price : 113.000000000000000000, amount : 300"}, // 1
		{"price : 110.000000000000000000, amount : 300", "price : 120.000000000000000000, amount : 300"}, // 10
		{"price : 200.000000000000000000, amount : 600"},                                                 // 100
		{"price : 1000.000000000000000000, amount : 600"},                                                // 1000
		{"price : 100.900000000000000000, amount : 300", "price : 112.900000000000000000, amount : 300"}, // 0.1
		{"price : 100.820000000000000000, amount : 300", "price : 112.820000000000000000, amount : 300"}, // 0.01
		{"price : 100.816000000000000000, amount : 300", "price : 112.813000000000000000, amount : 300"}, // 0.001
		{"price : 100.815700000000000000, amount : 300", "price : 112.812900000000000000, amount : 300"}, // 0.0001
		{"price : 100.815630000000000000, amount : 300", "price : 112.812870000000000000, amount : 300"}, // 0.00001
		{"price : 100.815624000000000000, amount : 300", "price : 112.812867000000000000, amount : 300"}, // 0.000001
		{"price : 100.815623300000000000, amount : 300", "price : 112.812866300000000000, amount : 300"}, // 0.0000001
		{"price : 100.815623250000000000, amount : 300", "price : 112.812866280000000000, amount : 300"}, // 0.00000001
		{"price : 100.815623248000000000, amount : 300", "price : 112.812866274000000000, amount : 300"}, // 0.000000001
		{"price : 100.815623247700000000, amount : 300", "price : 112.812866273500000000, amount : 300"}, // 0.0000000001
		{"price : 100.815623247630000000, amount : 300", "price : 112.812866273490000000, amount : 300"}, // 0.00000000001
		{"price : 100.815623247627000000, amount : 300", "price : 112.812866273483000000, amount : 300"}, // 0.000000000001
		{"price : 100.815623247626500000, amount : 300", "price : 112.812866273482900000, amount : 300"}, // 0.0000000000001
		{"price : 100.815623247626490000, amount : 300", "price : 112.812866273482860000, amount : 300"}, // 0.00000000000001
		{"price : 100.815623247626482000, amount : 300", "price : 112.812866273482851000, amount : 300"}, // 0.000000000000001
		{"price : 100.815623247626482000, amount : 300", "price : 112.812866273482850400, amount : 300"}, // 0.0000000000000001
		{"price : 100.815623247626481950, amount : 300", "price : 112.812866273482850330, amount : 300"}, // 0.00000000000000001
		{"price : 100.815623247626481946, amount : 300", "price : 112.812866273482850328, amount : 300"}, // 0.000000000000000001
	}
	for i, lev := range levels {
		ret := mergePrice(update, lev, false)
		for _, point := range ret {
			report := fmt.Sprintf("price : %s, amount : %s", point.Price, point.Amount)
			if len(correctStr[i]) == 2 {
				if report != correctStr[i][0] && report != correctStr[i][1] {
					t.Errorf("actual : %s, expect one : %s, expect two : %s\n", report, correctStr[i][0], correctStr[i][1])
					return
				}
			} else {
				if report != correctStr[i][0] {
					t.Errorf("actual : %s, expect : %s\n", report, correctStr[i][0])
					return
				}
			}
		}
	}
}

func TestGetTopicAndParams(t *testing.T) {
	subMsg := "blockinfo"
	topic, params, err := GetTopicAndParams(subMsg)
	require.Nil(t, err)
	require.EqualValues(t, BlockInfoKey, topic)
	require.EqualValues(t, 0, len(params))

	//ticker:abc/cet; ticker:B:abc/cet
	subMsg = "ticker:abc/cet"
	topic, params, err = GetTopicAndParams(subMsg)
	require.Nil(t, err)
	require.EqualValues(t, TickerKey, topic)
	require.EqualValues(t, 1, len(params))

	subMsg = "ticker:B:abc/cet"
	topic, params, err = GetTopicAndParams(subMsg)
	require.Nil(t, err)
	require.EqualValues(t, TickerKey, topic)
	require.EqualValues(t, 2, len(params))
}

func TestUnmarshal(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
	}
	var p Person

	err := unmarshalAndLogErr([]byte("{\"name\":\"cetd\"}"), &p)
	require.Nil(t, err)
	require.EqualValues(t, "cetd", p.Name)
}
