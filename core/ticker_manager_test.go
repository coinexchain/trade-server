package core

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestTickerManager_UpdateNewestPrice_AndGetTicker(t *testing.T) {
	ticker := NewTickerManager("abc/cet")
	// init tickerManager
	ticker.UpdateNewestPrice(sdk.NewDec(99), 30)
	require.EqualValues(t, true, ticker.Initialized)
	require.EqualValues(t, sdk.NewDec(99), ticker.Price1st)
	require.EqualValues(t, sdk.NewDec(99), ticker.Price2nd)
	require.EqualValues(t, 30, ticker.Minute1st)
	require.EqualValues(t, 30, ticker.Minute2nd)

	// update priceList [nil]
	ticker.UpdateNewestPrice(sdk.NewDec(29), 60)
	require.EqualValues(t, sdk.NewDec(29), ticker.Price1st)
	require.EqualValues(t, sdk.NewDec(99), ticker.Price2nd)
	require.EqualValues(t, 60, ticker.Minute1st)
	require.EqualValues(t, 30, ticker.Minute2nd)
	assertGetTickerInOneDay(t, ticker, 60, 29, 99)

	// update priceList [30, 60) --> 169
	ticker.Price2nd = sdk.NewDec(169)
	ticker.UpdateNewestPrice(sdk.NewDec(69), 90)
	require.EqualValues(t, sdk.NewDec(69), ticker.Price1st)
	require.EqualValues(t, sdk.NewDec(29), ticker.Price2nd)
	require.EqualValues(t, 90, ticker.Minute1st)
	require.EqualValues(t, 60, ticker.Minute2nd)
	for i := 30; i < 60; i++ {
		require.EqualValues(t, sdk.NewDec(169), ticker.PriceList[i])
	}
	assertGetTickerInOneDay(t, ticker, 90, 69, 99)
	assertGetTickerInTwoDay(t, ticker, 30, 69, 169)

	// update priceList [60, 90) --> 29
	ticker.UpdateNewestPrice(sdk.NewDec(269), 91)
	for i := 60; i < 90; i++ {
		require.EqualValues(t, sdk.NewDec(29), ticker.PriceList[i])
	}
	// update priceList [90, 91) --> 69
	ticker.UpdateNewestPrice(sdk.NewDec(369), 92)
	require.EqualValues(t, sdk.NewDec(369), ticker.Price1st)
	require.EqualValues(t, sdk.NewDec(269), ticker.Price2nd)
	require.EqualValues(t, 92, ticker.Minute1st)
	require.EqualValues(t, 91, ticker.Minute2nd)
	require.EqualValues(t, sdk.NewDec(69), ticker.PriceList[90])
	require.EqualValues(t, sdk.NewDec(99), ticker.PriceList[91])
	assertGetTickerInOneDay(t, ticker, 92, 369, 99)
	assertGetTickerInTwoDay(t, ticker, 90, 369, 69)
}

func assertGetTickerInOneDay(t *testing.T, manager *TickerManager, currMinute int, newestPrice, oldPrice int64) {
	ticker := manager.GetTicker(currMinute)
	require.EqualValues(t, sdk.NewDec(newestPrice), ticker.NewPrice)
	require.EqualValues(t, sdk.NewDec(oldPrice), ticker.OldPriceOneDayAgo)
	require.EqualValues(t, manager.Market, ticker.Market)
}

func assertGetTickerInTwoDay(t *testing.T, manager *TickerManager, queryTime int, newestPrice, oldPrice int64) {
	ticker := manager.GetTicker(queryTime)
	require.EqualValues(t, sdk.NewDec(newestPrice), ticker.NewPrice)
	require.EqualValues(t, sdk.NewDec(oldPrice), ticker.OldPriceOneDayAgo)
	require.EqualValues(t, manager.Market, ticker.Market)
}
