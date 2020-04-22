package core

import (
	"encoding/binary"
	"strings"
	"sync/atomic"
	"time"
)

// for serialization and deserialization of Hub

// to dump hub's in-memory state to json
type HubForJSON struct {
	Sid             int64                `json:"sid"`
	CSMan           CandleStickManager   `json:"csman"`
	TickerMap       map[string]*Ticker   `json:"ticker_map"`
	CurrBlockHeight int64                `json:"curr_block_height"`
	CurrBlockTime   int64                `json:"curr_block_time"`
	LastBlockTime   int64                `json:"last_block_time"`
	Markets         []*MarketInfoForJSON `json:"markets"`
}

type MarketInfoForJSON struct {
	TkMan           *TickerManager `json:"tkman"`
	SellPricePoints []*PricePoint  `json:"sells"`
	BuyPricePoints  []*PricePoint  `json:"buys"`
}

func (hub *Hub) Load(hub4j *HubForJSON) {
	hub.sid = hub4j.Sid
	hub.csMan = hub4j.CSMan
	hub.tickerMap = hub4j.TickerMap
	hub.currBlockHeight = hub4j.CurrBlockHeight
	hub.currBlockTime = time.Unix(0, hub4j.CurrBlockTime)
	hub.lastBlockTime = time.Unix(0, hub4j.LastBlockTime)

	for _, info := range hub4j.Markets {
		triman := &TripleManager{
			sell: NewDepthManager("sell"),
			buy:  NewDepthManager("buy"),
			tkm:  info.TkMan,
		}
		for _, pp := range info.SellPricePoints {
			triman.sell.DeltaChange(pp.Price, pp.Amount)
		}
		for _, pp := range info.BuyPricePoints {
			triman.buy.DeltaChange(pp.Price, pp.Amount)
		}
		hub.managersMap[info.TkMan.Market] = triman
	}
}

func (hub *Hub) Dump(hub4j *HubForJSON) {
	hub4j.Sid = hub.sid
	hub4j.CSMan = hub.csMan
	hub4j.TickerMap = hub.tickerMap
	hub4j.CurrBlockHeight = hub.currBlockHeight
	hub4j.CurrBlockTime = hub.currBlockTime.UnixNano()
	hub4j.LastBlockTime = hub.lastBlockTime.UnixNano()

	hub4j.Markets = make([]*MarketInfoForJSON, 0, len(hub.managersMap))
	for _, triman := range hub.managersMap {
		hub4j.Markets = append(hub4j.Markets, &MarketInfoForJSON{
			TkMan:           triman.tkm,
			SellPricePoints: triman.sell.DumpPricePoints(),
			BuyPricePoints:  triman.buy.DumpPricePoints(),
		})
	}
}

func (hub *Hub) LoadDumpData() []byte {
	key := GetDumpKey()
	hub.dbMutex.RLock()
	defer hub.dbMutex.RUnlock()
	bz := hub.db.Get(key)
	return bz
}

func (hub *Hub) UpdateOffset(partition int32, offset int64) {
	hub.partition = partition
	hub.lastOffset = hub.offset
	hub.offset = offset

	now := time.Now()
	// dump data every <interval> offset
	if hub.offset-hub.lastOffset >= DumpInterval && now.Sub(hub.lastDumpTime) > DumpMinTime {
		hub.dumpFlagLock.Lock()
		defer hub.dumpFlagLock.Unlock()
		atomic.StoreInt32(&hub.dumpFlag, 1)
	}
}

func (hub *Hub) LoadOffset(partition int32) int64 {
	key := GetOffsetKey(partition)
	hub.partition = partition

	hub.dbMutex.RLock()
	defer hub.dbMutex.RUnlock()

	offsetBuf := hub.db.Get(key)
	if offsetBuf == nil {
		hub.offset = 0
	} else {
		hub.offset = int64(binary.BigEndian.Uint64(offsetBuf))
	}
	return hub.offset
}

func (hub *Hub) Close() {
	// try to dump the latest block
	hub.dumpFlagLock.Lock()
	atomic.StoreInt32(&hub.dumpFlag, 1)
	hub.dumpFlagLock.Unlock()
	for i := 0; i < 5; i++ { // Why? is 5 seconds enough?
		time.Sleep(time.Second)
		if !hub.isDumped() {
			break
		}
	}
	// close db
	hub.dbMutex.Lock()
	defer hub.dbMutex.Unlock()
	hub.db.Close()
	hub.stopped = true
}

func (hub *Hub) isDumped() bool {
	hub.dumpFlagLock.Lock()
	defer hub.dumpFlagLock.Unlock()
	return atomic.LoadInt32(&hub.dumpFlag) != 0
}

func (hub *Hub) AddLevel(market, level string) error {
	if !hub.HasMarket(market) {
		hub.AddMarket(market)
	}
	tripleMan := hub.managersMap[market]
	err := tripleMan.sell.AddLevel(level)
	if err != nil {
		return err
	}
	return tripleMan.buy.AddLevel(level)
}

func (hub *Hub) HasMarket(market string) bool {
	_, ok := hub.managersMap[market]
	return ok
}

func (hub *Hub) AddMarket(market string) {
	if strings.HasPrefix(market, "B:") {
		// A bancor market has no depth information
		hub.managersMap[market] = &TripleManager{
			tkm: NewTickerManager(market),
		}
	} else {
		// A normal market
		hub.managersMap[market] = &TripleManager{
			sell: NewDepthManager("sell"),
			buy:  NewDepthManager("buy"),
			tkm:  NewTickerManager(market),
		}
	}
	hub.csMan.AddMarket(market)
}
