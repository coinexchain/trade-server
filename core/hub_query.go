package core

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync/atomic"
)

//============================================================
var _ Querier = &Hub{}

func (hub *Hub) QueryTickers(marketList []string) []*Ticker {
	tickerList := make([]*Ticker, 0, len(marketList))
	hub.tickerMapMutex.RLock()
	atomic.AddInt64(&hub.tickerMapLockCount, 1)
	for _, market := range marketList {
		ticker, ok := hub.tickerMap[market]
		if ok {
			tickerList = append(tickerList, ticker)
		}
	}
	hub.tickerMapMutex.RUnlock()
	return tickerList
}

func (hub *Hub) QueryLatestHeight() int64 {
	bz := hub.db.Get([]byte{LatestHeightByte})
	if len(bz) == 0 {
		return 0
	}
	return BigEndianBytesToInt64(bz)
}

func (hub *Hub) QueryBlockInfo() *HeightInfo {
	heightBytes := Int64ToBigEndianBytes(math.MaxInt64)
	key := append([]byte{BlockHeightByte}, heightBytes...)

	hub.dbMutex.RLock()
	iter := hub.db.ReverseIterator(nil, key)
	defer func() {
		iter.Close()
		hub.dbMutex.RUnlock()
	}()

	var v HeightInfo
	if iter.Valid() {
		v.Height = binary.BigEndian.Uint64(iter.Key()[1:])
		v.TimeStamp = binary.LittleEndian.Uint64(iter.Value())
	}
	return &v
}

func (hub *Hub) QueryBlockTime(height int64, count int) []int64 {
	count = limitCount(count)
	data := make([]int64, 0, count)
	end := append([]byte{BlockHeightByte}, Int64ToBigEndianBytes(height+1)...)
	start := []byte{BlockHeightByte}
	hub.dbMutex.RLock()
	iter := hub.db.ReverseIterator(start, end)
	defer func() {
		iter.Close()
		hub.dbMutex.RUnlock()
	}()
	for ; iter.Valid(); iter.Next() {
		unixSec := binary.LittleEndian.Uint64(iter.Value())
		data = append(data, int64(unixSec))
		if count--; count == 0 {
			break
		}
	}
	return data
}

func (hub *Hub) QueryDepth(market string, count int) (sell []*PricePoint, buy []*PricePoint) {
	count = limitCount(count)
	if !hub.HasMarket(market) {
		return
	}
	tripleMan := hub.managersMap[market]
	tripleMan.mutex.RLock()
	defer tripleMan.mutex.RUnlock()
	atomic.AddInt64(&hub.trimanLockCount, 1)
	sell = tripleMan.sell.GetLowest(count)
	buy = tripleMan.buy.GetHighest(count)
	return
}

func (hub *Hub) QueryCandleStick(market string, timespan byte, time int64, sid int64, count int) []json.RawMessage {
	count = limitCount(count)
	data := make([]json.RawMessage, 0, count)
	end := getCandleStickEndKey(market, timespan, time, sid)
	start := getCandleStickStartKey(market, timespan)
	hub.dbMutex.RLock()
	iter := hub.db.ReverseIterator(start, end)
	defer func() {
		iter.Close()
		hub.dbMutex.RUnlock()
	}()
	for ; iter.Valid(); iter.Next() {
		data = append(data, json.RawMessage(iter.Value()))
		if count--; count == 0 {
			break
		}
	}
	return data
}

func (hub *Hub) QueryTxByHashID(hexHashID string) json.RawMessage {
	hub.dbMutex.RLock()
	defer hub.dbMutex.RUnlock()
	key := append([]byte{DetailByte}, []byte(hexHashID)...)
	iter := hub.db.Iterator(key, nil)
	defer iter.Close()
	if !iter.Valid() {
		return json.RawMessage([]byte{})
	}
	key = iter.Key()
	value := iter.Value()
	if len(key) > 1+len(hexHashID) && string(key[1:1+len(hexHashID)]) == hexHashID {
		return json.RawMessage(value)
	}
	return json.RawMessage([]byte{})
}

//=========

func (hub *Hub) QueryOrder(account string, time int64, sid int64, count int) (
	data []json.RawMessage, tags []byte, timesid []int64) {
	return hub.query(false, OrderByte, []byte(account), time, sid, count, nil)
}

func (hub *Hub) QueryBancorDeal(market string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	data, _, timesid = hub.query(false, BancorDealByte, []byte(market), time, sid, count, nil)
	return
}

func (hub *Hub) QueryDeal(market string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	data, _, timesid = hub.query(false, DealByte, []byte(market), time, sid, count, nil)
	return
}

