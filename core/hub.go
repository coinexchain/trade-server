package core

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/coinexchain/dex/app"
	"github.com/coinexchain/dex/modules/authx"
	"github.com/coinexchain/dex/modules/bancorlite"
	"github.com/coinexchain/dex/modules/comment"
	"github.com/coinexchain/dex/modules/market"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dbm "github.com/tendermint/tm-db"
)

const (
	MaxCount         = 1024
	CandleStickByte  = byte(0x10)
	DealByte         = byte(0x12)
	OrderByte        = byte(0x14)
	BancorInfoByte   = byte(0x16)
	BancorTradeByte  = byte(0x18)
	IncomeByte       = byte(0x1A)
	TxByte           = byte(0x1C)
	CommentByte      = byte(0x1D)
	BlockHeightByte  = byte(0x20)
	DetailByte       = byte(0x22)
	RedelegationByte = byte(0x30)
	UnbondingByte    = byte(0x32)
	UnlockByte       = byte(0x34)

	CreateOrderEndByte = byte('c')
	FillOrderEndByte   = byte('f')
	CancelOrderEndByte = byte('d')
)

func limitCount(count int) int {
	if count > MaxCount {
		return MaxCount
	}
	return count
}

func int64ToBigEndianBytes(n int64) []byte {
	var result [8]byte
	for i := 0; i < 8; i++ {
		result[i] = byte(n >> (8 * uint(i)))
	}
	return result[:]
}

func (hub *Hub) getCandleStickKey(market string, timespan byte) []byte {
	res := make([]byte, 0, 1+len(market)+3+8)
	res = append(res, CandleStickByte)
	res = append(res, []byte(market)...)
	res = append(res, []byte{0, timespan, 0}...)
	res = append(res, int64ToBigEndianBytes(hub.currBlockTime.Unix())...)
	return res
}

func getCandleStickKeyEnd(market string, timespan byte, endTime int64) []byte {
	res := make([]byte, 0, 1+len(market)+3+8)
	res = append(res, CandleStickByte)
	res = append(res, []byte(market)...)
	res = append(res, []byte{0, timespan, 0}...)
	res = append(res, int64ToBigEndianBytes(endTime)...)
	return res
}

func getCandleStickKeyStart(market string, timespan byte) []byte {
	res := make([]byte, 0, 1+len(market)+3)
	res = append(res, CandleStickByte)
	res = append(res, []byte(market)...)
	res = append(res, []byte{0, timespan, 0}...)
	return res
}

func (hub *Hub) getKeyFromBytes(firstByte byte, bz []byte, lastByte byte) []byte {
	res := make([]byte, 0, 1+1+len(bz)+1+16+1)
	res = append(res, firstByte)
	res = append(res, byte(len(bz)))
	res = append(res, bz...)
	res = append(res, byte(0))
	res = append(res, int64ToBigEndianBytes(hub.currBlockTime.Unix())...)
	res = append(res, int64ToBigEndianBytes(hub.sid)...)
	res = append(res, lastByte)
	return res
}

func (hub *Hub) getKeyFromBytesAndTime(firstByte byte, bz []byte, lastByte byte, unixTime int64) []byte {
	res := make([]byte, 0, 1+1+len(bz)+1+16+1)
	res = append(res, firstByte)
	res = append(res, byte(len(bz)))
	res = append(res, bz...)
	res = append(res, byte(0))
	res = append(res, int64ToBigEndianBytes(unixTime)...)
	res = append(res, int64ToBigEndianBytes(hub.sid)...)
	res = append(res, lastByte)
	return res
}

func getStartKeyFromBytes(firstByte byte, bz []byte) []byte {
	res := make([]byte, 0, 1+1+len(bz)+1)
	res = append(res, firstByte)
	res = append(res, byte(len(bz)))
	res = append(res, bz...)
	res = append(res, byte(0))
	return res
}

