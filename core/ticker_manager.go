package core

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const MinuteNumInDay = 24 * 60

type TickerManager struct {
	// The prices in last 24 hours
	PriceList [MinuteNumInDay]sdk.Dec `json:"price_list"`
	// The following 4 members act like a small fifo
	// new Price goes first into Price1st, then goes into Price2nd,
	// finally Price2nd goes into PriceList
	Price1st  sdk.Dec `json:"price_1st"`
	Minute1st int     `json:"minute_1st"`
	Price2nd  sdk.Dec `json:"price_2nd"`
	Minute2nd int     `json:"minute_2nd"`

	Market      string `json:"market"`
	Initialized bool   `json:"initialized"`
}

func NewTickerManager(Market string) *TickerManager {
	return &TickerManager{
		Market: Market,
	}
}

// Flush the cached NewestPrice and NewestMinute to PriceList,
// and assign currPrice to NewestPrice, currMinute to NewestMinute
func (tm *TickerManager) UpdateNewestPrice(currPrice sdk.Dec, currMinute int) {
	//fmt.Printf("@@ %s: %s %d\n", tm.Market, currPrice, currMinute)
	if currMinute >= MinuteNumInDay || currMinute < 0 {
		panic(fmt.Sprintf("Minute not Valid: %d", currMinute))
	}
	if !tm.Initialized {
		// Initialize all the members with currPrice
		tm.Initialized = true
		for i := 0; i < MinuteNumInDay; i++ {
			tm.PriceList[i] = currPrice
		}
		tm.Price1st = currPrice
		tm.Minute1st = currMinute
		tm.Price2nd = currPrice
		tm.Minute2nd = currMinute
		return
	}

	tm.fillHistory(tm.Minute2nd, tm.Minute1st, tm.Price2nd)
	tm.Price2nd = tm.Price1st
	tm.Minute2nd = tm.Minute1st
	tm.Price1st = currPrice
	tm.Minute1st = currMinute
}

func (tm *TickerManager) fillHistory(beginTime, endTime int, price sdk.Dec) {
	for {
		tm.PriceList[beginTime] = price
		beginTime++
		if beginTime >= MinuteNumInDay {
			beginTime = 0
		}
		if beginTime == endTime {
			break
		}
	}
}

// Return a Ticker if NewPrice or OldPriceOneDayAgo is different from its previous minute
func (tm *TickerManager) GetTicker(currMinute int) *Ticker {
	if !tm.Initialized {
		return nil
	}
	if currMinute >= MinuteNumInDay || currMinute < 0 {
		panic(fmt.Sprintf("Minute not Valid: %d", currMinute))
	}
	lastMinute := currMinute - 1
	if lastMinute < 0 {
		lastMinute = MinuteNumInDay - 1
	}
	// If the price changes at currMinute of today
	if (tm.Minute1st == currMinute && !tm.Price1st.Equal(tm.Price2nd)) ||
		// Or If price changes at currMinute of yesterday
		!tm.PriceList[currMinute].Equal(tm.PriceList[lastMinute]) {
		return &Ticker{
			NewPrice:          tm.Price1st,
			OldPriceOneDayAgo: tm.PriceList[currMinute],
			Market:            tm.Market,
			MinuteInDay:       currMinute,
		}
	}
	if tm.Minute1st != currMinute { //flush the price into the small fifo
		tm.UpdateNewestPrice(tm.Price1st, currMinute)
	}
	return nil
}
