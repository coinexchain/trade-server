package core

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	log "github.com/sirupsen/logrus"
	dbm "github.com/tendermint/tm-db"
)

const (
	MaxCount = 1024
	//These bytes are used as the first byte in key
	CandleStickByte  = byte(0x10)
	DealByte         = byte(0x12)
	OrderByte        = byte(0x14)
	BancorInfoByte   = byte(0x16)
	BancorTradeByte  = byte(0x18)
	IncomeByte       = byte(0x1A)
	TxByte           = byte(0x1C)
	CommentByte      = byte(0x1E)
	BlockHeightByte  = byte(0x20)
	DetailByte       = byte(0x22)
	SlashByte        = byte(0x24)
	RedelegationByte = byte(0x30)
	UnbondingByte    = byte(0x32)
	UnlockByte       = byte(0x34)
	OffsetByte       = byte(0xF0)
	LockedByte       = byte(0x36)
)

func limitCount(count int) int {
	if count > MaxCount {
		return MaxCount
	}
	return count
}

func int64ToBigEndianBytes(n int64) []byte {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(n))
	return b[:]
}

// Following are some functions to generate keys to access the KVStore
func (hub *Hub) getKeyFromBytesAndTime(firstByte byte, bz []byte, lastByte byte, unixTime int64) []byte {
	res := make([]byte, 0, 1+1+len(bz)+1+16+1)
	res = append(res, firstByte)
	res = append(res, byte(len(bz)))
	res = append(res, bz...)
	res = append(res, byte(0))
	res = append(res, int64ToBigEndianBytes(unixTime)...) //the block's time at which the KV pair is generated
	res = append(res, int64ToBigEndianBytes(hub.sid)...)  // the serial ID for a KV pair
	res = append(res, lastByte)
	return res
}

func (hub *Hub) getKeyFromBytes(firstByte byte, bz []byte, lastByte byte) []byte {
	return hub.getKeyFromBytesAndTime(firstByte, bz, lastByte, hub.currBlockTime.Unix())
}

func getStartKeyFromBytes(firstByte byte, bz []byte) []byte {
	res := make([]byte, 0, 1+1+len(bz)+1)
	res = append(res, firstByte)
	res = append(res, byte(len(bz)))
	res = append(res, bz...)
	res = append(res, byte(0))
	return res
}

func getEndKeyFromBytes(firstByte byte, bz []byte, time int64, sid int64) []byte {
	res := make([]byte, 0, 1+1+len(bz)+1+16)
	res = append(res, firstByte)
	res = append(res, byte(len(bz)))
	res = append(res, bz...)
	res = append(res, byte(0))
	res = append(res, int64ToBigEndianBytes(time)...)
	res = append(res, int64ToBigEndianBytes(sid)...)
	return res
}

//==========

func (hub *Hub) getCandleStickKey(market string, timespan byte) []byte {
	bz := append([]byte(market), []byte{0, timespan}...)
	return hub.getKeyFromBytes(CandleStickByte, bz, 0)
}

func (hub *Hub) getLockedKey(addr string) []byte {
	return hub.getKeyFromBytes(LockedByte, []byte(addr), 0)
}

func getCandleStickEndKey(market string, timespan byte, endTime int64, sid int64) []byte {
	bz := append([]byte(market), []byte{0, timespan}...)
	return getEndKeyFromBytes(CandleStickByte, bz, endTime, sid)
}

func getCandleStickStartKey(market string, timespan byte) []byte {
	bz := append([]byte(market), []byte{0, timespan}...)
	return getStartKeyFromBytes(CandleStickByte, bz)
}

func (hub *Hub) getDealKey(market string) []byte {
	return hub.getKeyFromBytes(DealByte, []byte(market), 0)
}
func (hub *Hub) getBancorInfoKey(market string) []byte {
	return hub.getKeyFromBytes(BancorInfoByte, []byte(market), 0)
}
func (hub *Hub) getCommentKey(token string) []byte {
	return hub.getKeyFromBytes(CommentByte, []byte(token), 0)
}