func getEndKeyFromBytes(firstByte byte, bz []byte, unixSec int64) []byte {
	res := make([]byte, 0, 1+1+len(bz)+1+8)
	res = append(res, firstByte)
	res = append(res, byte(len(bz)))
	res = append(res, bz...)
	res = append(res, byte(0))
	res = append(res, int64ToBigEndianBytes(unixSec)...)
	return res
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

type (
	NewHeightInfo                    = market.NewHeightInfo
	CreateOrderInfo                  = market.CreateOrderInfo
	FillOrderInfo                    = market.FillOrderInfo
	CancelOrderInfo                  = market.CancelOrderInfo
	NotificationSlash                = app.NotificationSlash
	TransferRecord                   = app.TransferRecord
	NotificationTx                   = app.NotificationTx
	NotificationBeginRedelegation    = app.NotificationBeginRedelegation
	NotificationBeginUnbonding       = app.NotificationBeginUnbonding
	NotificationCompleteRedelegation = app.NotificationCompleteRedelegation
	NotificationCompleteUnbonding    = app.NotificationCompleteUnbonding
	NotificationUnlock               = authx.NotificationUnlock
	TokenComment                     = comment.TokenComment
	CommentRef                       = comment.CommentRef
	MsgBancorTradeInfoForKafka       = bancorlite.MsgBancorTradeInfoForKafka
	MsgBancorInfoForKafka            = bancorlite.MsgBancorInfoForKafka
)

type Subscriber interface {
	Detail() interface{}
}

type SubscribeManager interface {
	GetSlashSubscribeInfo() []Subscriber
	GetHeightSubscribeInfo() []Subscriber
	GetTickerSubscribeInfo() []Subscriber
	GetCandleStickSubscribeInfo() map[string][]Subscriber

	GetDepthSubscribeInfo() map[string][]Subscriber
	GetDealSubscribeInfo() map[string][]Subscriber
	GetOrderSubscribeInfo() map[string][]Subscriber
	GetBancorInfoSubscribeInfo() map[string][]Subscriber
	GetBancorTradeSubscribeInfo() map[string][]Subscriber
	GetIncomeSubscribeInfo() map[string][]Subscriber
	GetUnbondingSubscribeInfo() map[string][]Subscriber
	GetRedelegationSubscribeInfo() map[string][]Subscriber
	GetUnlockSubscribeInfo() map[string][]Subscriber
	GetTxSubscribeInfo() map[string][]Subscriber
	GetCommentSubscribeInfo() map[string][]Subscriber

	PushSlash(subscriber Subscriber, info []byte)
	PushHeight(subscriber Subscriber, info []byte)
	PushTicker(subscriber Subscriber, t []*Ticker)
	PushDepthSell(subscriber Subscriber, info []byte)
	PushDepthBuy(subscriber Subscriber, info []byte)
	PushCandleStick(subscriber Subscriber, cs *CandleStick)
	PushDeal(subscriber Subscriber, info []byte)
	PushCreateOrder(subscriber Subscriber, info []byte)
	PushFillOrder(subscriber Subscriber, info []byte)
	PushCancelOrder(subscriber Subscriber, info []byte)
	PushBancorInfo(subscriber Subscriber, info []byte)
	PushBancorTrade(subscriber Subscriber, info []byte)
	PushIncome(subscriber Subscriber, info []byte)
	PushUnbonding(subscriber Subscriber, info []byte)
	PushRedelegation(subscriber Subscriber, info []byte)
	PushUnlock(subscriber Subscriber, info []byte)
	PushTx(subscriber Subscriber, info []byte)
	PushComment(subscriber Subscriber, info []byte)
}

//func DefaultDepthManager() *DepthManager
//func (dm *DepthManager) EndBlock() map[*types.PricePoint]bool

type TripleManager struct {
	sell *DepthManager
	buy  *DepthManager
	tman *TickerManager
}

type Hub struct {
	db            dbm.DB
	batch         dbm.Batch
	dbMutex       *sync.RWMutex
	tickerMutex   *sync.RWMutex
	depthMutex    *sync.RWMutex
	subMan        SubscribeManager
	managersMap   map[string]TripleManager
	csMan         *CandleStickManager
	currBlockTime time.Time
	lastBlockTime time.Time
	tickerMap     map[string]*Ticker
	sid           int64
}

func (hub *Hub) HasMarket(market string) bool {
	_, ok := hub.managersMap[market]
	return ok
}

func (hub *Hub) AddMarket(market string) {
	hub.managersMap[market] = TripleManager{
		sell: DefaultDepthManager(),
		buy:  DefaultDepthManager(),
		tman: DefaultTickerManager(market),
	}
	hub.csMan.Add(market)
}

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
	default:
		hub.Log(fmt.Sprintf("Unknown Message Type:%s", msgType))
	}
}