func (hub *Hub) QueryBancorInfo(market string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	data, _, timesid = hub.query(false, BancorInfoByte, []byte(market), time, sid, count, nil)
	return
}

func (hub *Hub) QueryBancorTrade(account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	data, _, timesid = hub.query(false, BancorTradeByte, []byte(account), time, sid, count, nil)
	return
}

func (hub *Hub) QueryRedelegation(account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	data, _, timesid = hub.query(false, RedelegationByte, []byte(account), time, sid, count, nil)
	return
}
func (hub *Hub) QueryUnbonding(account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	data, _, timesid = hub.query(false, UnbondingByte, []byte(account), time, sid, count, nil)
	return
}
func (hub *Hub) QueryUnlock(account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	data, _, timesid = hub.query(false, UnlockByte, []byte(account), time, sid, count, nil)
	return
}

func (hub *Hub) QueryIncome(account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	data, _, timesid = hub.query(true, IncomeByte, []byte(account), time, sid, count, nil)
	return
}

func (hub *Hub) QueryTx(account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	data, _, timesid = hub.query(true, TxByte, []byte(account), time, sid, count, nil)
	return
}

func (hub *Hub) QueryLocked(account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	data, _, timesid = hub.query(false, LockedByte, []byte(account), time, sid, count, nil)
	return
}

func (hub *Hub) QueryComment(token string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	data, _, timesid = hub.query(false, CommentByte, []byte(token), time, sid, count, nil)
	return
}

func (hub *Hub) QuerySlash(time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	data, _, timesid = hub.query(false, SlashByte, []byte{}, time, sid, count, nil)
	return
}

func (hub *Hub) QueryDonation(time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	data, _, timesid = hub.query(false, DonationByte, []byte{}, time, sid, count, nil)
	return
}

func (hub *Hub) QueryDelist(market string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	data, _, timesid = hub.query(false, DelistByte, []byte(market), time, sid, count, nil)
	return
}

func (hub *Hub) QueryDelists(time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	data, _, timesid = hub.query(false, DelistsByte, []byte{}, time, sid, count, nil)
	return
}

// --------------

func (hub *Hub) QueryOrderAboutToken(tag, token, account string, time int64, sid int64, count int) (
	data []json.RawMessage, tags []byte, timesid []int64) {
	firstByte := OrderByte
	if tag == CreateOrderStr {
		firstByte = CreateOrderOnlyByte
	} else if tag == FillOrderStr {
		firstByte = FillOrderOnlyByte
	} else if tag == CancelOrderStr {
		firstByte = CancelOrderOnlyByte
	}
	if token == "" { // no token-based-filtering
		return hub.query(false, firstByte, []byte(account), time, sid, count, nil)
	}
	return hub.query(false, firstByte, []byte(account), time, sid, count, func(tag byte, entry []byte) bool {
		s1 := fmt.Sprintf("/%s\",\"height\":", token)
		s2 := fmt.Sprintf("\"trading_pair\":\"%s/", token)
		if tag == CreateOrderEndByte {
			s1 = fmt.Sprintf("/%s\",\"order_type\":", token)
		}
		return strings.Index(string(entry), s1) > 0 || strings.Index(string(entry), s2) > 0
	})

}

func (hub *Hub) QueryLockedAboutToken(token, account string, time int64, sid int64, count int) (
	data []json.RawMessage, timesid []int64) {
	if token == "" { // no token-based-filtering
		data, _, timesid = hub.query(false, LockedByte, []byte(account), time, sid, count, nil)
		return
	}
	data, _, timesid = hub.query(false, LockedByte, []byte(account),
		time, sid, count, func(tag byte, entry []byte) bool {
			s := fmt.Sprintf("\"denom\":\"%s\",\"amount\":", token)
			return strings.Index(string(entry), s) > 0
		})
	return

}

func (hub *Hub) QueryBancorTradeAboutToken(token, account string, time int64, sid int64, count int) (
	data []json.RawMessage, timesid []int64) {
	if token == "" { // no token-based-filtering
		data, _, timesid = hub.query(false, BancorTradeByte, []byte(account), time, sid, count, nil)
		return
	}
	data, _, timesid = hub.query(false, BancorTradeByte, []byte(account), time, sid,
		count, func(tag byte, entry []byte) bool {
			s1 := fmt.Sprintf("\"money\":\"%s\"", token)
			s2 := fmt.Sprintf("\"stock\":\"%s\"", token)
			return strings.Index(string(entry), s1) > 0 || strings.Index(string(entry), s2) > 0
		})
	return

}