func (hub *Hub) getCreateOrderKey(addr string) []byte {
	return hub.getKeyFromBytes(OrderByte, []byte(addr), CreateOrderEndByte)
}
func (hub *Hub) getFillOrderKey(addr string) []byte {
	return hub.getKeyFromBytes(OrderByte, []byte(addr), FillOrderEndByte)
}
func (hub *Hub) getCancelOrderKey(addr string) []byte {
	return hub.getKeyFromBytes(OrderByte, []byte(addr), CancelOrderEndByte)
}
func (hub *Hub) getBancorTradeKey(addr string) []byte {
	return hub.getKeyFromBytes(BancorTradeByte, []byte(addr), byte(0))
}
func (hub *Hub) getIncomeKey(addr string) []byte {
	return hub.getKeyFromBytes(IncomeByte, []byte(addr), byte(0))
}
func (hub *Hub) getTxKey(addr string) []byte {
	return hub.getKeyFromBytes(TxByte, []byte(addr), byte(0))
}
func (hub *Hub) getRedelegationEventKey(addr string, time int64) []byte {
	return hub.getKeyFromBytesAndTime(RedelegationByte, []byte(addr), byte(0), time)
}
func (hub *Hub) getUnbondingEventKey(addr string, time int64) []byte {
	return hub.getKeyFromBytesAndTime(UnbondingByte, []byte(addr), byte(0), time)
}
func (hub *Hub) getUnlockEventKey(addr string) []byte {
	return hub.getKeyFromBytes(UnlockByte, []byte(addr), byte(0))
}

type TripleManager struct {
	sell *DepthManager
	buy  *DepthManager
	tkm  *TickerManager
}

type Hub struct {
	// the serial ID for a KV pair in KVStore
	sid int64
	// KVStore and its batch
	db    dbm.DB
	batch dbm.Batch
	// Mutex to protect shared storage and variables
	dbMutex        sync.RWMutex
	tickerMapMutex sync.RWMutex
	depthMutex     sync.RWMutex

	csMan CandleStickManager

	// Updating logic and query logic share these variables
	managersMap map[string]TripleManager
	tickerMap   map[string]*Ticker

	// interface to the subscribe functions
	subMan SubscribeManager

	currBlockTime time.Time
	lastBlockTime time.Time

	// cache for NotificationSlash
	slashSlice []*NotificationSlash
}

func NewHub(db dbm.DB, subMan SubscribeManager) Hub {
	return Hub{
		db:            db,
		batch:         db.NewBatch(),
		subMan:        subMan,
		managersMap:   make(map[string]TripleManager),
		csMan:         NewCandleStickManager(nil),
		currBlockTime: time.Unix(0, 0),
		lastBlockTime: time.Unix(0, 0),
		tickerMap:     make(map[string]*Ticker),
		slashSlice:    make([]*NotificationSlash, 0, 10),
	}
}

func (hub *Hub) HasMarket(market string) bool {
	_, ok := hub.managersMap[market]
	return ok
}

func (hub *Hub) AddMarket(market string) {
	hub.managersMap[market] = TripleManager{
		sell: DefaultDepthManager(),
		buy:  DefaultDepthManager(),
		tkm:  DefaultTickerManager(market),
	}
	hub.csMan.AddMarket(market)
}

func (hub *Hub) Log(s string) {
	log.Error(s)
}

//============================================================
var _ Consumer = &Hub{}

func (hub *Hub) ConsumeMessage(msgType string, bz []byte) {
	switch msgType {
	case "height_info":
		hub.handleNewHeightInfo(bz)
	case "notify_slash":
		hub.handleNotificationSlash(bz)
	case "notify_tx":
		hub.handleNotificationTx(bz)
	case "begin_redelegation":
		hub.handleNotificationBeginRedelegation(bz)
	case "begin_unbonding":
		hub.handleNotificationBeginUnbonding(bz)
	case "complete_redelegation":
		hub.handleNotificationCompleteRedelegation(bz)
	case "complete_unbonding":
		hub.handleNotificationCompleteUnbonding(bz)
	case "notify_unlock":
		hub.handleNotificationUnlock(bz)
	case "token_comment":
		hub.handleTokenComment(bz)
	case "create_order_info":
		hub.handleCreateOrderInfo(bz)
	case "fill_order_info":
		hub.handleFillOrderInfo(bz)
	case "del_order_info":
		hub.handleCancelOrderInfo(bz)
	case "bancor_trade":
		hub.handleMsgBancorTradeInfoForKafka(bz)
	case "bancor_info":
		hub.handleMsgBancorInfoForKafka(bz)
	case "commit":
		hub.commit()
	case "send_lock_coins":
		hub.handleLockedCoinsMsg(bz)
	default:
		hub.Log(fmt.Sprintf("Unknown Message Type:%s", msgType))
	}
}

