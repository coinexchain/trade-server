package core

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	log "github.com/sirupsen/logrus"
	dbm "github.com/tendermint/tm-db"
)

//==========

// One entry of message from kafka
type msgEntry struct {
	msgType string
	bz      []byte
}

// A message to be sent to websocket subscriber through hub.msgsChannel
type MsgToPush struct {
	topic string
	extra interface{}
	bz    []byte
}

type TimeAndSidWithAddr struct {
	currTime int64
	lastTime int64
	sid      int64
	addr     string
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
	tickerMap   map[string]*Ticker // it caches the tickers from managersMap[*].tkm

	// interface to the subscribe functions
	subMan      SubscribeManager
	msgsChannel chan MsgToPush

	// buffers the messages from kafka to execute them in batch
	msgEntryList []msgEntry
	// skip the repeating messages after restart
	skipHeight      bool
	currBlockHeight int64
	chainID         string

	currBlockTime time.Time
	lastBlockTime time.Time

	// every 'blocksInterval' blocks, we push full depth data and change rocksdb's prune info
	blocksInterval int64
	// in rocksdb, we only keep the records which are in the most recent 'keepRecent' seconds
	keepRecent int64

	// cache for NotificationSlash
	slashSlice []*NotificationSlash

	// dump hub's in-memory information to file, which can not be stored in rocksdb
	partition    int32
	offset       int64
	lastOffset   int64
	dumpFlag     int32
	dumpFlagLock sync.Mutex
	lastDumpTime time.Time

	stopped bool

	// this member is set during handleNotificationTx and is used to fill
	currTxHashID string

	// Count how many times the mutexes are locked
	dbLockCount        int64
	tickerMapLockCount int64
	trimanLockCount    int64

	// stop operation for new chain
	oldChainID    string
	upgradeHeight int64
}

func NewHub(db dbm.DB, subMan SubscribeManager, interval int64, monitorInterval int64,
	keepRecent int64, initChainHeight int64, oldChainID string, upgradeHeight int64) (hub *Hub) {
	hub = &Hub{
		db:              db,
		batch:           db.NewBatch(),
		subMan:          subMan,
		managersMap:     make(map[string]*TripleManager),
		csMan:           NewCandleStickManager(nil),
		currBlockTime:   time.Unix(0, 0),
		lastBlockTime:   time.Unix(0, 0),
		tickerMap:       make(map[string]*Ticker),
		slashSlice:      make([]*NotificationSlash, 0, 10),
		partition:       0,
		offset:          0,
		lastOffset:      -1,
		dumpFlag:        0,
		lastDumpTime:    time.Now(),
		stopped:         false,
		msgEntryList:    make([]msgEntry, 0, 1000),
		blocksInterval:  interval,
		keepRecent:      keepRecent,
		msgsChannel:     make(chan MsgToPush, 10000),
		currBlockHeight: initChainHeight,
		oldChainID:      oldChainID,
		upgradeHeight:   upgradeHeight,
	}

	go hub.pushMsgToWebsocket()
	if monitorInterval <= 0 {
		return
	}
	go hub.healthMonitor(monitorInterval)
	return
}

func (hub *Hub) healthMonitor(monitorInterval int64) {
	// this goroutine monitors the progress of trade-server
	// if it is stuck, panic here
	oldDbLockCount := atomic.LoadInt64(&hub.dbLockCount)
	oldTickerMapLockCount := atomic.LoadInt64(&hub.tickerMapLockCount)
	oldTrimanLockCount := atomic.LoadInt64(&hub.trimanLockCount)
	for {
		time.Sleep(time.Duration(monitorInterval * int64(time.Second)))
		if oldTrimanLockCount == hub.trimanLockCount {
			hub.Log("No progress for a long time for triman lock count")
		}
		if oldTickerMapLockCount == hub.tickerMapLockCount {
			hub.Log("No progress for a long time for ticker map lock count")
		}
		if oldDbLockCount == hub.dbLockCount {
			panic("No progress for a long time for db lock count")
		}
		oldDbLockCount = hub.dbLockCount
		oldTickerMapLockCount = hub.tickerMapLockCount
		oldTrimanLockCount = hub.trimanLockCount
	}
}

// Not safe for concurrency; Only be used in init
func (hub *Hub) StoreLeastHeight() {
	heightBytes := Int64ToBigEndianBytes(hub.currBlockHeight)
	hub.db.Set([]byte{LatestHeightByte}, heightBytes)
}

func (hub *Hub) Log(s string) {
	log.Error(s)
}

//============================================================
var _ Consumer = &Hub{}