func (hub *Hub) QueryUnlockAboutToken(token, account string, time int64, sid int64, count int) (
	data []json.RawMessage, timesid []int64) {
	if token == "" { // no token-based-filtering
		data, _, timesid = hub.query(false, UnlockByte, []byte(account), time, sid, count, nil)
		return
	}
	data, _, timesid = hub.query(false, UnlockByte, []byte(account), time, sid, count,
		func(tag byte, entry []byte) bool {
			entryStr := string(entry)
			ending := strings.Index(entryStr, "\"locked_coins\":")
			if ending < 0 {
				return false
			}
			s := fmt.Sprintf("\"denom\":\"%s\",\"amount\":", token)
			return strings.Index(entryStr[:ending], s) > 0
		})
	return

}

func (hub *Hub) QueryIncomeAboutToken(token, account string, time int64, sid int64, count int) (
	data []json.RawMessage, timesid []int64) {
	if token == "" { // no token-based-filtering
		data, _, timesid = hub.query(true, IncomeByte, []byte(account), time, sid, count, nil)
		return
	}
	data, _, timesid = hub.query(true, IncomeByte, []byte(account), time, sid, count,
		func(tag byte, entry []byte) bool {
			return strings.Contains(string(entry), "|"+token+"|")
		})
	return

}

func (hub *Hub) QueryTxAboutToken(token, account string, time int64, sid int64, count int) (
	data []json.RawMessage, timesid []int64) {
	if token == "" { // no token-based-filtering
		data, _, timesid = hub.query(true, TxByte, []byte(account), time, sid, count, nil)
		return
	}
	data, _, timesid = hub.query(true, TxByte, []byte(account), time, sid, count,
		func(tag byte, entry []byte) bool {
			return strings.Contains(string(entry), "|"+token+"|")
		})
	return
}

type filterFunc func(tag byte, entry []byte) bool

func (hub *Hub) query(fetchTxDetail bool, firstByteIn byte, bz []byte, time int64, sid int64,
	count int, filter filterFunc) (data []json.RawMessage, tags []byte, timesid []int64) {
	firstByte := firstByteIn
	if firstByteIn == CreateOrderOnlyByte || firstByteIn == FillOrderOnlyByte || firstByteIn == CancelOrderOnlyByte {
		firstByte = OrderByte
	} else if firstByteIn == DelistsByte {
		firstByte = DelistByte
	} else {
		firstByteIn = 0
	}
	count = limitCount(count)
	data = make([]json.RawMessage, 0, count)
	tags = make([]byte, 0, count)
	timesid = make([]int64, 0, 2*count)
	start := getStartKeyFromBytes(firstByte, bz)
	end := getEndKeyFromBytes(firstByte, bz, time, sid)
	if firstByteIn == DelistsByte {
		end = []byte{DelistsByte} // To a different 'firstByte'
	}
	hub.dbMutex.RLock()
	iter := hub.db.ReverseIterator(start, end)
	defer func() {
		iter.Close()
		hub.dbMutex.RUnlock()
	}()
	for ; iter.Valid(); iter.Next() {
		iKey := iter.Key()
		idx := len(iKey) - 1
		tag := iKey[idx] // the last byte of key
		sid := binary.BigEndian.Uint64(iKey[idx-8 : idx])
		idx -= 8
		timeBytes := iKey[idx-8 : idx]
		timeNum := binary.BigEndian.Uint64(timeBytes)
		entry := json.RawMessage(iter.Value())
		if filter != nil && !filter(tag, entry) {
			continue
		}
		if fetchTxDetail { // iter.Value is not desired data, it's just tx_hash which points to the desired data
			hexTxHashID := getTxHashID(iter.Value())
			key := append([]byte{DetailByte}, hexTxHashID...)
			key = append(key, timeBytes...)
			entry = hub.db.Get(key)
		}
		if firstByteIn == DelistsByte {
			data = append(data, iKey[2:iKey[1]+2]) // length = iKey[1], market name = iKey[2:length+2]
		}
		data = append(data, entry)
		tags = append(tags, tag)
		timesid = append(timesid, []int64{int64(timeNum), int64(sid)}...)
		if firstByteIn == 0 ||
			(firstByteIn == CreateOrderOnlyByte && tag == CreateOrderEndByte) ||
			(firstByteIn == FillOrderOnlyByte && tag == FillOrderEndByte) ||
			(firstByteIn == CancelOrderOnlyByte && tag == CancelOrderEndByte) {
			count--
		}
		if count == 0 {
			break
		}
	}
	return
}

func getTxHashID(v []byte) []byte {
	for i := len(v) - 1; i >= 0; i-- {
		if v[i] == byte('|') {
			return v[i+1:]
		}
	}
	return []byte{}
}