func (hub *Hub) handleNewHeightInfo(bz []byte) {
	var v NewHeightInfo
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal NewHeightInfo")
		return
	}
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b[:], uint64(v.TimeStamp.Unix()))
	key := append([]byte{BlockHeightByte}, int64ToBigEndianBytes(v.Height)...)
	hub.batch.Set(key, b)

	for _, ss := range hub.subMan.GetHeightSubscribeInfo() {
		hub.subMan.PushHeight(ss, bz)
	}

	hub.lastBlockTime = hub.currBlockTime
	hub.currBlockTime = v.TimeStamp
	hub.beginForCandleSticks()
}

func (hub *Hub) beginForCandleSticks() {
	candleSticks := hub.csMan.NewBlock(hub.currBlockTime)
	var triman TripleManager
	var targets []Subscriber
	sym := ""
	var ok bool
	currMinute := hub.currBlockTime.Hour() * hub.currBlockTime.Minute()
	for _, cs := range candleSticks {
		if sym != cs.Market {
			triman, ok = hub.managersMap[cs.Market]
			if !ok {
				sym = ""
				continue
			}
			info := hub.subMan.GetCandleStickSubscribeInfo()
			if info == nil {
				targets = []Subscriber{}
			} else {
				sym = cs.Market
				targets, ok = info[sym]
				if !ok {
					targets = []Subscriber{}
				}
			}
		}
		if len(sym) == 0 {
			continue
		}
		// Update tickers' prices
		if cs.TimeSpan == MinuteStr {
			triman.tkm.UpdateNewestPrice(cs.ClosePrice, currMinute)
		}
		bz := formatCandleStick(&cs)
		if bz == nil {
			continue
		}
		// Push candle sticks to subscribers
		for _, target := range targets {
			timespan, ok := target.Detail().(string)
			if !ok || timespan != cs.TimeSpan {
				continue
			}
			hub.subMan.PushCandleStick(target, bz)
		}
		// Save candle sticks to KVStore
		key := hub.getCandleStickKey(cs.Market, getSpanFromSpanStr(cs.TimeSpan))
		if len(bz) == 0 {
			continue
		}
		//fmt.Printf("Here4! %v\n", cs)
		hub.batch.Set(key, bz)
		hub.sid++
	}
}

func formatCandleStick(info *CandleStick) []byte {
	bz, err := json.Marshal(info)
	if err != nil {
		log.Errorf(err.Error())
		return nil
	}
	return bz
}

func (hub *Hub) handleNotificationSlash(bz []byte) {
	var v NotificationSlash
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal NotificationSlash")
		return
	}
	hub.slashSlice = append(hub.slashSlice, &v)
}

func (hub *Hub) handleLockedCoinsMsg(bz []byte) {
	var v LockedSendMsg
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Err in Unmarshal LockedSendMsg")
	}
	key := hub.getLockedKey(v.ToAddress)
	hub.batch.Set(key, bz)
	hub.sid++
	infos := hub.subMan.GetLockedSubscribeInfo()
	if conns, ok := infos[v.ToAddress]; ok {
		for _, c := range conns {
			hub.subMan.PushLockedSendMsg(c, bz)
		}
	}
}

