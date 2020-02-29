package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestXTickerManager_UpdateNewestPriceTmp(t *testing.T) {
	ticker := NewTickerManager("abc/cet")
	ticker.UpdateNewestPriceTmp(sdk.NewDec(99), 30)
	require.EqualValues(t, true, ticker.Initialized)
	require.EqualValues(t, sdk.NewDec(99), ticker.Price1st)
	require.EqualValues(t, sdk.NewDec(99), ticker.Price2nd)
	require.EqualValues(t, 30, ticker.Minute1st)
	require.EqualValues(t, 30, ticker.Minute2nd)

	ticker.UpdateNewestPriceTmp(sdk.NewDec(29), 60)
	fmt.Println(ticker.Price1st, ticker.Price2nd, ticker.Minute1st, ticker.Minute2nd)
	require.EqualValues(t, sdk.NewDec(29), ticker.Price1st)
	require.EqualValues(t, sdk.NewDec(99), ticker.Price2nd)
	require.EqualValues(t, 60, ticker.Minute1st)
	require.EqualValues(t, 30, ticker.Minute2nd)

	ticker.Price2nd = sdk.NewDec(169)
	ticker.UpdateNewestPriceTmp(sdk.NewDec(69), 90)
	fmt.Println(ticker.Price1st, ticker.Price2nd, ticker.Minute1st, ticker.Minute2nd)
	require.EqualValues(t, sdk.NewDec(69), ticker.Price1st)
	require.EqualValues(t, sdk.NewDec(29), ticker.Price2nd)
	require.EqualValues(t, 90, ticker.Minute1st)
	require.EqualValues(t, 60, ticker.Minute2nd)

	for i := 30; i < 60; i++ {
		require.EqualValues(t, sdk.NewDec(169), ticker.PriceList[i])
	}

	ticker.UpdateNewestPriceTmp(sdk.NewDec(269), 91)
	ticker.UpdateNewestPriceTmp(sdk.NewDec(369), 92)
	require.EqualValues(t, sdk.NewDec(369), ticker.Price1st)
	require.EqualValues(t, sdk.NewDec(269), ticker.Price2nd)
	require.EqualValues(t, 92, ticker.Minute1st)
	require.EqualValues(t, 91, ticker.Minute2nd)
	require.EqualValues(t, sdk.NewDec(69), ticker.PriceList[90])
	require.EqualValues(t, sdk.NewDec(99), ticker.PriceList[91])

}
