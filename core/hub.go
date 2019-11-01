package core

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	log "github.com/sirupsen/logrus"
	dbm "github.com/tendermint/tm-db"
)

// only following methods from dbm.DB are used
//type Batch interface {
//	Set(key, value []byte)
//	WriteSync()
//	Close()
//}
//
//type Iterator interface {
//	Valid() bool
//	Next()
//	Key() (key []byte)
//	Value() (value []byte)
//	Close()
//}
//
//type DB interface {
//	Get([]byte) []byte
//	Iterator(start, end []byte) Iterator
//	ReverseIterator(start, end []byte) Iterator
//	Close()
//	NewBatch() Batch
//}

const (
	MaxCount     = 1024
	DumpVersion  = byte(0)
	DumpInterval = 1000
	DumpMinTime  = 10 * time.Minute
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
	LatestHeightByte = byte(0x26)
	BancorDealByte   = byte(0x28)
	RedelegationByte = byte(0x30)
	UnbondingByte    = byte(0x32)
	UnlockByte       = byte(0x34)
	LockedByte       = byte(0x36)
	DonationByte     = byte(0x38)
	OffsetByte       = byte(0xF0)
	DumpByte         = byte(0xF1)

	CreateOrderOnlyByte = byte(0x40)
	FillOrderOnlyByte   = byte(0x42)
	CancelOrderOnlyByte = byte(0x44)
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

func BigEndianBytesToInt64(bz []byte) int64 {
	val := binary.BigEndian.Uint64(bz)
	return int64(val)
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
func (hub *Hub) getBancorDealKey(market string) []byte {
	return hub.getKeyFromBytes(BancorDealByte, []byte(market), byte(0))
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
func (hub *Hub) getEventKeyWithSidAndTime(firstByte byte, addr string, time int64, sid int64) []byte {
	res := make([]byte, 0, 1+1+len(addr)+1+16+1)
	res = append(res, firstByte)
	res = append(res, byte(len(addr)))
	res = append(res, []byte(addr)...)
	res = append(res, byte(0))
	res = append(res, int64ToBigEndianBytes(time)...) //the block's time at which the KV pair is generated
	res = append(res, int64ToBigEndianBytes(sid)...)  // the serial ID for a KV pair
	res = append(res, 0)
	return res
}

type TripleManager struct {
	sell *DepthManager
	buy  *DepthManager
	tkm  *TickerManager

	isChangedInCurrBlock bool
	mutex                sync.RWMutex
}

type msgEntry struct {
	msgType string
	bz      []byte
}

type MsgToPush struct {
	topic string
	extra interface{}
	bz    []byte
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

	csMan CandleStickManager

	// Updating logic and query logic share these variables
	managersMap map[string]*TripleManager
	tickerMap   map[string]*Ticker

	// interface to the subscribe functions
	subMan      SubscribeManager
	msgsChannel chan MsgToPush

	msgEntryList    []msgEntry
	currBlockHeight int64
	skipHeight      bool

	currBlockTime time.Time
	lastBlockTime time.Time

	blocksInterval int64

	// cache for NotificationSlash
	slashSlice []*NotificationSlash

	// dump
	partition    int32
	offset       int64
	dumpFlag     bool
	lastDumpTime time.Time

	stopped bool

	currTxHashID string
}

func NewHub(db dbm.DB, subMan SubscribeManager, interval int64, monitorInterval int64) (hub *Hub) {
	hub = &Hub{
		db:             db,
		batch:          db.NewBatch(),
		subMan:         subMan,
		managersMap:    make(map[string]*TripleManager),
		csMan:          NewCandleStickManager(nil),
		currBlockTime:  time.Unix(0, 0),
		lastBlockTime:  time.Unix(0, 0),
		tickerMap:      make(map[string]*Ticker),
		slashSlice:     make([]*NotificationSlash, 0, 10),
		partition:      0,
		offset:         0,
		dumpFlag:       false,
		lastDumpTime:   time.Now(),
		stopped:        false,
		msgEntryList:   make([]msgEntry, 0, 1000),
		blocksInterval: interval,
		msgsChannel:    make(chan MsgToPush, 10000),
	}

	go hub.pushMsgToWebsocket()

	if monitorInterval <= 0 {
		return
	}
	go func() {
		oldSid := hub.sid
		for {
			time.Sleep(time.Duration(monitorInterval * int64(time.Second)))
			if oldSid == hub.sid {
				panic("No progress for a long time")
			}
			oldSid = hub.sid
		}
	}()
	return
}

func (hub *Hub) pushMsgToWebsocket() {
	for {
		entry := <-hub.msgsChannel
		switch entry.topic {
		case BlockInfoKey:
			hub.PushHeightInfoMsg(entry.bz)
		case KlineKey:
			vals := entry.extra.([]string)
			hub.PushCandleMsg(vals[0], entry.bz, vals[1])
		case LockedKey:
			hub.PushLockedCoinsMsg(entry.extra.(string), entry.bz)
		case IncomeKey:
			hub.PushInComeMsg(entry.extra.(string), entry.bz)
		case TxKey:
			hub.PushTxMsg(entry.extra.(string), entry.bz)
		case RedelegationKey:
			hub.PushRedelegationMsg(entry.extra.(TimeAndSidWithAddr))
		case UnbondingKey:
			hub.PushUnbondingMsg(entry.extra.(TimeAndSidWithAddr))
		case UnlockKey:
			hub.PushUnlockMsg(entry.extra.(string), entry.bz)
		case CommentKey:
			hub.PushCommentMsg(entry.extra.(string), entry.bz)
		case CreateOrderKey:
			hub.PushCreateOrderInfoMsg(entry.extra.(string), entry.bz)
		case FillOrderKey:
			hub.PushFillOrderInfoMsg(entry.extra.(string), entry.bz)
		case DealKey:
			hub.PushDealInfoMsg(entry.extra.(string), entry.bz)
		case CancelOrderKey:
			hub.PushCancelOrderMsg(entry.extra.(string), entry.bz)
		case BancorTradeKey:
			hub.PushBancorTradeInfoMsg(entry.extra.(string), entry.bz)
		case BancorDealKey:
			hub.PushBancorDealMsg(entry.extra.(string), entry.bz)
		case BancorKey:
			hub.PushBancorMsg(entry.extra.(string), entry.bz)
		case SlashKey:
			hub.PushSlashMsg(entry.bz)
		case TickerKey:
			hub.PushTickerMsg(entry.bz)
		case DepthFull:
			hub.PushDepthFullMsg(entry.extra.(string))
		case DepthKey:
			hub.PushDepthMsg(entry.bz, entry.extra.(map[string][]byte))
		case OptionKey:
			hub.subMan.SetSkipOption(entry.extra.(bool))
		}
	}
}

func (hub *Hub) HasMarket(market string) bool {
	_, ok := hub.managersMap[market]
	return ok
}

func (hub *Hub) AddMarket(market string) {
	if strings.HasPrefix(market, "B:") {
		hub.managersMap[market] = &TripleManager{
			tkm: DefaultTickerManager(market),
		}
	} else {
		hub.managersMap[market] = &TripleManager{
			sell: DefaultDepthManager("sell"),
			buy:  DefaultDepthManager("buy"),
			tkm:  DefaultTickerManager(market),
		}
	}
	hub.csMan.AddMarket(market)
}

func (hub *Hub) Log(s string) {
	log.Error(s)
}

//============================================================
var _ Consumer = &Hub{}

func (hub *Hub) ConsumeMessage(msgType string, bz []byte) {

	// fmt.Printf("msgType : %s, values : %s\n", msgType, string(bz))
	if msgType == "height_info" {
		hub.prehandleNewHeightInfo(bz)
	}
	if hub.skipHeight {
		return
	}

	hub.msgEntryList = append(hub.msgEntryList, msgEntry{msgType: msgType, bz: bz})
	// if msgType == "fill_order_info" {
	// 	fmt.Printf("trade-server have received fillOrderInfo, time : %d\n", time.Now().Second())
	// }
	// if msgType == "height_info" {
	// 	fmt.Printf("trade-server received newBlockInfo, time : %d\n", time.Now().Second())
	// }

	if msgType != "commit" {
		return
	}

	for _, entry := range hub.msgEntryList {
		switch entry.msgType {
		case "height_info":
			hub.handleNewHeightInfo(entry.bz)
		case "slash":
			hub.handleNotificationSlash(entry.bz)
		case "notify_tx":
			hub.handleNotificationTx(entry.bz)
		case "begin_redelegation":
			hub.handleNotificationBeginRedelegation(entry.bz)
		case "begin_unbonding":
			hub.handleNotificationBeginUnbonding(entry.bz)
		case "complete_redelegation":
			hub.handleNotificationCompleteRedelegation(entry.bz)
		case "complete_unbonding":
			hub.handleNotificationCompleteUnbonding(entry.bz)
		case "notify_unlock":
			hub.handleNotificationUnlock(entry.bz)
		case "token_comment":
			hub.handleTokenComment(entry.bz)
		case "create_order_info":
			hub.handleCreateOrderInfo(entry.bz)
		case "fill_order_info":
			hub.handleFillOrderInfo(entry.bz)
		case "del_order_info":
			hub.handleCancelOrderInfo(entry.bz)
		case "bancor_trade":
			hub.handleMsgBancorTradeInfoForKafka(entry.bz)
		case "bancor_info":
			hub.handleMsgBancorInfoForKafka(entry.bz)
		case "commit":
			hub.commit()
		case "send_lock_coins":
			hub.handleLockedCoinsMsg(entry.bz)
		default:
			hub.Log(fmt.Sprintf("Unknown Message Type:%s", entry.msgType))
		}
	}

	hub.msgEntryList = hub.msgEntryList[:0]
}

func (hub *Hub) prehandleNewHeightInfo(bz []byte) {
	var v NewHeightInfo
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal NewHeightInfo")
		return
	}

	if hub.currBlockHeight >= v.Height {
		hub.skipHeight = true
		hub.Log(fmt.Sprintf("Skipping Height %d<%d\n", hub.currBlockHeight, v.Height))
	} else if hub.currBlockHeight+1 == v.Height {
		hub.currBlockHeight = v.Height
		hub.skipHeight = false
	} else {
		hub.skipHeight = true
		hub.Log(fmt.Sprintf("Invalid Height! %d+1!=%d\n", hub.currBlockHeight, v.Height))
		panic("here")
	}

	hub.msgEntryList = hub.msgEntryList[:0]
}

func (hub *Hub) handleNewHeightInfo(bz []byte) {
	var v NewHeightInfo
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal NewHeightInfo")
		return
	}

	latestHeight := hub.QueryLatestHeight()
	if latestHeight >= v.Height {
		hub.msgsChannel <- MsgToPush{topic: OptionKey, extra: true}
		//hub.subMan.SetSkipOption(true)
		hub.Log(fmt.Sprintf("Skipping Height websocket %d<%d\n", latestHeight, v.Height))
	} else if latestHeight+1 == v.Height {
		hub.msgsChannel <- MsgToPush{topic: OptionKey, extra: false}
		//hub.subMan.SetSkipOption(false)
	} else if latestHeight < 0 {
		hub.msgsChannel <- MsgToPush{topic: OptionKey, extra: false}
		//hub.subMan.SetSkipOption(false)
	} else {
		hub.msgsChannel <- MsgToPush{topic: OptionKey, extra: true}
		//hub.subMan.SetSkipOption(true)
		hub.Log(fmt.Sprintf("Invalid Height! websocket %d+1!=%d\n", latestHeight, v.Height))
	}

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b[:], uint64(v.TimeStamp.Unix()))
	heightBytes := int64ToBigEndianBytes(v.Height)
	key := append([]byte{BlockHeightByte}, heightBytes...)
	hub.batch.Set(key, b)
	hub.batch.Set([]byte{LatestHeightByte}, heightBytes)
	hub.msgsChannel <- MsgToPush{topic: BlockInfoKey, bz: bz}
	hub.lastBlockTime = hub.currBlockTime
	hub.currBlockTime = v.TimeStamp
	hub.beginForCandleSticks()
}