func (hub *Hub) Log(s string) {
	fmt.Print(s)
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

	candleSticks := hub.csMan.NewBlock(hub.currBlockTime)
	var triman TripleManager
	var targets []Subscriber
	sym := ""
	var ok bool
	for _, cs := range candleSticks {
		if sym != cs.MarketSymbol {
			triman, ok = hub.managersMap[cs.MarketSymbol]
			if !ok {
				sym = ""
				continue
			}
			info := hub.subMan.GetCandleStickSubscribeInfo()
			sym = cs.MarketSymbol
			targets, ok = info[sym]
			if !ok {
				sym = ""
				continue
			}
		}
		if len(sym) != 0 {
			if cs.TimeSpan == Minute {
				currMinute := hub.currBlockTime.Hour() * hub.currBlockTime.Minute()
				triman.tman.UpdateNewestPrice(cs.EndPrice, currMinute)
			}
			for _, target := range targets {
				hub.subMan.PushCandleStick(target, &cs)
			}
		}
	}
}

func (hub *Hub) handleNotificationSlash(bz []byte) {
	for _, ss := range hub.subMan.GetSlashSubscribeInfo() {
		hub.subMan.PushSlash(ss, bz)
	}
}

func (hub *Hub) handleNotificationTx(bz []byte) {
	var v NotificationTx
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal NotificationTx")
		return
	}
	snBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(snBytes[:], uint64(v.SerialNumber))

	key := append([]byte{DetailByte}, snBytes...)
	hub.batch.Set(key, bz)
	hub.sid++

	for _, acc := range v.Signers {
		signer := acc.String()
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
		return
	}
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
		return
	}
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
	info := hub.subMan.GetUnlockSubscribeInfo()
	targets, ok := info[v.Address.String()]
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
	if !hub.HasMarket(v.TradingPair) {
		hub.AddMarket(v.TradingPair)
	}
	key := hub.getCreateOrderKey(v.TradingPair)
	hub.batch.Set(key, bz)
	hub.sid++
	info := hub.subMan.GetOrderSubscribeInfo()
	targets, ok := info[v.Sender]
	if !ok {
		return
	}
	for _, target := range targets {
		hub.subMan.PushCreateOrder(target, bz)
	}

	managers, ok := hub.managersMap[v.TradingPair]
	if !ok {
		return
	}
	amount := sdk.NewInt(v.Quantity)
	hub.depthMutex.Lock()
	defer func() {
		hub.depthMutex.Unlock()
	}()
	if v.Side == market.SELL {
		managers.sell.DeltaChange(v.Price, amount)
	} else {
		managers.buy.DeltaChange(v.Price, amount)
	}
}
func (hub *Hub) handleFillOrderInfo(bz []byte) {
	var v FillOrderInfo
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal FillOrderInfo")
		return
	}
	key := hub.getFillOrderKey(v.TradingPair)
	hub.batch.Set(key, bz)
	hub.sid++
	info := hub.subMan.GetOrderSubscribeInfo()
	accAndSeq := strings.Split(v.OrderID, "-")
	if len(accAndSeq) != 2 {
		return
	}
	targets, ok := info[accAndSeq[0]]
	if !ok {
		return
	}
	for _, target := range targets {
		hub.subMan.PushFillOrder(target, bz)
	}

	csRec := hub.csMan.GetRecord(v.TradingPair)
	if csRec == nil {
		return
	}
	price := sdk.NewDec(v.DealMoney).QuoInt64(v.DealStock)
	csRec.Update(hub.currBlockTime, price, v.DealStock)

	managers, ok := hub.managersMap[v.TradingPair]
	if !ok {
		return
	}
	negStock := sdk.NewInt(-v.DealStock)
	hub.depthMutex.Lock()
	defer func() {
		hub.depthMutex.Unlock()
	}()
	managers.sell.DeltaChange(v.Price, negStock)
	managers.buy.DeltaChange(v.Price, negStock)
}