func (hub *Hub) handleNotificationTx(bz []byte) {
	var v NotificationTx
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log(fmt.Sprintf("Error in Unmarshal NotificationTx: %s", string(bz)))
		return
	}
	snBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(snBytes[:], uint64(v.SerialNumber))

	// Use the transaction's serial number as key, save its detail
	key := append([]byte{DetailByte}, snBytes...)
	hub.batch.Set(key, bz)
	hub.sid++

	for _, acc := range v.Signers {
		signer := acc
		k := hub.getTxKey(signer)
		hub.batch.Set(k, snBytes)
		hub.sid++

		info := hub.subMan.GetTxSubscribeInfo()
		targets, ok := info[signer]
		if !ok {
			continue
		}
		for _, target := range targets {
			hub.subMan.PushTx(target, bz)
		}
	}

	for _, transRec := range v.Transfers {
		recipient := transRec.Recipient
		k := hub.getIncomeKey(recipient)
		hub.batch.Set(k, snBytes)
		hub.sid++

		info := hub.subMan.GetIncomeSubscribeInfo()
		targets, ok := info[recipient]
		if !ok {
			continue
		}
		for _, target := range targets {
			hub.subMan.PushIncome(target, bz)
		}
	}
}
func (hub *Hub) handleNotificationBeginRedelegation(bz []byte) {
	var v NotificationBeginRedelegation
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal NotificationBeginRedelegation")
		return
	}
	t, err := time.Parse(time.RFC3339, v.CompletionTime)
	if err != nil {
		hub.Log("Error in Parsing Time")
		return
	}
	// Use completion time as the key
	key := hub.getRedelegationEventKey(v.Delegator, t.Unix())
	hub.batch.Set(key, bz)
	hub.sid++
}
func (hub *Hub) handleNotificationBeginUnbonding(bz []byte) {
	var v NotificationBeginUnbonding
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal NotificationBeginUnbonding")
		return
	}
	t, err := time.Parse(time.RFC3339, v.CompletionTime)
	if err != nil {
		hub.Log("Error in Parsing Time")
		return
	}
	// Use completion time as the key
	key := hub.getUnbondingEventKey(v.Delegator, t.Unix())
	hub.batch.Set(key, bz)
	hub.sid++
}
func (hub *Hub) handleNotificationCompleteRedelegation(bz []byte) {
	var v NotificationCompleteRedelegation
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal NotificationCompleteRedelegation")
		return
	}
	info := hub.subMan.GetRedelegationSubscribeInfo()
	targets, ok := info[v.Delegator]
	if !ok {
		return
	}
	// query the redelegations whose completion time is between current block and last block
	end := hub.getRedelegationEventKey(v.Delegator, hub.currBlockTime.Unix())
	start := hub.getRedelegationEventKey(v.Delegator, hub.lastBlockTime.Unix()-1)
	hub.dbMutex.RLock()
	iter := hub.db.ReverseIterator(start, end)
	defer func() {
		iter.Close()
		hub.dbMutex.RUnlock()
	}()
	for ; iter.Valid(); iter.Next() {
		for _, target := range targets {
			hub.subMan.PushRedelegation(target, iter.Value())
		}
	}
}
func (hub *Hub) handleNotificationCompleteUnbonding(bz []byte) {
	var v NotificationCompleteUnbonding
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal NotificationCompleteUnbonding")
		return
	}
	info := hub.subMan.GetUnbondingSubscribeInfo()
	targets, ok := info[v.Delegator]
	if !ok {
		return
	}
	// query the unbondings whose completion time is between current block and last block
	end := hub.getUnbondingEventKey(v.Delegator, hub.currBlockTime.Unix())
	start := hub.getUnbondingEventKey(v.Delegator, hub.lastBlockTime.Unix()-1)
	hub.dbMutex.RLock()
	iter := hub.db.ReverseIterator(start, end)
	defer func() {
		iter.Close()
		hub.dbMutex.RUnlock()
	}()
	for ; iter.Valid(); iter.Next() {
		for _, target := range targets {
			hub.subMan.PushUnbonding(target, iter.Value())
		}
	}
}
func (hub *Hub) handleNotificationUnlock(bz []byte) {
	var v NotificationUnlock
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal NotificationUnlock")
		return
	}
	addr := v.Address
	key := hub.getUnlockEventKey(addr)
	hub.batch.Set(key, bz)
	hub.sid++
	info := hub.subMan.GetUnlockSubscribeInfo()
	targets, ok := info[addr]
	if !ok {
		return
	}
	for _, target := range targets {
		hub.subMan.PushUnlock(target, bz)
	}
}
func (hub *Hub) handleTokenComment(bz []byte) {
	var v TokenComment
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal TokenComment")
		return
	}
	key := hub.getCommentKey(v.Token)
	hub.batch.Set(key, bz)
	hub.sid++
	info := hub.subMan.GetCommentSubscribeInfo()
	targets, ok := info[v.Token]
	if !ok {
		return
	}
	for _, target := range targets {
		hub.subMan.PushComment(target, bz)
	}
}
func (hub *Hub) handleCreateOrderInfo(bz []byte) {
	var v CreateOrderInfo
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal CreateOrderInfo")
		return
	}
	// Add a new market which is seen for the first time
	if !hub.HasMarket(v.TradingPair) {
		hub.AddMarket(v.TradingPair)
	}
	//Save to KVStore
	key := hub.getCreateOrderKey(v.Sender)
	hub.batch.Set(key, bz)
	hub.sid++
	//Push to subscribers
	info := hub.subMan.GetOrderSubscribeInfo()
	targets, ok := info[v.Sender]
	if ok {
		for _, target := range targets {
			hub.subMan.PushCreateOrder(target, bz)
		}
	}
	//Update depth info
	triman, ok := hub.managersMap[v.TradingPair]
	if !ok {
		return
	}
	amount := sdk.NewInt(v.Quantity)
	hub.depthMutex.Lock()
	defer func() {
		hub.depthMutex.Unlock()
	}()
	if v.Side == SELL {
		triman.sell.DeltaChange(v.Price, amount)
	} else {
		triman.buy.DeltaChange(v.Price, amount)
	}
}
func (hub *Hub) handleFillOrderInfo(bz []byte) {
	var v FillOrderInfo
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal FillOrderInfo")
		return
	}
	if v.DealStock == 0 {
		return
	}
	// Add a new market which is seen for the first time
	if !hub.HasMarket(v.TradingPair) {
		hub.AddMarket(v.TradingPair)
	}
	//Save to KVStore
	accAndSeq := strings.Split(v.OrderID, "-")
	if len(accAndSeq) != 2 {
		return
	}
	key := hub.getFillOrderKey(accAndSeq[0])
	hub.batch.Set(key, bz)
	hub.sid++
	key = hub.getDealKey(v.TradingPair)
	hub.batch.Set(key, bz)
	hub.sid++
	//Push to subscribers
	info := hub.subMan.GetOrderSubscribeInfo()
	targets, ok := info[accAndSeq[0]]
	if ok {
		for _, target := range targets {
			hub.subMan.PushFillOrder(target, bz)
		}
	}
	info = hub.subMan.GetDealSubscribeInfo()
	targets, ok = info[v.TradingPair]
	if ok {
		for _, target := range targets {
			hub.subMan.PushDeal(target, bz)
		}
	}
	//Update candle sticks
	csRec := hub.csMan.GetRecord(v.TradingPair)
	if csRec != nil {
		price := sdk.NewDec(v.DealMoney).QuoInt64(v.DealStock)
		csRec.Update(hub.currBlockTime, price, v.DealStock)
	}
	//Update depth info
	triman, ok := hub.managersMap[v.TradingPair]
	if !ok {
		return
	}
	negStock := sdk.NewInt(-v.DealStock)
	hub.depthMutex.Lock()
	defer func() {
		hub.depthMutex.Unlock()
	}()
	if v.Side == SELL {
		triman.sell.DeltaChange(v.Price, negStock)
	} else {
		triman.buy.DeltaChange(v.Price, negStock)
	}
}

