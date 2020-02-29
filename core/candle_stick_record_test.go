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

func TestCandleStickRecord_NewBlock(t *testing.T) {
	record := NewCandleStickRecord("abc/cet")
	now := time.Now()
	updateCandleStick := record.newBlock(true, false, true, now)
	_ = updateCandleStick
}