// Record the Msgs in msgEntryList and handle them in batch after commit
func (hub *Hub) ConsumeMessage(msgType string, bz []byte) {
	hub.preHandleNewHeightInfo(msgType, bz)
	if hub.skipHeight {
		return
	}
	hub.recordMsg(msgType, bz)
	if !hub.isTimeToHandleMsg(msgType) {
		return
	}
	if hub.skipToOldChain(msgType) {
		return
	}
	hub.handleMsg()
}

func (hub *Hub) skipToOldChain(msgType string) bool {
	if hub.oldChainID == hub.chainID && hub.currBlockHeight > hub.upgradeHeight {
		if msgType == "commit" {
			// drop height info msg
			hub.batch.Close()
			hub.batch = hub.db.NewBatch()
		}
		return true
	}
	return false
}

// analyze height_info to decide whether to skip
func (hub *Hub) preHandleNewHeightInfo(msgType string, bz []byte) {
	if msgType != "height_info" {
		return
	}

	chainID := getChainID(bz)
	v := hub.parseHeightInfo(bz, chainID == hub.oldChainID)
	if v == nil {
		return
	}
	if hub.currBlockHeight >= v.Height {
		//The incoming msg is lagging behind hub's internal state
		hub.skipHeight = true
		log.Info(fmt.Sprintf("Skipping Height %d<%d\n", hub.currBlockHeight, v.Height))

		// When receive new chain, reset height info and chain-id
		if v.ChainID != hub.oldChainID && v.Height == hub.upgradeHeight+1 {
			hub.chainID = v.ChainID
			hub.currBlockHeight = v.Height
			hub.skipHeight = false
		}
	} else if hub.currBlockHeight+1 == v.Height {
		//The incoming msg catches up hub's internal state
		hub.currBlockHeight = v.Height
		hub.skipHeight = false
		hub.chainID = v.ChainID
	} else {
		//The incoming msg can not be higher than hub's internal state
		hub.skipHeight = true
		log.Info(fmt.Sprintf("Invalid Height! %d+1!=%d\n", hub.currBlockHeight, v.Height))
		panic(fmt.Sprintf("currBlockHeight height info [%d] < v.Height[%d] msg", hub.currBlockHeight, 9))
	}
	hub.msgEntryList = hub.msgEntryList[:0]
}

func (hub *Hub) recordMsg(msgType string, bz []byte) {
	hub.msgEntryList = append(hub.msgEntryList, msgEntry{msgType: msgType, bz: bz})
}

func (hub *Hub) isTimeToHandleMsg(msgType string) bool {
	return msgType == "commit"
}

func (hub *Hub) handleMsg() {
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
		case "create_market_info":
			hub.handleCreatMarketInfo(entry.bz)
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
		case "bancor_create":
			hub.handleMsgBancorInfoForKafka(entry.bz)
		case "send_lock_coins":
			hub.handleLockedCoinsMsg(entry.bz)
		case "delegator_rewards":
			//hub.handleDelegatorRewards(entry.bz)
		case "validator_commission":
			//hub.handleValidatorCommission(entry.bz)
		case "commit":
			hub.commit()
		default:
			hub.Log(fmt.Sprintf("Unknown Message Type:%s", entry.msgType))
		}
	}
	// clear the recorded Msgs
	hub.msgEntryList = hub.msgEntryList[:0]
}

func (hub *Hub) handleNewHeightInfo(bz []byte) {

	v := hub.parseHeightInfo(bz, hub.chainID == hub.oldChainID)
	if v == nil {
		return
	}
	bz, err := json.Marshal(v)
	if err != nil {
		log.Error("marshal heightInfo failed, ", err)
		return
	}

	timestamp := uint64(v.TimeStamp)
	hub.pruneDB(v.Height, timestamp)
	latestHeight := hub.QueryLatestHeight()
	// If extra==false, then this is an invalid or out-of-date Height
	if latestHeight >= v.Height {
		hub.msgsChannel <- MsgToPush{topic: OptionKey, extra: true}
		hub.Log(fmt.Sprintf("Skipping Height websocket %d<%d\n", latestHeight, v.Height))
	} else if latestHeight+1 == v.Height || latestHeight < 0 {
		hub.msgsChannel <- MsgToPush{topic: OptionKey, extra: false}
		log.Info("push height info, height : ", v.Height)
	} else {
		hub.msgsChannel <- MsgToPush{topic: OptionKey, extra: true}
		hub.Log(fmt.Sprintf("Invalid Height! websocket %d+1!=%d\n", latestHeight, v.Height))
	}

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b[:], timestamp)
	heightBytes := Int64ToBigEndianBytes(v.Height)
	key := append([]byte{BlockHeightByte}, heightBytes...)
	hub.batch.Set(key, b)
	hub.batch.Set([]byte{LatestHeightByte}, heightBytes)
	hub.msgsChannel <- MsgToPush{topic: BlockInfoKey, bz: bz}
	hub.lastBlockTime = hub.currBlockTime
	hub.currBlockTime = time.Unix(v.TimeStamp, 0)
	hub.beginForCandleSticks()
}