func (hub *Hub) handleCancelOrderInfo(bz []byte) {
	var v CancelOrderInfo
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal CancelOrderInfo")
		return
	}
	key := hub.getCancelOrderKey(v.TradingPair)
	hub.batch.Set(key, bz)
	hub.sid++
	info := hub.subMan.GetOrderSubscribeInfo()
	accAndSeq := strings.Split(v.OrderID, "-")
	if len(accAndSeq) != 2 {
		return
	}
	targets, ok := info[accAndSeq[0]]
	if !ok {
		return
	}
	for _, target := range targets {
		hub.subMan.PushCancelOrder(target, bz)
	}

	managers, ok := hub.managersMap[v.TradingPair]
	if !ok {
		return
	}
	negStock := sdk.NewInt(-v.LeftStock)
	hub.depthMutex.Lock()
	defer func() {
		hub.depthMutex.Unlock()
	}()
	if v.Side == market.SELL {
		managers.sell.DeltaChange(v.Price, negStock)
	} else {
		managers.buy.DeltaChange(v.Price, negStock)
	}
}

func (hub *Hub) handleMsgBancorTradeInfoForKafka(bz []byte) {
	var v MsgBancorTradeInfoForKafka
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal MsgBancorTradeInfoForKafka")
		return
	}
	addr := v.Sender.String()
	key := hub.getBancorTradeKey(addr)
	hub.batch.Set(key, bz)
	hub.sid++
	info := hub.subMan.GetBancorTradeSubscribeInfo()
	targets, ok := info[addr]
	if !ok {
		return
	}
	for _, target := range targets {
		hub.subMan.PushBancorTrade(target, bz)
	}
	info = hub.subMan.GetBancorInfoSubscribeInfo()
	targets, ok = info[v.Stock+"/"+v.Money]
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
	key := hub.getBancorInfoKey(v.Money + "/" + v.Stock)
	hub.batch.Set(key, bz)
	hub.sid++
	info := hub.subMan.GetBancorInfoSubscribeInfo()
	targets, ok := info[v.Stock+"/"+v.Money]
	if !ok {
		return
	}
	for _, target := range targets {
		hub.subMan.PushBancorInfo(target, bz)
	}
}

func (hub *Hub) commit() {
	tickerMap := make(map[string]*Ticker)
	currMinute := hub.currBlockTime.Hour() * hub.currBlockTime.Minute()
	for _, triman := range hub.managersMap {
		ticker := triman.tman.GetTiker(currMinute)
		if ticker != nil {
			tickerMap[ticker.Market] = ticker
		}
	}
	for _, subscriber := range hub.subMan.GetTickerSubscribeInfo() {
		marketList, ok := subscriber.Detail().([]string)
		if !ok {
			continue
		}
		tickerList := make([]*Ticker, 0, len(marketList))
		for _, market := range marketList {
			ticker, ok := tickerMap[market]
			if ok {
				tickerList = append(tickerList, ticker)
			}
		}
		hub.subMan.PushTicker(subscriber, tickerList)
	}

	hub.tickerMutex.Lock()
	for market, ticker := range tickerMap {
		hub.tickerMap[market] = ticker
	}
	hub.tickerMutex.Unlock()

	hub.dbMutex.Lock()
	hub.batch.WriteSync()
	hub.batch = hub.db.NewBatch()
	hub.dbMutex.Unlock()
}

