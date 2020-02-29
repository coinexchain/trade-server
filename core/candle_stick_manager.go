package core

import "time"

// Manager of the CandleStickRecords for all markets
type CandleStickManager struct {
	CsrMap        map[string]*CandleStickRecord `json:"csr_map"`
	LastBlockTime time.Time                     `json:"last_block_time"`
}

func NewCandleStickManager(markets []string) CandleStickManager {
	res := CandleStickManager{CsrMap: make(map[string]*CandleStickRecord)}
	for _, market := range markets {
		res.CsrMap[market] = NewCandleStickRecord(market)
	}
	return res
}

func (manager *CandleStickManager) AddMarket(market string) {
	manager.CsrMap[market] = NewCandleStickRecord(market)
}

// When a new blocks come at 't', flush out all the finished candle sticks util 't'
// The recorded manager.LastBlockTime will be used as the EndingTime of candle sticks
func (manager *CandleStickManager) NewBlock(currBlockTime time.Time) []*CandleStick {
	res := make([]*CandleStick, 0, 100)
	isNewDay := currBlockTime.UTC().Day() != manager.LastBlockTime.UTC().Day() || currBlockTime.UTC().Unix()-manager.LastBlockTime.UTC().Unix() > 60*60*24
	isNewHour := currBlockTime.UTC().Hour() != manager.LastBlockTime.UTC().Hour() || currBlockTime.UTC().Unix()-manager.LastBlockTime.UTC().Unix() > 60*60
	isNewMinute := currBlockTime.UTC().Minute() != manager.LastBlockTime.UTC().Minute() || currBlockTime.UTC().Unix()-manager.LastBlockTime.UTC().Unix() > 60
	for _, csr := range manager.CsrMap {
		csSlice := csr.newBlock(isNewDay, isNewHour, isNewMinute, manager.LastBlockTime)
		res = append(res, csSlice...)
	}
	manager.LastBlockTime = currBlockTime
	return res
}

func (manager *CandleStickManager) GetRecord(Market string) *CandleStickRecord {
	csr, ok := manager.CsrMap[Market]
	if ok {
		return csr
	}
	return nil
}
