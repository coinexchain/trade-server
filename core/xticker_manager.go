package core

import sdk "github.com/cosmos/cosmos-sdk/types"

type XTickerManager struct {
	PriceList         [MinuteNumInDay]sdk.Dec `json:"price_list"`
	AmountList        [MinuteNumInDay]sdk.Int `json:"amount_list"`
	Highest           sdk.Dec                 `json:"highest"`
	Lowest            sdk.Dec                 `json:"lowest"`
	Market            string                  `json:"market"`
	Initialized       bool                    `json:"initialized"`
	TotalAmount       sdk.Int                 `json:"amount"`
	LastUpdate        int                     `json:"amount_last_update"`
	OldPriceOneDayAgo sdk.Dec                 `json:"old_price"`
}

func NewXTickerManager(Market string) *XTickerManager {
	return &XTickerManager{
		Market:      Market,
		TotalAmount: sdk.ZeroInt(),
	}
}

func (tm *XTickerManager) updateHighest() {
	tm.Highest = sdk.ZeroDec()
	for i := 0; i < MinuteNumInDay; i++ {
		price := tm.PriceList[i]
		if !tm.AmountList[i].IsZero() && price.GT(tm.Highest) {
			tm.Highest = price
		}
	}
}
func (tm *XTickerManager) updateLowest() {
	tm.Lowest = sdk.ZeroDec()
	for i := 0; i < MinuteNumInDay; i++ {
		price := tm.PriceList[i]
		if !tm.AmountList[i].IsZero() && (price.LT(tm.Lowest) || tm.Lowest.IsZero()) {
			tm.Lowest = price
		}
	}
}

func (tm *XTickerManager) setPrice(i int, price sdk.Dec, amount sdk.Int, highestMayChange *bool, lowestMayChange *bool) {
	*highestMayChange = *highestMayChange || (tm.PriceList[i].Equal(tm.Highest) && !tm.AmountList[i].IsZero())
	*lowestMayChange = *lowestMayChange || (tm.PriceList[i].Equal(tm.Lowest) && !tm.AmountList[i].IsZero())
	tm.PriceList[i] = price
	if !amount.IsZero() && price.GT(tm.Highest) {
		tm.Highest = price
	}
	if !amount.IsZero() && (price.LT(tm.Lowest) || tm.Lowest.IsZero()) {
		tm.Lowest = price
	}
}

func (tm *XTickerManager) UpdateNewestPrice(currPrice sdk.Dec, currMinute int, currAmount sdk.Int) {
	if currMinute >= MinuteNumInDay || currMinute < 0 {
		panic("Minute not valid")
	}
	if !tm.Initialized {
		tm.Initialized = true
		for i := 0; i < MinuteNumInDay; i++ {
			tm.PriceList[i] = currPrice
			tm.AmountList[i] = sdk.ZeroInt()
		}
		tm.Highest = currPrice
		tm.Lowest = currPrice
		tm.TotalAmount = currAmount
		//fmt.Printf("Here! %d %s %s\n", currMinute, tm.Market, tm.TotalAmount)
		tm.AmountList[currMinute] = currAmount
		tm.LastUpdate = currMinute
		return
	}

	tm.OldPriceOneDayAgo = tm.PriceList[currMinute]
	lastPrice := tm.PriceList[tm.LastUpdate]
	highestMayChange := false
	lowestMayChange := false
	for {
		tm.LastUpdate++
		if tm.LastUpdate >= MinuteNumInDay {
			tm.LastUpdate = 0
		}
		tm.TotalAmount = tm.TotalAmount.Sub(tm.AmountList[tm.LastUpdate])
		//fmt.Printf("Here Sub! %d %s %s\n", tm.LastUpdate, tm.AmountList[tm.LastUpdate], tm.TotalAmount)
		if tm.LastUpdate == currMinute {
			tm.setPrice(tm.LastUpdate, currPrice, currAmount, &highestMayChange, &lowestMayChange)
			tm.AmountList[tm.LastUpdate] = currAmount
			break
		} else {
			tm.setPrice(tm.LastUpdate, lastPrice, currAmount, &highestMayChange, &lowestMayChange)
			tm.AmountList[tm.LastUpdate] = sdk.ZeroInt()
		}
	}
	if highestMayChange {
		tm.updateHighest()
	}
	if lowestMayChange {
		tm.updateLowest()
	}
	tm.TotalAmount = tm.TotalAmount.Add(currAmount)
}

func (tm *XTickerManager) GetXTicker(currMinute int) *XTicker {
	if !tm.Initialized {
		return nil
	}
	if currMinute != tm.LastUpdate {
		tm.UpdateNewestPrice(tm.PriceList[tm.LastUpdate], currMinute, sdk.ZeroInt())
	}
	//fmt.Printf("Here Sicker! %s %s\n", tm.Market, tm.TotalAmount)
	return &XTicker{
		Market:            tm.Market,
		MinuteInDay:       tm.LastUpdate,
		NewPrice:          tm.PriceList[tm.LastUpdate],
		OldPriceOneDayAgo: tm.OldPriceOneDayAgo,
		HighPrice:         tm.Highest,
		LowPrice:          tm.Lowest,
		TotalDeal:         tm.TotalAmount,
	}
}
