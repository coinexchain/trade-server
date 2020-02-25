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

func (manager *CandleStickManager) NewBlock(t time.Time) []CandleStick {
	res := make([]CandleStick, 0, 100)
	isNewDay := t.UTC().Day() != manager.LastBlockTime.UTC().Day() || t.Unix()-manager.LastBlockTime.Unix() > 60*60*24
	isNewHour := t.UTC().Hour() != manager.LastBlockTime.UTC().Hour() || t.Unix()-manager.LastBlockTime.Unix() > 60*60
	isNewMinute := t.UTC().Minute() != manager.LastBlockTime.UTC().Minute() || t.Unix()-manager.LastBlockTime.Unix() > 60
	for _, csr := range manager.CsrMap {
		csSlice := csr.newBlock(isNewDay, isNewHour, isNewMinute, manager.LastBlockTime)
		res = append(res, csSlice...)
	}
	manager.LastBlockTime = t
	return res
}

func (manager *CandleStickManager) GetRecord(Market string) *CandleStickRecord {
	csr, ok := manager.CsrMap[Market]
	if ok {
		return csr
	}
	return nil
}