func (hub *Hub) handleCancelOrderInfo(bz []byte) {
	var v CancelOrderInfo
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal CancelOrderInfo")
		return
	}
	// Add a new market which is seen for the first time
	if !hub.HasMarket(v.TradingPair) {
		hub.AddMarket(v.TradingPair)
	}
	//Save to KVStore
	accAndSeq := strings.Split(v.OrderID, "-")
	if len(accAndSeq) != 2 {
		return
	}
	key := hub.getCancelOrderKey(accAndSeq[0])
	hub.batch.Set(key, bz)
	hub.sid++
	//Push to subscribers
	info := hub.subMan.GetOrderSubscribeInfo()
	targets, ok := info[accAndSeq[0]]
	if !ok {
		return
	}
	for _, target := range targets {
		hub.subMan.PushCancelOrder(target, bz)
	}
	//Update depth info
	triman, ok := hub.managersMap[v.TradingPair]
	if !ok {
		return
	}
	negStock := sdk.NewInt(-v.LeftStock)
	hub.depthMutex.Lock()
	defer func() {
		hub.depthMutex.Unlock()
	}()
	if v.Side == SELL {
		triman.sell.DeltaChange(v.Price, negStock)
	} else {
		triman.buy.DeltaChange(v.Price, negStock)
	}
}