// Tell the underlying DB the oldest records are useless now.
func (hub *Hub) pruneDB(height int64, timestamp uint64) {
	if height%hub.blocksInterval == 0 {
		if pdb, ok := hub.db.(Pruneable); ok && hub.keepRecent > 0 {
			pdb.SetPruneTimestamp(timestamp - uint64(hub.keepRecent))
		}
	}
}

// When a block begins, we need to run some logic related to Kline and Ticker
func (hub *Hub) beginForCandleSticks() {
	// maybe have minute, hour, day candle stick before the current block time
	var candleSticks = hub.csMan.NewBlock(hub.currBlockTime)
	for _, cs := range candleSticks {
		if !hub.updateTicker(cs) {
			continue
		}
		// push candle stick data to ws
		bz := formatCandleStick(cs)
		if bz == nil {
			continue
		}
		extra := []string{cs.Market, cs.TimeSpan}
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

func (hub *Hub) updateTicker(candleStick *CandleStick) bool {
	tripleManager, ok := hub.managersMap[candleStick.Market]
	if !ok {
		return false
	}
	// Update tickers' prices
	t := time.Unix(candleStick.EndingUnixTime, 0)
	csMinute := t.UTC().Hour()*60 + t.UTC().Minute() // minute of the data
	if candleStick.TimeSpan == MinuteStr {
		tripleManager.tkm.UpdateNewestPrice(candleStick.ClosePrice, csMinute)
	}
	return true
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
		return
	}
	bz = appendHashID(bz, hub.currTxHashID)
	key := hub.getLockedKey(v.ToAddress)
	hub.batch.Set(key, bz)
	hub.sid++
	hub.msgsChannel <- MsgToPush{topic: LockedKey, bz: bz, extra: v.ToAddress}
}

func (hub *Hub) handleDelegatorRewards(bz []byte) {
	var v NotificationDelegatorRewards
	if err := unmarshalAndLogErr(bz, &v); err != nil {
		return
	}
	key := hub.getDelegatorRewardsKey(v.Validator)
	hub.commonAction(DelegationRewardsKey, key, bz, v.Validator)
}

func (hub *Hub) handleValidatorCommission(bz []byte) {
	var v NotificationValidatorCommission
	if err := unmarshalAndLogErr(bz, &v); err != nil {
		return
	}
	key := hub.getValidatorCommissionKey(v.Validator)
	hub.commonAction(ValidatorCommissionKey, key, bz, v.Validator)
}

func (hub *Hub) commonAction(pushKey string, storeKey []byte, bz []byte, extra interface{}) {
	bz = appendHashID(bz, hub.currTxHashID)
	hub.batch.Set(storeKey, bz)
	hub.sid++
	hub.msgsChannel <- MsgToPush{topic: pushKey, bz: bz, extra: extra}
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

	// NotificationTx comes before all the kafka messages generated by this Tx
	// So we can record its tx_hash to be attached to the kafka messages
	hub.currTxHashID = v.Hash

	// Use the transaction's hashid as key, save its detail
	timeBytes := Int64ToBigEndianBytes(hub.currBlockTime.Unix())
	key := append([]byte{DetailByte}, []byte(v.Hash)...)
	key = append(key, timeBytes...)
	hub.batch.Set(key, bz)
	hub.sid++ // we do not include sid into the key, but we still increase it

	tokenNames := make(map[string]struct{})
	for _, transRec := range v.Transfers {
		tokenName := getTokenNameFromAmount(transRec.Amount)
		tokenNames[tokenName] = struct{}{}

		recipient := transRec.Recipient
		k := hub.getIncomeKey(recipient)
		hub.batch.Set(k, []byte("|"+tokenName+"|"+v.Hash))
		hub.sid++
		hub.msgsChannel <- MsgToPush{topic: IncomeKey, bz: bz, extra: recipient}
	}

	tokensAndHash := make([]byte, 1, 100)
	tokensAndHash[0] = byte('|')
	// the format is "|token1|token2|...|tx_hash"
	for tokenName := range tokenNames {
		tokensAndHash = append(tokensAndHash, []byte(tokenName+"|")...)
	}
	tokensAndHash = append(tokensAndHash, []byte(v.Hash)...)

	for _, acc := range v.Signers {
		// Notify the signers about which tokens and which tx they have signed
		signer := acc
		k := hub.getTxKey(signer)
		hub.batch.Set(k, tokensAndHash)
		hub.sid++
		hub.msgsChannel <- MsgToPush{topic: TxKey, bz: bz, extra: signer}
	}
	if len(v.ExtraInfo) == 0 {
		hub.analyzeMessages(v.MsgTypes, v.TxJSON)
	}
}

// Some important information are not shown as a dedicated kafka message,
// so we need to extract them from transactions
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
	for i, msgType := range MsgTypes {
		var donation Donation
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
	for i, msgType := range MsgTypes {
		msg, ok := msgList[i].(map[string]interface{})
		if !ok {
			continue
		}
		if msgType == "MsgCancelTradingPair" {
			market := msg["trading_pair"].(string)
			// json use float64 to represent numbers,
			// see https://stackoverflow.com/questions/22343083/json-unmarshaling-with-long-numbers-gives-floating-point-number
			effTime := msg["effective_time"].(float64)
			key := hub.getKeyFromBytes(DelistByte, []byte(market), 0)
			hub.batch.Set(key, Int64ToBigEndianBytes(int64(effTime)))
			hub.sid++
		}
	}
}

func (hub *Hub) handleNotificationBeginRedelegation(bz []byte) {
	v := hub.parseNotificationBeginRedelegation(bz)
	if v == nil {
		return
	}
	bz, err := json.Marshal(v)
	if err != nil {
		log.Error("marshal NotificationBeginRedelegation failed: ", err)
		return
	}
	bz = appendHashID(bz, hub.currTxHashID)
	t := time.Unix(v.CompletionTime, 0)
	// Use completion time as the key
	key := hub.getRedelegationEventKey(v.Delegator, t.Unix())
	hub.batch.Set(key, bz)
	hub.sid++
}

func (hub *Hub) handleNotificationBeginUnbonding(bz []byte) {
	v := hub.parseNotificationBeginUnbonding(bz)
	if v == nil {
		return
	}
	bz, err := json.Marshal(v)
	if err != nil {
		log.Error("marshal NotificationBeginUnbonding failed: ", err)
		return
	}
	bz = appendHashID(bz, hub.currTxHashID)

	// Use completion time as the key
	key := hub.getUnbondingEventKey(v.Delegator, v.CompletionTime)
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
	hub.msgsChannel <- MsgToPush{
		topic: RedelegationKey,
		extra: TimeAndSidWithAddr{
			addr:     v.Delegator,
			sid:      hub.sid,
			currTime: hub.currBlockTime.Unix(),
			lastTime: hub.lastBlockTime.Unix(),
		},
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

func (hub *Hub) handleTokenComment(bz []byte) {
	var v TokenComment
	err := json.Unmarshal(bz, &v)
	if err != nil {
		log.Error("Error in Unmarshal TokenComment")
		return
	}
	bz = appendHashID(bz, hub.currTxHashID)
	key := hub.getCommentKey(v.Token)
	hub.batch.Set(key, bz)
	hub.sid++
	hub.msgsChannel <- MsgToPush{topic: CommentKey, bz: bz, extra: v.Token}
}

func (hub *Hub) handleCreatMarketInfo(bz []byte) {
	var v MarketInfo
	err := json.Unmarshal(bz, &v)
	if err != nil {
		log.Errorf("Error in Unmarshal MarketInfo")
		return
	}
	bz = appendHashID(bz, hub.currTxHashID)
	key := hub.getCreateMarketKey(getMarketName(v))
	hub.batch.Set(key, bz)
	hub.sid++
	hub.msgsChannel <- MsgToPush{topic: CreateMarketInfoKey, bz: bz, extra: getMarketName(v)}
}

func (hub *Hub) handleCreateOrderInfo(bz []byte) {
	var v CreateOrderInfo
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal CreateOrderInfo")
		return
	}
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

	triman.AddDeltaChange(v.Side == SELL, v.Price, amount)
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
	negStock := sdk.NewInt(-v.CurrStock)
	triman.AddDeltaChange(v.Side == SELL, v.Price, negStock)
}

func (hub *Hub) handleCancelOrderInfo(bz []byte) {
	var v CancelOrderInfo
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal CancelOrderInfo")
		return
	}
	if v.DelReason == "Manually cancel the order" {
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
	negStock := sdk.NewInt(-v.LeftStock)
	triman.AddDeltaChange(v.Side == SELL, v.Price, negStock)
	//Push to subscribers
	hub.msgsChannel <- MsgToPush{topic: CancelOrderKey, bz: bz, extra: accAndSeq[0]}
}

func (hub *Hub) handleMsgBancorTradeInfoForKafka(bz []byte) {
	var v MsgBancorTradeInfoForKafka
	err := json.Unmarshal(bz, &v)
	if err != nil {
		hub.Log("Error in Unmarshal MsgBancorTradeInfoForKafka")
		return
	}
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
	//Push to subscribers
	hub.msgsChannel <- MsgToPush{topic: BancorTradeKey, bz: bz, extra: addr}
	hub.msgsChannel <- MsgToPush{topic: BancorDealKey, bz: bz, extra: v.Stock + "/" + v.Money}
}

func (hub *Hub) handleMsgBancorInfoForKafka(bz []byte) {
	v := hub.parseBancorInfo(bz)
	if v == nil {
		return
	}
	bz, err := json.Marshal(v)
	if err != nil {
		log.Error("Error in Unmarshal MsgBancorInfoForKafka")
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

func (hub *Hub) commit() {
	if hub.isStopped() {
		return
	}
	hub.commitForSlash()
	hub.commitForTicker()
	hub.commitForDepth()
	hub.pushDepthFull()
	hub.dumpHubState()
	hub.refreshDB()
}

func (hub *Hub) isStopped() bool {
	hub.dbMutex.RLock()
	defer hub.dbMutex.RUnlock()
	return hub.stopped
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

func (hub *Hub) commitForTicker() {
	tkMap := make(map[string]*Ticker)
	isNewMinute := hub.currBlockTime.UTC().Minute() != hub.lastBlockTime.UTC().Minute() ||
		hub.currBlockTime.Unix()-hub.lastBlockTime.Unix() > 60
	if !isNewMinute {
		return
	}
	currMinute := hub.currBlockTime.UTC().Hour()*60 + hub.currBlockTime.UTC().Minute() - 1
	if currMinute < 0 {
		currMinute = MinuteNumInDay - 1
	}
	for _, triman := range hub.managersMap {
		if ticker := triman.tkm.GetTicker(currMinute); ticker != nil {
			tkMap[ticker.Market] = ticker
		}
	}
	hub.msgsChannel <- MsgToPush{topic: TickerKey, extra: tkMap}
	hub.tickerMapMutex.Lock()
	defer hub.tickerMapMutex.Unlock()
	atomic.AddInt64(&hub.tickerMapLockCount, 1)
	for market, ticker := range tkMap {
		hub.tickerMap[market] = ticker
	}
}

func (hub *Hub) commitForDepth() {
	for market, triman := range hub.managersMap {
		if strings.HasPrefix(market, "B:") {
			continue
		}

		triman.Update(&hub.trimanLockCount) // consumes the recorded delta changes

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

		levelsData := encodeDepthLevels(market, mergeDeltaBuy, mergeDeltaSell)
		if bz, err := encodeDepthLevel(market, depthDeltaBuy, depthDeltaSell); err == nil {
			levelsData["all"] = bz
		}
		hub.msgsChannel <- MsgToPush{topic: DepthKey, bz: []byte(market), extra: levelsData}
	}
}

// Push full depth data every 'blocksInterval' blocks
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

func (hub *Hub) dumpHubState() {
	hub.dumpFlagLock.Lock()
	defer hub.dumpFlagLock.Unlock()
	if atomic.LoadInt32(&hub.dumpFlag) != 0 {
		atomic.StoreInt32(&hub.dumpFlag, 0)
		hub.commitForDump()
	}
}

func (hub *Hub) commitForDump() {
	// save dump data
	hub4j := &HubForJSON{}
	hub.Dump(hub4j)
	dumpKey := GetDumpKey()
	dumpBuf, err := json.Marshal(hub4j)
	if err != nil {
		log.WithError(err).Error("hub json marshal fail")
		return
	}
	hub.batch.Set(dumpKey, dumpBuf)

	// save offset
	offsetKey := GetOffsetKey(hub.partition)
	offsetBuf := Int64ToBigEndianBytes(hub.offset)
	hub.batch.Set(offsetKey, offsetBuf)

	hub.lastDumpTime = time.Now()
	log.Infof("dump data at offset: %v", hub.offset)
}

func (hub *Hub) refreshDB() {
	hub.dbMutex.Lock()
	defer hub.dbMutex.Unlock()
	atomic.AddInt64(&hub.dbLockCount, 1)
	hub.batch.WriteSync()
	hub.batch.Close()
	hub.batch = hub.db.NewBatch()
}