func (hub *Hub) PushHeightInfoMsg(bz []byte) {
	infos := hub.subMan.GetHeightSubscribeInfo()
	for _, ss := range infos {
		hub.subMan.PushHeight(ss, bz)
	}
}

func (hub *Hub) beginForCandleSticks() {
	candleSticks := hub.csMan.NewBlock(hub.currBlockTime)
	var triman *TripleManager
	sym := ""
	var ok bool
	currMinute := hub.currBlockTime.UTC().Hour() * hub.currBlockTime.UTC().Minute()
	for _, cs := range candleSticks {
		if sym != cs.Market {
			triman, ok = hub.managersMap[cs.Market]
			if !ok {
				sym = ""
				continue
			}
			sym = cs.Market
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
		extra := []string{sym, cs.TimeSpan}
		hub.msgsChannel <- MsgToPush{topic: KlineKey, bz: bz, extra: extra}
		// Save candle sticks to KVStore
		key := hub.getCandleStickKey(cs.Market, GetSpanFromSpanStr(cs.TimeSpan))
		if len(bz) == 0 {
			continue
		}
		hub.batch.Set(key, bz)
		hub.sid++
	}
}

// TODO. Add test
func (hub *Hub) PushCandleMsg(market string, bz []byte, timeSpan string) {
	var targets []Subscriber
	info := hub.subMan.GetCandleStickSubscribeInfo()
	if info != nil {
		targets = info[market]
	}
	for _, target := range targets {
		ts, ok := target.Detail().(string)
		if !ok || ts != timeSpan {
			continue
		}
		hub.subMan.PushCandleStick(target, bz)
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

func appendHashID(bz []byte, hashID string) []byte {
	if len(hashID) == 0 {
		return bz
	}
	return append(bz[0:len(bz)-1], []byte(fmt.Sprintf(`,"tx_hash":"%s"}`, hashID))...)
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
	//v.TxHash = hub.currTxHashID
	bz = appendHashID(bz, hub.currTxHashID)
	key := hub.getLockedKey(v.ToAddress)
	hub.batch.Set(key, bz)
	hub.sid++
	hub.msgsChannel <- MsgToPush{topic: LockedKey, bz: bz, extra: v.ToAddress}
}

func (hub *Hub) PushLockedCoinsMsg(addr string, bz []byte) {
	infos := hub.subMan.GetLockedSubscribeInfo()
	if conns, ok := infos[addr]; ok {
		for _, c := range conns {
			hub.subMan.PushLockedSendMsg(c, bz)
		}
	}
}

func (hub *Hub) analyzeMessages(MsgTypes []string, TxJSON string) {
	if len(TxJSON) == 0 {
		return
	}
	var tx map[string]interface{}
	err := json.Unmarshal([]byte(TxJSON), &tx)
	if err != nil {
		hub.Log(fmt.Sprintf("Error in Unmarshal NotificationTx: %s (%v)", TxJSON, err))
		return
	}
	msgListRaw, ok := tx["msg"]
	if !ok {
		hub.Log(fmt.Sprintf("No msg found: %s", TxJSON))
		return
	}
	msgList, ok := msgListRaw.([]interface{})
	if !ok {
		hub.Log(fmt.Sprintf("msg is not array: %s", TxJSON))
		return
	}
	if len(msgList) != len(MsgTypes) {
		hub.Log(fmt.Sprintf("Length mismatch in Unmarshal NotificationTx: %s %s", TxJSON, MsgTypes))
		return
	}
	var donation Donation
	for i, msgType := range MsgTypes {
		msg, ok := msgList[i].(map[string]interface{})
		if !ok {
			continue
		}
		if msgType == "MsgDonateToCommunityPool" {
			donation.Sender, _ = msg["from_addr"].(string)
			amount, _ := msg["amount"].([]interface{})
			amount0, _ := amount[0].(map[string]interface{})
			amountCET, _ := amount0["amount"].(string)
			donation.Amount = amountCET
		} else if msgType == "MsgCommentToken" {
			donation.Sender, _ = msg["sender"].(string)
			amount := int64(msg["donation"].(float64))
			donation.Amount = fmt.Sprintf("%d", amount)
		} else {
			continue
		}

		bz, err := json.Marshal(&donation)
		if err != nil {
			hub.Log(fmt.Sprintf("Error in Marshal Donation: %v", donation))
			continue
		}
		key := hub.getKeyFromBytes(DonationByte, []byte{}, 0)
		hub.batch.Set(key, bz)
		hub.sid++
	}
}

func (hub *Hub) handleNotificationTx(bz []byte) {
	var v NotificationTx
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log(fmt.Sprintf("Error in Unmarshal NotificationTx: %s", string(bz)))
		return
	}

	// base64 -> hex
	decodeBytes, err := base64.StdEncoding.DecodeString(v.Hash)
	if err == nil {
		v.Hash = strings.ToUpper(hex.EncodeToString(decodeBytes))
		if bz, err = json.Marshal(v); err != nil {
			hub.Log(fmt.Sprintf("Error in Marshal NotificationTx: %v (%v)", v, err))
		}
	} else {
		hub.Log(fmt.Sprintf("Error in decode base64 tx hash: %v (%v)", v.Hash, err))
	}

	hub.currTxHashID = v.Hash

	// Use the transaction's hashid as key, save its detail
	key := append([]byte{DetailByte}, []byte(v.Hash)...)
	hub.batch.Set(key, bz)
	hub.sid++

	tokenNames := make(map[string]bool)
	for _, transRec := range v.Transfers {
		tokenName := getTokenNameFromAmount(transRec.Amount)
		tokenNames[tokenName] = true

		recipient := transRec.Recipient
		k := hub.getIncomeKey(recipient)
		hub.batch.Set(k, []byte("|"+tokenName+"|"+v.Hash))
		hub.sid++
		hub.msgsChannel <- MsgToPush{topic: IncomeKey, bz: bz, extra: recipient}
	}

	tokensAndHash := make([]byte, 1, 60)
	tokensAndHash[0] = byte('|')
	for tokenName := range tokenNames {
		tokensAndHash = append(tokensAndHash, []byte(tokenName+"|")...)
	}
	tokensAndHash = append(tokensAndHash, []byte(v.Hash)...)

	for _, acc := range v.Signers {
		signer := acc
		k := hub.getTxKey(signer)
		hub.batch.Set(k, tokensAndHash)
		hub.sid++
		hub.msgsChannel <- MsgToPush{topic: TxKey, bz: bz, extra: signer}
	}

	hub.analyzeMessages(v.MsgTypes, v.TxJSON)
}

func (hub *Hub) PushTxMsg(addr string, bz []byte) {
	info := hub.subMan.GetTxSubscribeInfo()
	targets, ok := info[addr]
	if ok {
		for _, target := range targets {
			hub.subMan.PushTx(target, bz)
		}
	}
}

func (hub *Hub) PushInComeMsg(receiver string, bz []byte) {
	info := hub.subMan.GetIncomeSubscribeInfo()
	targets, ok := info[receiver]
	if ok {
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
	//v.TxHash = hub.currTxHashID
	bz = appendHashID(bz, hub.currTxHashID)
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
	//v.TxHash = hub.currTxHashID
	bz = appendHashID(bz, hub.currTxHashID)
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
	hub.msgsChannel <- MsgToPush{topic: RedelegationKey, extra: TimeAndSidWithAddr{addr: v.Delegator, sid: hub.sid,
		currTime: hub.currBlockTime.Unix(), lastTime: hub.lastBlockTime.Unix()}}
}

type TimeAndSidWithAddr struct {
	currTime int64
	lastTime int64
	sid      int64
	addr     string
}

func (hub *Hub) PushRedelegationMsg(param TimeAndSidWithAddr) {
	info := hub.subMan.GetRedelegationSubscribeInfo()
	targets, ok := info[param.addr]
	if ok {
		// query the redelegations whose completion time is between current block and last block
		end := hub.getEventKeyWithSidAndTime(RedelegationByte, param.addr, param.currTime, param.sid)
		start := hub.getEventKeyWithSidAndTime(RedelegationByte, param.addr, param.lastTime-1, param.sid)
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
}

func (hub *Hub) handleNotificationCompleteUnbonding(bz []byte) {
	var v NotificationCompleteUnbonding
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal NotificationCompleteUnbonding")
		return
	}
	hub.msgsChannel <- MsgToPush{topic: UnbondingKey, bz: bz, extra: TimeAndSidWithAddr{addr: v.Delegator, sid: hub.sid,
		currTime: hub.currBlockTime.Unix(), lastTime: hub.lastBlockTime.Unix()}}
}

func (hub *Hub) PushUnbondingMsg(param TimeAndSidWithAddr) {
	info := hub.subMan.GetUnbondingSubscribeInfo()
	targets, ok := info[param.addr]
	if ok {
		// query the unbondings whose completion time is between current block and last block
		end := hub.getEventKeyWithSidAndTime(UnbondingByte, param.addr, param.currTime, param.sid)
		start := hub.getEventKeyWithSidAndTime(UnbondingByte, param.addr, param.lastTime-1, param.sid)
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
	hub.msgsChannel <- MsgToPush{topic: UnlockKey, bz: bz, extra: addr}
}

func (hub *Hub) PushUnlockMsg(addr string, bz []byte) {
	info := hub.subMan.GetUnlockSubscribeInfo()
	targets, ok := info[addr]
	if ok {
		for _, target := range targets {
			hub.subMan.PushUnlock(target, bz)
		}
	}
}

func (hub *Hub) handleTokenComment(bz []byte) {
	var v TokenComment
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal TokenComment")
		return
	}
	//v.TxHash = hub.currTxHashID
	bz = appendHashID(bz, hub.currTxHashID)
	key := hub.getCommentKey(v.Token)
	hub.batch.Set(key, bz)
	hub.sid++
	hub.msgsChannel <- MsgToPush{topic: CommentKey, bz: bz, extra: v.Token}
}

func (hub *Hub) PushCommentMsg(token string, bz []byte) {
	info := hub.subMan.GetCommentSubscribeInfo()
	targets, ok := info[token]
	if ok {
		for _, target := range targets {
			hub.subMan.PushComment(target, bz)
		}
	}
}

func (hub *Hub) handleCreateOrderInfo(bz []byte) {
	var v CreateOrderInfo
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal CreateOrderInfo")
		return
	}
	//v.TxHash = hub.currTxHashID
	bz = appendHashID(bz, hub.currTxHashID)
	// Add a new market which is seen for the first time
	if !hub.HasMarket(v.TradingPair) {
		hub.AddMarket(v.TradingPair)
	}
	//Save to KVStore
	key := hub.getCreateOrderKey(v.Sender)
	hub.batch.Set(key, bz)
	hub.sid++
	//Push to subscribers
	hub.msgsChannel <- MsgToPush{topic: CreateOrderKey, bz: bz, extra: v.Sender}
	//Update depth info
	triman, ok := hub.managersMap[v.TradingPair]
	if !ok {
		return
	}
	amount := sdk.NewInt(v.Quantity)

	if !triman.isChangedInCurrBlock {
		triman.mutex.Lock()
		triman.isChangedInCurrBlock = true
	}
	if v.Side == SELL {
		triman.sell.DeltaChange(v.Price, amount)
	} else {
		triman.buy.DeltaChange(v.Price, amount)
	}
}

func (hub *Hub) PushCreateOrderInfoMsg(addr string, bz []byte) {
	//Push to subscribers
	info := hub.subMan.GetOrderSubscribeInfo()
	targets, ok := info[addr]
	if ok {
		for _, target := range targets {
			hub.subMan.PushCreateOrder(target, bz)
		}
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
	if v.Side == SELL {
		key = hub.getDealKey(v.TradingPair)
		hub.batch.Set(key, bz)
		hub.sid++
	}
	//Push to subscribers
	hub.msgsChannel <- MsgToPush{topic: FillOrderKey, bz: bz, extra: accAndSeq[0]}
	if v.Side == SELL {
		hub.msgsChannel <- MsgToPush{topic: DealKey, bz: bz, extra: v.TradingPair}
	}
	//Update candle sticks
	if v.Side == SELL {
		csRec := hub.csMan.GetRecord(v.TradingPair)
		if csRec != nil {
			csRec.Update(hub.currBlockTime, v.FillPrice, v.CurrStock)
		}
	}
	//Update depth info
	triman, ok := hub.managersMap[v.TradingPair]
	if !ok {
		return
	}
	if !triman.isChangedInCurrBlock {
		triman.mutex.Lock()
		triman.isChangedInCurrBlock = true
	}
	negStock := sdk.NewInt(-v.CurrStock)
	if v.Side == SELL {
		triman.sell.DeltaChange(v.Price, negStock)
	} else {
		triman.buy.DeltaChange(v.Price, negStock)
	}
}

func (hub *Hub) PushFillOrderInfoMsg(addr string, bz []byte) {
	//Push to subscribers
	info := hub.subMan.GetOrderSubscribeInfo()
	targets, ok := info[addr]
	if ok {
		for _, target := range targets {
			hub.subMan.PushFillOrder(target, bz)
		}
	}
}

func (hub *Hub) PushDealInfoMsg(tradingPair string, bz []byte) {
	info := hub.subMan.GetDealSubscribeInfo()
	targets, ok := info[tradingPair]
	if ok {
		for _, target := range targets {
			hub.subMan.PushDeal(target, bz)
		}
	}
}

func (hub *Hub) handleCancelOrderInfo(bz []byte) {
	var v CancelOrderInfo
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal CancelOrderInfo")
		return
	}
	if v.DelReason == "Manually cancel the order" {
		//v.TxHash = hub.currTxHashID
		bz = appendHashID(bz, hub.currTxHashID)
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
	//Update depth info
	triman, ok := hub.managersMap[v.TradingPair]
	if !ok {
		return
	}
	if !triman.isChangedInCurrBlock {
		triman.mutex.Lock()
		triman.isChangedInCurrBlock = true
	}
	negStock := sdk.NewInt(-v.LeftStock)
	if v.Side == SELL {
		triman.sell.DeltaChange(v.Price, negStock)
	} else {
		triman.buy.DeltaChange(v.Price, negStock)
	}

	hub.msgsChannel <- MsgToPush{topic: CancelOrderKey, bz: bz, extra: accAndSeq[0]}
}

func (hub *Hub) PushCancelOrderMsg(addr string, bz []byte) {
	info := hub.subMan.GetOrderSubscribeInfo()
	targets, ok := info[addr]
	if ok {
		for _, target := range targets {
			hub.subMan.PushCancelOrder(target, bz)
		}
	}
}

func (hub *Hub) handleMsgBancorTradeInfoForKafka(bz []byte) {
	var v MsgBancorTradeInfoForKafka
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal MsgBancorTradeInfoForKafka")
		return
	}
	//v.TxHash = hub.currTxHashID
	bz = appendHashID(bz, hub.currTxHashID)
	//Create market if not exist
	marketName := "B:" + v.Stock + "/" + v.Money
	if !hub.HasMarket(marketName) {
		hub.AddMarket(marketName)
	}
	//Save to KVStore
	addr := v.Sender
	key := hub.getBancorTradeKey(addr)
	hub.batch.Set(key, bz)
	hub.sid++
	key = hub.getBancorDealKey(marketName)
	hub.batch.Set(key, bz)
	hub.sid++
	//Update candle sticks
	csRec := hub.csMan.GetRecord(marketName)
	if csRec != nil {
		csRec.Update(hub.currBlockTime, v.TxPrice, v.Amount)
	}
	hub.msgsChannel <- MsgToPush{topic: BancorTradeKey, bz: bz, extra: addr}
	hub.msgsChannel <- MsgToPush{topic: BancorDealKey, bz: bz, extra: v.Stock + "/" + v.Money}
}

func (hub *Hub) PushBancorTradeInfoMsg(addr string, bz []byte) {
	info := hub.subMan.GetBancorTradeSubscribeInfo()
	targets, ok := info[addr]
	if ok {
		for _, target := range targets {
			hub.subMan.PushBancorTrade(target, bz)
		}
	}
}

func (hub *Hub) PushBancorDealMsg(tradingPair string, bz []byte) {
	info := hub.subMan.GetBancorDealSubscribeInfo()
	targets, ok := info[tradingPair]
	if ok {
		for _, target := range targets {
			hub.subMan.PushBancorDeal(target, bz)
		}
	}
}

func (hub *Hub) handleMsgBancorInfoForKafka(bz []byte) {
	var v MsgBancorInfoForKafka
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal MsgBancorInfoForKafka")
		return
	}
	//Create market if not exist
	marketName := "B:" + v.Stock + "/" + v.Money
	if !hub.HasMarket(marketName) {
		hub.AddMarket(marketName)
	}
	//Save to KVStore
	key := hub.getBancorInfoKey(v.Stock + "/" + v.Money)
	hub.batch.Set(key, bz)
	hub.sid++
	//Push to subscribers
	hub.msgsChannel <- MsgToPush{topic: BancorKey, bz: bz, extra: v.Stock + "/" + v.Money}
}

func (hub *Hub) PushBancorMsg(tradingPair string, bz []byte) {
	info := hub.subMan.GetBancorInfoSubscribeInfo()
	targets, ok := info[tradingPair]
	if ok {
		for _, target := range targets {
			hub.subMan.PushBancorInfo(target, bz)
		}
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
		hub.msgsChannel <- MsgToPush{topic: SlashKey, bz: bz}
		key := hub.getKeyFromBytes(SlashByte, []byte{}, 0)
		hub.batch.Set(key, bz)
		hub.sid++
	}
	hub.slashSlice = hub.slashSlice[:0]
}

func (hub *Hub) PushSlashMsg(bz []byte) {
	infos := hub.subMan.GetSlashSubscribeInfo()
	for _, ss := range infos {
		hub.subMan.PushSlash(ss, bz)
	}
}

func (hub *Hub) commitForTicker() {
	tkMap := make(map[string]*Ticker)
	isNewMinute := hub.currBlockTime.UTC().Minute() != hub.lastBlockTime.UTC().Minute() || hub.currBlockTime.Unix()-hub.lastBlockTime.Unix() > 60
	if !isNewMinute {
		return
	}
	currMinute := hub.currBlockTime.UTC().Hour() * hub.currBlockTime.UTC().Minute()
	for _, triman := range hub.managersMap {
		if ticker := triman.tkm.GetTicker(currMinute); ticker != nil {
			tkMap[ticker.Market] = ticker
		}
	}
	bz, err := json.Marshal(tkMap)
	if err != nil {
		return
	}
	hub.msgsChannel <- MsgToPush{topic: TickerKey, bz: bz}
	hub.tickerMapMutex.Lock()
	for market, ticker := range tkMap {
		hub.tickerMap[market] = ticker
	}
	hub.tickerMapMutex.Unlock()
}

// TODO, Add test for map marshal and unmarshal
func (hub *Hub) PushTickerMsg(bz []byte) {
	tkMap := make(map[string]*Ticker)
	err := json.Unmarshal(bz, &tkMap)
	if err != nil {
		hub.Log(err.Error())
		return
	}

	infos := hub.subMan.GetTickerSubscribeInfo()
	for _, subscriber := range infos {
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
}

func (hub *Hub) pushDepthFull() {
	if hub.currBlockHeight%hub.blocksInterval != 0 {
		return
	}
	for market := range hub.managersMap {
		if strings.HasPrefix(market, "B:") {
			continue
		}
		hub.msgsChannel <- MsgToPush{topic: DepthFull, extra: market}
	}
}

// TODO, will check topic
func (hub *Hub) PushDepthFullMsg(market string) {
	info := hub.subMan.GetDepthSubscribeInfo()
	targets, ok := info[market]
	if ok {
		for _, target := range targets {
			level := target.Detail().(string)
			if err := queryDepthAndPush(hub, target, market, level, 0); err != nil {
				hub.Log(err.Error())
			}
		}
	}
}

func (hub *Hub) commitForDepth() {
	for market, triman := range hub.managersMap {
		if strings.HasPrefix(market, "B:") {
			continue
		}

		if triman.isChangedInCurrBlock {
			triman.isChangedInCurrBlock = false
			triman.mutex.Unlock()
		} else {
			continue
		}

		depthDeltaSell, mergeDeltaSell := triman.sell.EndBlock()
		depthDeltaBuy, mergeDeltaBuy := triman.buy.EndBlock()
		if len(depthDeltaSell) == 0 && len(depthDeltaBuy) == 0 {
			continue
		}

		lowestPP := triman.sell.GetLowest(1)
		highestPP := triman.buy.GetLowest(1)
		if len(lowestPP) != 0 && len(highestPP) != 0 {
			if lowestPP[0].Price.LTE(highestPP[0].Price) {
				s := fmt.Sprintf("Error! %d sel_low:%s buy_high:%s\n",
					hub.currBlockHeight, lowestPP[0].Price, highestPP[0].Price)
				hub.Log(s)
			}
		}

		data := encodeDepthLevel(market, depthDeltaBuy, depthDeltaSell)
		levelsData := encodeDepthLevels(market, mergeDeltaBuy, mergeDeltaSell)
		if levelsData == nil {
			levelsData = make(map[string][]byte)
		}
		levelsData["all"] = data
		hub.msgsChannel <- MsgToPush{topic: DepthKey, bz: []byte(market), extra: levelsData}
	}
}

// TODO, will replace string with interface
func (hub *Hub) PushDepthMsg(data []byte, levelsData map[string][]byte) {
	market := string(data)
	info := hub.subMan.GetDepthSubscribeInfo()
	targets, ok := info[market]
	if ok {
		for _, target := range targets {
			level := target.Detail().(string)
			if len(levelsData[level]) != 0 {
				if level == "all" {
					hub.subMan.PushDepthWithChange(target, levelsData[level])
				} else {
					hub.subMan.PushDepthWithDelta(target, levelsData[level])
				}
			}
		}
	}
}

func (hub *Hub) commit() {
	if hub.stopped {
		return
	}
	hub.commitForSlash()
	hub.commitForTicker()
	hub.commitForDepth()
	hub.pushDepthFull()
	if hub.dumpFlag {
		hub.commitForDump()
		hub.dumpFlag = false
	}
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

func (hub *Hub) QueryLatestHeight() int64 {
	bz := hub.db.Get([]byte{LatestHeightByte})
	if len(bz) == 0 {
		return 0
	}
	return BigEndianBytesToInt64(bz)
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
	tripleMan.mutex.RLock()
	sell = tripleMan.sell.GetLowest(count)
	buy = tripleMan.buy.GetHighest(count)
	tripleMan.mutex.RUnlock()
	return
}

func (hub *Hub) AddLevel(market, level string) error {
	if !hub.HasMarket(market) {
		return errors.New("No such market")
	}
	tripleMan := hub.managersMap[market]
	err := tripleMan.sell.AddLevel(level)
	if err != nil {
		return err
	}
	return tripleMan.buy.AddLevel(level)
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

// --------------

func (hub *Hub) QueryOrderAboutToken(tag, token, account string, time int64, sid int64, count int) (data []json.RawMessage, tags []byte, timesid []int64) {
	firstByte := OrderByte
	if tag == CreateOrderStr {
		firstByte = CreateOrderOnlyByte
	} else if tag == FillOrderStr {
		firstByte = FillOrderOnlyByte
	} else if tag == CancelOrderStr {
		firstByte = CancelOrderOnlyByte
	}
	if token == "" {
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

func (hub *Hub) QueryLockedAboutToken(token, account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	if token == "" {
		data, _, timesid = hub.query(false, LockedByte, []byte(account), time, sid, count, nil)
		return
	}
	data, _, timesid = hub.query(false, LockedByte, []byte(account), time, sid, count, func(tag byte, entry []byte) bool {
		s := fmt.Sprintf("\"denom\":\"%s\",\"amount\":", token)
		return strings.Index(string(entry), s) > 0
	})
	return

}

func (hub *Hub) QueryBancorTradeAboutToken(token, account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	if token == "" {
		data, _, timesid = hub.query(false, BancorTradeByte, []byte(account), time, sid, count, nil)
		return
	}
	data, _, timesid = hub.query(false, BancorTradeByte, []byte(account), time, sid, count, func(tag byte, entry []byte) bool {
		s1 := fmt.Sprintf("\"money\":\"%s\"", token)
		s2 := fmt.Sprintf("\"stock\":\"%s\"", token)
		return strings.Index(string(entry), s1) > 0 || strings.Index(string(entry), s2) > 0
	})
	return

}

func (hub *Hub) QueryUnlockAboutToken(token, account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	if token == "" {
		data, _, timesid = hub.query(false, UnlockByte, []byte(account), time, sid, count, nil)
		return
	}
	data, _, timesid = hub.query(false, UnlockByte, []byte(account), time, sid, count, func(tag byte, entry []byte) bool {
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

func (hub *Hub) QueryIncomeAboutToken(token, account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	if token == "" {
		data, _, timesid = hub.query(true, IncomeByte, []byte(account), time, sid, count, nil)
		return
	}
	data, _, timesid = hub.query(true, IncomeByte, []byte(account), time, sid, count, func(tag byte, entry []byte) bool {
		return strings.Contains(string(entry), "|"+token+"|")
	})
	return

}

func (hub *Hub) QueryTxAboutToken(token, account string, time int64, sid int64, count int) (data []json.RawMessage, timesid []int64) {
	if token == "" {
		data, _, timesid = hub.query(true, TxByte, []byte(account), time, sid, count, nil)
		return
	}
	data, _, timesid = hub.query(true, TxByte, []byte(account), time, sid, count, func(tag byte, entry []byte) bool {
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
	} else {
		firstByteIn = 0
	}
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
		if filter != nil && !filter(tag, entry) {
			continue
		}
		if fetchTxDetail {
			hexTxHashID := getTxHashID(iter.Value())
			key := append([]byte{DetailByte}, hexTxHashID...)
			entry = json.RawMessage(hub.db.Get(key))
		}
		data = append(data, entry)
		tags = append(tags, tag)
		timesid = append(timesid, []int64{int64(time), int64(sid)}...)
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

func (hub *Hub) QueryTxByHashID(hexHashID string) json.RawMessage {
	hub.dbMutex.RLock()
	defer hub.dbMutex.RUnlock()
	key := append([]byte{DetailByte}, []byte(hexHashID)...)
	return json.RawMessage(hub.db.Get(key))
}

//===================================
// for serialization and deserialization of Hub

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
			sell: DefaultDepthManager("sell"),
			buy:  DefaultDepthManager("buy"),
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
	key := getDumpKey()
	hub.dbMutex.RLock()
	defer hub.dbMutex.RUnlock()
	bz := hub.db.Get(key)
	return bz
}

func (hub *Hub) commitForDump() {
	// save dump data
	hub4j := &HubForJSON{}
	hub.Dump(hub4j)
	dumpKey := getDumpKey()
	dumpBuf, err := json.Marshal(hub4j)
	if err != nil {
		log.WithError(err).Error("hub json marshal fail")
		return
	}
	hub.batch.Set(dumpKey, dumpBuf)

	// save offset
	offsetKey := getOffsetKey(hub.partition)
	offsetBuf := int64ToBigEndianBytes(hub.offset)
	hub.batch.Set(offsetKey, offsetBuf)

	hub.lastDumpTime = time.Now()
	log.Infof("dump date at offset: %v", hub.offset)
}

func getDumpKey() []byte {
	key := make([]byte, 2)
	key[0] = DumpByte
	key[1] = DumpVersion
	return key
}

//===================================
func (hub *Hub) UpdateOffset(partition int32, offset int64) {
	hub.partition = partition
	hub.offset = offset

	now := time.Now()
	// dump data every <interval> offset
	if offset%DumpInterval == 0 && now.Sub(hub.lastDumpTime) > DumpMinTime {
		hub.dumpFlag = true
	}
}

func (hub *Hub) LoadOffset(partition int32) int64 {
	key := getOffsetKey(partition)
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

func getOffsetKey(partition int32) []byte {
	key := make([]byte, 5)
	key[0] = OffsetByte
	binary.BigEndian.PutUint32(key[1:], uint32(partition))
	return key
}

func (hub *Hub) Close() {
	// try to dump the latest block
	hub.dumpFlag = true
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second)
		if !hub.dumpFlag {
			break
		}
	}
	// close db
	hub.dbMutex.Lock()
	hub.db.Close()
	hub.stopped = true
	hub.dbMutex.Unlock()
}