func (hub *Hub) handleMsgBancorTradeInfoForKafka(bz []byte) {
	var v MsgBancorTradeInfoForKafka
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal MsgBancorTradeInfoForKafka")
		return
	}
	//Save to KVStore
	addr := v.Sender
	key := hub.getBancorTradeKey(addr)
	hub.batch.Set(key, bz)
	hub.sid++
	//Push to subscribers
	info := hub.subMan.GetBancorTradeSubscribeInfo()
	targets, ok := info[addr]
	if !ok {
		return
	}
	for _, target := range targets {
		hub.subMan.PushBancorTrade(target, bz)
	}
}

func (hub *Hub) handleMsgBancorInfoForKafka(bz []byte) {
	var v MsgBancorInfoForKafka
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal MsgBancorInfoForKafka")
		return
	}
	//Save to KVStore
	key := hub.getBancorInfoKey(v.Stock + "/" + v.Money)
	hub.batch.Set(key, bz)
	hub.sid++
	//Push to subscribers
	info := hub.subMan.GetBancorInfoSubscribeInfo()
	targets, ok := info[v.Stock+"/"+v.Money]
	if !ok {
		return
	}
	for _, target := range targets {
		hub.subMan.PushBancorInfo(target, bz)
	}
}

func (hub *Hub) commitForSlash() {
	newSlice := make([]*NotificationSlash, 0, len(hub.slashSlice))
	// To fix a bug of cosmos-sdk
	for _, slash := range hub.slashSlice {
		if len(slash.Validator) == 0 && len(newSlice) != 0 {
			newSlice[len(newSlice)-1].Jailed = slash.Jailed
		}
		if len(slash.Validator) != 0 {
			newSlice = append(newSlice, slash)
		}
	}
	for _, slash := range newSlice {
		bz, _ := json.Marshal(slash)
		for _, ss := range hub.subMan.GetSlashSubscribeInfo() {
			hub.subMan.PushSlash(ss, bz)
		}
		key := hub.getKeyFromBytes(SlashByte, []byte{}, 0)
		hub.batch.Set(key, bz)
		hub.sid++
	}
	hub.slashSlice = hub.slashSlice[:0]
}

func (hub *Hub) commitForTicker() {
	tkMap := make(map[string]*Ticker)
	currMinute := hub.currBlockTime.Hour() * hub.currBlockTime.Minute()
	for _, triman := range hub.managersMap {
		if ticker := triman.tkm.GetTicker(currMinute); ticker != nil {
			tkMap[ticker.Market] = ticker
		}
	}
	for _, subscriber := range hub.subMan.GetTickerSubscribeInfo() {
		marketList := subscriber.Detail().(map[string]struct{})
		tickerList := make([]*Ticker, 0, len(marketList))
		for market := range marketList {
			if ticker, ok := tkMap[market]; ok {
				tickerList = append(tickerList, ticker)
			}
		}
		if len(tickerList) != 0 {
			hub.subMan.PushTicker(subscriber, tickerList)
		}
	}

	hub.tickerMapMutex.Lock()
	for market, ticker := range tkMap {
		hub.tickerMap[market] = ticker
	}
	hub.tickerMapMutex.Unlock()
}