func (hub *Hub) QueryTikers(marketList []string) []*Ticker {
	tickerList := make([]*Ticker, 0, len(marketList))
	hub.tickerMutex.RLock()
	for _, market := range marketList {
		ticker, ok := hub.tickerMap[market]
		if ok {
			tickerList = append(tickerList, ticker)
		}
	}
	hub.tickerMutex.RUnlock()
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
		count--
		if count < 0 {
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

func (hub *Hub) QueryCandleStick(market string, timespan byte, unixSec int64, count int) [][]byte {
	count = limitCount(count)
	data := make([][]byte, 0, count)
	end := getCandleStickKeyEnd(market, timespan, unixSec)
	start := getCandleStickKeyStart(market, timespan)
	hub.dbMutex.RLock()
	iter := hub.db.ReverseIterator(start, end)
	defer func() {
		iter.Close()
		hub.dbMutex.RUnlock()
	}()
	for ; iter.Valid(); iter.Next() {
		data = append(data, iter.Value())
		count--
		if count < 0 {
			break
		}
	}
	return data
}

//=========
func (hub *Hub) QueryDeal(market string, unixSec int64, count int) [][]byte {
	data, _ := hub.query(false, DealByte, []byte(market), unixSec, count)
	return data
}

func (hub *Hub) QueryOrder(account string, unixSec int64, count int) (data [][]byte, tags []byte) {
	data, tags = hub.query(false, OrderByte, []byte(account), unixSec, count)
	return
}

func (hub *Hub) QueryBancorInfo(market string, unixSec int64, count int) [][]byte {
	data, _ := hub.query(false, BancorInfoByte, []byte(market), unixSec, count)
	return data
}

func (hub *Hub) QueryBancorTrade(account string, unixSec int64, count int) [][]byte {
	data, _ := hub.query(false, BancorTradeByte, []byte(account), unixSec, count)
	return data
}

func (hub *Hub) QueryRedelegation(account string, unixSec int64, count int) [][]byte {
	data, _ := hub.query(false, RedelegationByte, []byte(account), unixSec, count)
	return data
}
func (hub *Hub) QueryUnbonding(account string, unixSec int64, count int) [][]byte {
	data, _ := hub.query(false, UnbondingByte, []byte(account), unixSec, count)
	return data
}
func (hub *Hub) QueryUnlock(account string, unixSec int64, count int) [][]byte {
	data, _ := hub.query(false, UnlockByte, []byte(account), unixSec, count)
	return data
}

func (hub *Hub) QueryIncome(account string, unixSec int64, count int) [][]byte {
	data, _ := hub.query(true, IncomeByte, []byte(account), unixSec, count)
	return data
}

func (hub *Hub) QueryTx(account string, unixSec int64, count int) [][]byte {
	data, _ := hub.query(true, TxByte, []byte(account), unixSec, count)
	return data
}

func (hub *Hub) QueryComment(token string, unixSec int64, count int) [][]byte {
	data, _ := hub.query(false, CommentByte, []byte(token), unixSec, count)
	return data
}

func (hub *Hub) query(fetchTxDetail bool, firstByte byte, bz []byte, unixSec int64, count int) (data [][]byte, tags []byte) {
	count = limitCount(count)
	data = make([][]byte, 0, MaxCount)
	tags = make([]byte, 0, MaxCount)
	start := getStartKeyFromBytes(firstByte, bz)
	end := getEndKeyFromBytes(firstByte, bz, unixSec)
	hub.dbMutex.RLock()
	iter := hub.db.ReverseIterator(start, end)
	defer func() {
		iter.Close()
		hub.dbMutex.RUnlock()
	}()
	for ; iter.Valid(); iter.Next() {
		iKey := iter.Key()
		tags = append(tags, iKey[len(iKey)-1])
		if fetchTxDetail {
			key := append([]byte{DetailByte}, iter.Value()...)
			data = append(data, hub.db.Get(key))
		} else {
			data = append(data, iter.Value())
		}
		count--
		if count < 0 {
			break
		}
	}
	return data, tags
}
