package core

import sdk "github.com/cosmos/cosmos-sdk/types"

// This class models one candle stick in the K-line
// When we add extra information (EndingUnixTime, TimeSpan and Market) to it,
// it becomes a full CandleStick for display
type baseCandleStick struct {
	OpenPrice  sdk.Dec `json:"open"`
	ClosePrice sdk.Dec `json:"close"`
	HighPrice  sdk.Dec `json:"high"`
	LowPrice   sdk.Dec `json:"low"`
	TotalDeal  sdk.Int `json:"total"`
}

func newBaseCandleStick() baseCandleStick {
	return baseCandleStick{
		OpenPrice:  sdk.ZeroDec(),
		ClosePrice: sdk.ZeroDec(),
		HighPrice:  sdk.ZeroDec(),
		LowPrice:   sdk.ZeroDec(),
		TotalDeal:  sdk.ZeroInt(),
	}
}

func (cs *baseCandleStick) hasDeal() bool {
	return cs.TotalDeal.IsPositive()
}

// When new deal comes, update candle stick accordingly
func (cs *baseCandleStick) update(price sdk.Dec, amount int64) {
	if !cs.hasDeal() {
		cs.OpenPrice = price
		cs.HighPrice = price
		cs.LowPrice = price
	} else {
		if cs.HighPrice.LT(price) {
			cs.HighPrice = price
		}
		if cs.LowPrice.GT(price) {
			cs.LowPrice = price
		}
	}
	cs.ClosePrice = price
	cs.TotalDeal = cs.TotalDeal.AddRaw(amount)
}