func (hub *Hub) commitForDepth() {
	hub.depthMutex.Lock()
	defer func() {
		hub.depthMutex.Unlock()
	}()
	for market, triman := range hub.managersMap {
		depthDeltaSell := triman.sell.EndBlock()
		depthDeltaBuy := triman.buy.EndBlock()
		if len(depthDeltaSell) == 0 && len(depthDeltaBuy) == 0 {
			continue
		}
		info := hub.subMan.GetDepthSubscribeInfo()
		targets, ok := info[market]
		if !ok {
			continue
		}

		buyBz := encodeDepth(market, depthDeltaBuy, true)
		sellBz := encodeDepth(market, depthDeltaSell, false)
		for _, target := range targets {
			if len(depthDeltaSell) != 0 {
				hub.subMan.PushDepthSell(target, sellBz)
			}
			if len(depthDeltaBuy) != 0 {
				hub.subMan.PushDepthBuy(target, buyBz)
			}
		}
	}
}

func (hub *Hub) commit() {
	hub.commitForSlash()
	hub.commitForTicker()
	hub.commitForDepth()
	hub.dbMutex.Lock()
	hub.batch.WriteSync()
	hub.batch.Close()
	hub.batch = hub.db.NewBatch()
	hub.dbMutex.Unlock()
}

//============================================================
var _ Querier = &Hub{}

func (hub *Hub) QueryTickers(marketList []string) []*Ticker {
	tickerList := make([]*Ticker, 0, len(marketList))
	hub.tickerMapMutex.RLock()
	for _, market := range marketList {
		ticker, ok := hub.tickerMap[market]
		if ok {
			tickerList = append(tickerList, ticker)
		}
	}
	hub.tickerMapMutex.RUnlock()
	return tickerList
}

