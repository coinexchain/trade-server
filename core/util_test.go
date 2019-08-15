package core

import (
	"fmt"
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

//func (csr *CandleStickRecord) Update(t time.Time, price sdk.Dec, amount int64) {
//func (csr *CandleStickRecord) newBlock(t time.Time, isNewDay, isNewHour, isNewMinute bool) []CandleStick {
func TestCandleStickRecord(t *testing.T) {
	t0, err := time.Parse(time.RFC3339, "1990-12-31T23:59:40Z")
	require.Equal(t, nil, err)
	fmt.Printf("Unix: %d\n", t0.Unix())
}
