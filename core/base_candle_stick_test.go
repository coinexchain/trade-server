package core

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestBaseCandleStick_Update(t *testing.T) {
	assertDealZeroCandleStick(t)
	assertHaveDealCandleStick(t)
}

func assertDealZeroCandleStick(t *testing.T) {
	singleCandleStick := newBaseCandleStick()
	require.False(t, singleCandleStick.hasDeal())

	singleCandleStick.update(sdk.NewDec(100), 99)
	require.EqualValues(t, sdk.NewDec(100), singleCandleStick.LowPrice)
	require.EqualValues(t, sdk.NewDec(100), singleCandleStick.HighPrice)
	require.EqualValues(t, sdk.NewDec(100), singleCandleStick.ClosePrice)
	require.EqualValues(t, sdk.NewDec(100), singleCandleStick.OpenPrice)
	require.EqualValues(t, 99, singleCandleStick.TotalDeal.Int64())
}

func assertHaveDealCandleStick(t *testing.T) {
	singleCandleStick := newBaseCandleStick()
	singleCandleStick.TotalDeal = sdk.NewInt(100)
	singleCandleStick.LowPrice = sdk.NewDec(110)
	singleCandleStick.HighPrice = sdk.NewDec(210)

	singleCandleStick.update(sdk.NewDec(100), 99)
	require.EqualValues(t, sdk.NewDec(100), singleCandleStick.LowPrice)
	require.EqualValues(t, sdk.NewDec(210), singleCandleStick.HighPrice)
	require.EqualValues(t, sdk.NewDec(100), singleCandleStick.ClosePrice)
	require.EqualValues(t, sdk.NewDec(0), singleCandleStick.OpenPrice)
	require.EqualValues(t, 199, singleCandleStick.TotalDeal.Int64())
}