func (hub *Hub) QueryBlockTime(height int64, count int) []int64 {
	count = limitCount(count)
	data := make([]int64, 0, count)
	end := append([]byte{BlockHeightByte}, int64ToBigEndianBytes(height)...)
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
	hub.depthMutex.RLock()
	sell = tripleMan.sell.GetLowest(count)
	buy = tripleMan.buy.GetHighest(count)
	hub.depthMutex.RUnlock()
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

//=========
func (hub *Hub) QueryOrder(account string, time int64, sid int64, count int) (data []json.RawMessage, tags []byte, timesid []int64) {
	return hub.query(false, OrderByte, []byte(account), time, sid, count, nil)
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

// --------------

func (hub *Hub) QueryOrderAboutToken(token, account string, time int64, sid int64, count int) (data []json.RawMessage, tags []byte, timesid []int64) {
	return hub.query(false, OrderByte, []byte(account), time, sid, count, func(tag byte, entry []byte) bool {
		s1 := fmt.Sprintf("/%s\",\"height\":", token)
		s2 := fmt.Sprintf("\"trading_pair\":\"%s/", token)
		if tag == CreateOrderEndByte {
			s1 = fmt.Sprintf("/%s\",\"order_type\":", token)
		}
		return strings.Index(string(entry), s1) > 0 || strings.Index(string(entry), s2) > 0
	})
}

func (hub *Hub) QueryLockedAboutToken(token, account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	data, _, timesid = hub.query(false, LockedByte, []byte(account), time, sid, count, func(tag byte, entry []byte) bool {
		s := fmt.Sprintf("\"denom\":\"%s\",\"amount\":", token)
		return strings.Index(string(entry), s) > 0
	})
	return
}

func (hub *Hub) QueryBancorTradeAboutToken(token, account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	data, _, timesid = hub.query(false, BancorTradeByte, []byte(account), time, sid, count, func(tag byte, entry []byte) bool {
		s1 := fmt.Sprintf("\"money\":\"%s\"", token)
		s2 := fmt.Sprintf("\"stock\":\"%s\"", token)
		return strings.Index(string(entry), s1) > 0 || strings.Index(string(entry), s2) > 0
	})
	return
}

func (hub *Hub) QueryUnlockAboutToken(token, account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	data, _, timesid = hub.query(false, UnlockByte, []byte(account), time, sid, count, func(tag byte, entry []byte) bool {
		s := fmt.Sprintf("\"unlocked\":[{\"denom\":\"%s\",\"amount\":", token)
		return strings.Index(string(entry), s) > 0
	})
	return
}

func (hub *Hub) QueryIncomeAboutToken(token, account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	r := regexp.MustCompile(fmt.Sprintf("\"amount\":\"[0-9]+%s", token))
	data, _, timesid = hub.query(true, IncomeByte, []byte(account), time, sid, count, func(tag byte, entry []byte) bool {
		return r.MatchString(string(entry))
	})
	return
}

func (hub *Hub) QueryTxAboutToken(token, account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	r := regexp.MustCompile(fmt.Sprintf("\"amount\":\"[0-9]+%s", token))
	data, _, timesid = hub.query(true, TxByte, []byte(account), time, sid, count, func(tag byte, entry []byte) bool {
		return r.MatchString(string(entry))
	})
	return
}

type filterFunc func(tag byte, entry []byte) bool

func (hub *Hub) query(fetchTxDetail bool, firstByte byte, bz []byte, time int64, sid int64,
	count int, filter filterFunc) (data []json.RawMessage, tags []byte, timesid []int64) {
	count = limitCount(count)
	data = make([]json.RawMessage, 0, count)
	tags = make([]byte, 0, count)
	timesid = make([]int64, 0, 2*count)
	start := getStartKeyFromBytes(firstByte, bz)
	end := getEndKeyFromBytes(firstByte, bz, time, sid)
	hub.dbMutex.RLock()
	iter := hub.db.ReverseIterator(start, end)
	defer func() {
		iter.Close()
		hub.dbMutex.RUnlock()
	}()
	for ; iter.Valid(); iter.Next() {
		iKey := iter.Key()
		idx := len(iKey) - 1
		tag := iKey[idx]
		sid := binary.BigEndian.Uint64(iKey[idx-8 : idx])
		idx -= 8
		time := binary.BigEndian.Uint64(iKey[idx-8 : idx])
		entry := json.RawMessage(iter.Value())
		if fetchTxDetail {
			key := append([]byte{DetailByte}, iter.Value()...)
			entry = json.RawMessage(hub.db.Get(key))
		}
		if filter != nil && filter(tag, entry) {
			continue
		}
		data = append(data, entry)
		tags = append(tags, tag)
		timesid = append(timesid, []int64{int64(time), int64(sid)}...)
		if count--; count == 0 {
			break
		}
	}
	return
}

//===================================
// for serialization and deserialization of Hub

type HubForJSON struct {
	Sid           int64              `json:"sid"`
	CSMan         CandleStickManager `json:"csman"`
	TickerMap     map[string]*Ticker
	CurrBlockTime time.Time            `json:"curr_block_time"`
	LastBlockTime time.Time            `json:"last_block_time"`
	Markets       []*MarketInfoForJSON `json:"markets"`
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
	hub.currBlockTime = hub4j.CurrBlockTime
	hub.lastBlockTime = hub4j.LastBlockTime

	for _, info := range hub4j.Markets {
		triman := TripleManager{
			sell: DefaultDepthManager(),
			buy:  DefaultDepthManager(),
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
	hub4j.CurrBlockTime = hub.currBlockTime
	hub4j.LastBlockTime = hub.lastBlockTime

	hub4j.Markets = make([]*MarketInfoForJSON, 0, len(hub.managersMap))
	for _, triman := range hub.managersMap {
		hub4j.Markets = append(hub4j.Markets, &MarketInfoForJSON{
			TkMan:           triman.tkm,
			SellPricePoints: triman.sell.DumpPricePoints(),
			BuyPricePoints:  triman.buy.DumpPricePoints(),
		})
	}
}

//===================================
func (hub *Hub) UpdateOffset(partition int32, offset int64) {
	key := getOffsetKey(partition)
	offsetBuf := int64ToBigEndianBytes(offset)
	hub.batch.Set(key, offsetBuf)
}

func (hub *Hub) LoadOffset(partition int32) int64 {
	key := getOffsetKey(partition)
	hub.dbMutex.RLock()
	defer func() {
		hub.dbMutex.RUnlock()
	}()
	offsetBuf := hub.db.Get(key)
	if offsetBuf == nil {
		return 0
	}
	return int64(binary.BigEndian.Uint64(offsetBuf))
}

func getOffsetKey(partition int32) []byte {
	key := make([]byte, 5)
	key[0] = OffsetByte
	binary.BigEndian.PutUint32(key[1:], uint32(partition))
	return key
}

func (hub *Hub) Close() {
	hub.db.Close()
}
