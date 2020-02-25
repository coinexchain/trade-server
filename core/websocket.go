package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	SeparateArgu = ":"
	MaxArguNum   = 3
)

type ImplSubscriber struct {
	Conn  *Conn
	value interface{}
}

func (i ImplSubscriber) Detail() interface{} {
	return i.value
}

func (i ImplSubscriber) WriteMsg(v []byte) error {
	return i.Conn.WriteMsg(v)
}

func (i ImplSubscriber) GetConn() *Conn {
	return i.Conn
}

type WebsocketManager struct {
	mtx sync.RWMutex

	SkipPushed   bool
	wsConn2Conn  map[WsInterface]*Conn
	topics2Conns map[string]map[*Conn]struct{}
}

func NewWebSocketManager() *WebsocketManager {
	return &WebsocketManager{
		topics2Conns: make(map[string]map[*Conn]struct{}),
		wsConn2Conn:  make(map[WsInterface]*Conn),
	}
}

func (w *WebsocketManager) SetSkipOption(isSkip bool) {
	w.SkipPushed = isSkip
}

func (w *WebsocketManager) AddWsConn(c WsInterface) *Conn {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	if _, ok := w.wsConn2Conn[c]; !ok {
		w.wsConn2Conn[c] = NewConn(c)
	}
	return w.wsConn2Conn[c]
}

func (w *WebsocketManager) CloseWsConn(c *Conn) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	for topic := range c.allTopics {
		if conns, ok := w.topics2Conns[topic]; ok {
			delete(conns, c)
		}
	}
	delete(w.wsConn2Conn, c.Conn)
	if err := c.Close(); err != nil {
		log.WithError(err).Error(fmt.Sprintf("close connection failed"))
	}
}

func (w *WebsocketManager) AddSubscribeConn(c *Conn, topic string, params []string) error {
	w.addConnWithTopic(topic, c)
	c.addTopicAndParams(topic, params)
	return nil
}

func (w *WebsocketManager) PushFullInfo(hub *Hub, c *Conn, topic string, params []string, count int) error {
	err := PushFullInformation(topic, params, count, ImplSubscriber{Conn: c}, hub)
	return err
}

func (w *WebsocketManager) addConnWithTopic(topic string, conn *Conn) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	if len(w.topics2Conns[topic]) == 0 {
		w.topics2Conns[topic] = make(map[*Conn]struct{})
	}
	w.topics2Conns[topic][conn] = struct{}{}
}

func (w *WebsocketManager) RemoveSubscribeConn(c *Conn, topic string, params []string) error {
	c.removeTopicAndParams(topic, params)
	w.removeConnWithTopic(topic, c)
	return nil
}

func (w *WebsocketManager) removeConnWithTopic(topic string, conn *Conn) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	if conns, ok := w.topics2Conns[topic]; ok {
		if _, ok := conns[conn]; ok {
			if conn.isCleanedTopic(topic) {
				delete(conns, conn)
			}
		}
	}
	if len(w.topics2Conns[topic]) == 0 {
		delete(w.topics2Conns, topic)
	}
}

func groupOfDataPacket(topic string, data []json.RawMessage) []byte {
	bz := make([]byte, 0, len(data))
	bz = append(bz, []byte(fmt.Sprintf("{\"type\":\"%s\", \"payload\":[", topic))...)
	for _, v := range data {
		bz = append(bz, []byte(v)...)
		bz = append(bz, []byte(",")...)
	}
	if len(data) > 0 {
		bz[len(bz)-1] = byte(']')
		bz = append(bz, []byte("}")...)
	} else {
		bz = append(bz, []byte("]}")...)
	}
	return bz
}

func PushFullInformation(topic string, params []string, count int, c Subscriber, hub *Hub) error {
	var (
		err        error
		depthLevel = "all"
	)
	count = getCount(count)
	type queryFunc func(string, int64, int64, int) ([]json.RawMessage, []int64)
	queryAndPushFunc := func(typeKey string, param string, qf queryFunc) error {
		data, _ := qf(param, hub.currBlockTime.Unix(), hub.sid, count)
		bz := groupOfDataPacket(typeKey, data)
		err = c.WriteMsg(bz)
		if err != nil {
			return err
		}
		return nil
	}
	if len(params) == 2 && topic == DepthKey {
		depthLevel = params[1]
	}
	switch topic {
	case SlashKey:
		err = querySlashAndPush(hub, c, count)
	case KlineKey:
		err = queryKlineAndpush(hub, c, params, count)
	case DepthKey:
		err = queryDepthAndPush(hub, c, params[0], depthLevel, count)
	case OrderKey:
		err = queryOrderAndPush(hub, c, params[0], count)
	case TickerKey:
		market := params[0]
		if len(params) == 2 {
			market = strings.Join(params, SeparateArgu)
		}
		err = queryTickerAndPush(hub, c, market)
	case TxKey:
		err = queryAndPushFunc(TxKey, params[0], hub.QueryTx)
	case LockedKey:
		err = queryAndPushFunc(LockedKey, params[0], hub.QueryLocked)
	case UnlockKey:
		err = queryAndPushFunc(UnlockKey, params[0], hub.QueryUnlock)
	case IncomeKey:
		err = queryAndPushFunc(IncomeKey, params[0], hub.QueryIncome)
	case DealKey:
		err = queryAndPushFunc(DealKey, params[0], hub.QueryDeal)
	case BancorKey:
		err = queryAndPushFunc(BancorKey, params[0], hub.QueryBancorInfo)
	case BancorTradeKey:
		err = queryAndPushFunc(BancorTradeKey, params[0], hub.QueryBancorTrade)
	case BancorDealKey:
		err = queryAndPushFunc(BancorDealKey, "B:"+params[0], hub.QueryBancorDeal)
	case RedelegationKey:
		err = queryAndPushFunc(RedelegationKey, params[0], hub.QueryRedelegation)
	case UnbondingKey:
		err = queryAndPushFunc(UnbondingKey, params[0], hub.QueryUnbonding)
	case CommentKey:
		err = queryAndPushFunc(CommentKey, params[0], hub.QueryComment)
	}
	return err
}

func queryTickerAndPush(hub *Hub, c Subscriber, market string) error {
	tickers := hub.QueryTickers([]string{market})
	baseData, err := json.Marshal(tickers)
	if err != nil {
		return err
	}
	err = c.WriteMsg([]byte(fmt.Sprintf("{\"type\":\"%s\","+
		" \"payload\":%s}", TickerKey, string(baseData))))
	return err
}

func queryOrderAndPush(hub *Hub, c Subscriber, account string, count int) error {
	data, tags, _ := hub.QueryOrder(account, hub.currBlockTime.Unix(), hub.sid, count)
	if len(data) != len(tags) {
		return errors.Errorf("The number of orders and tags is not equal")
	}
	createData := make([]json.RawMessage, 0, len(data)/2)
	fillData := make([]json.RawMessage, 0, len(data)/2)
	cancelData := make([]json.RawMessage, 0, len(data)/2)
	for i := len(data) - 1; i >= 0; i-- {
		if tags[i] == CreateOrderEndByte {
			createData = append(createData, data[i])
		} else if tags[i] == FillOrderEndByte {
			fillData = append(fillData, data[i])
		} else if tags[i] == CancelOrderEndByte {
			cancelData = append(cancelData, data[i])
		}
	}
	bz := groupOfDataPacket(CreateOrderKey, createData)
	if err := c.WriteMsg(bz); err != nil {
		return err
	}
	bz = groupOfDataPacket(FillOrderKey, fillData)
	if err := c.WriteMsg(bz); err != nil {
		return err
	}
	bz = groupOfDataPacket(CancelOrderKey, cancelData)
	err := c.WriteMsg(bz)
	return err
}

func queryDepthAndPush(hub *Hub, c Subscriber, market string, level string, count int) error {
	bz, err := getDepthFullData(hub, market, level, count)
	if err != nil {
		return err
	}
	if err := c.WriteMsg(bz); err != nil {
		return err
	}
	return hub.AddLevel(market, level)
}

func getDepthFullData(hub *Hub, market string, level string, count int) ([]byte, error) {
	var (
		bz  []byte
		err error
	)
	sell, buy := hub.QueryDepth(market, count)
	if level == "all" {
		bz, err = encodeDepthData(market, buy, sell)
		return bz, err
	}

	buyLevel := mergePrice(buy, level, true)
	sellLevel := mergePrice(sell, level, false)
	bz, err = encodeDepthLevel(market, buyLevel, sellLevel)
	return bz, nil
}

func queryKlineAndpush(hub *Hub, c Subscriber, params []string, count int) error {
	tradingPair := params[0]
	if len(params) == 3 {
		tradingPair = strings.Join(params[:2], SeparateArgu)
	}
	candleBz := hub.QueryCandleStick(tradingPair, GetSpanFromSpanStr(params[1]), hub.currBlockTime.Unix(), hub.sid, count)
	bz := groupOfDataPacket(KlineKey, candleBz)
	err := c.WriteMsg(bz)
	if err != nil {
		return err
	}
	return nil
}

func querySlashAndPush(hub *Hub, c Subscriber, count int) error {
	data, _ := hub.QuerySlash(hub.currBlockTime.Unix(), hub.sid, count)
	bz := groupOfDataPacket(SlashKey, data)
	err := c.WriteMsg(bz)
	if err != nil {
		return err
	}
	return nil
}

func (w *WebsocketManager) GetSlashSubscribeInfo() []Subscriber {
	w.mtx.RLock()
	defer w.mtx.RUnlock()
	conns := w.topics2Conns[SlashKey]
	res := make([]Subscriber, 0, len(conns))
	for conn := range conns {
		res = append(res, ImplSubscriber{Conn: conn})
	}
	return res
}

func (w *WebsocketManager) GetHeightSubscribeInfo() []Subscriber {
	w.mtx.RLock()
	defer w.mtx.RUnlock()
	conns := w.topics2Conns[BlockInfoKey]
	res := make([]Subscriber, 0, len(conns))
	for conn := range conns {
		res = append(res, ImplSubscriber{Conn: conn})
	}
	return res
}

func (w *WebsocketManager) GetTickerSubscribeInfo() []Subscriber {
	w.mtx.RLock()
	defer w.mtx.RUnlock()
	conns := w.topics2Conns[TickerKey]
	res := make([]Subscriber, 0, len(conns))
	for conn := range conns {
		res = append(res, ImplSubscriber{
			Conn:  conn,
			value: conn.topicWithParams[TickerKey],
		})
	}
	return res
}

func (w *WebsocketManager) GetCandleStickSubscribeInfo() map[string][]Subscriber {
	w.mtx.RLock()
	defer w.mtx.RUnlock()
	conns := w.topics2Conns[KlineKey]
	res := make(map[string][]Subscriber)
	for conn := range conns {
		params := conn.topicWithParams[KlineKey]
		for p := range params {
			vals := strings.Split(p, SeparateArgu)
			market := vals[0]
			timeSpan := vals[1]
			if len(vals) == 3 {
				market = strings.Join(vals[:2], SeparateArgu)
				timeSpan = vals[2]
			}
			res[market] = append(res[vals[0]], ImplSubscriber{
				Conn:  conn,
				value: timeSpan,
			})
		}
	}

	return res
}

func (w *WebsocketManager) getNoDetailSubscribe(topic string) map[string][]Subscriber {
	w.mtx.RLock()
	defer w.mtx.RUnlock()
	conns := w.topics2Conns[topic]
	res := make(map[string][]Subscriber)
	for conn := range conns {
		for param := range conn.topicWithParams[topic] {
			res[param] = append(res[param], ImplSubscriber{
				Conn: conn,
			})
		}
	}

	return res
}

func (w *WebsocketManager) GetDepthSubscribeInfo() map[string][]Subscriber {
	w.mtx.RLock()
	defer w.mtx.RUnlock()
	conns := w.topics2Conns[DepthKey]
	res := make(map[string][]Subscriber)
	for conn := range conns {
		params := conn.topicWithParams[DepthKey]
		for p := range params {
			var (
				level  = "all"
				vals   = strings.Split(p, SeparateArgu)
				market = vals[0]
			)
			if len(vals) == 2 {
				level = vals[1]
			}
			res[market] = append(res[market], ImplSubscriber{
				Conn:  conn,
				value: level,
			})
		}
	}
	return res
}

func (w *WebsocketManager) GetDealSubscribeInfo() map[string][]Subscriber {
	return w.getNoDetailSubscribe(DealKey)
}
func (w *WebsocketManager) GetBancorInfoSubscribeInfo() map[string][]Subscriber {
	return w.getNoDetailSubscribe(BancorKey)
}
func (w *WebsocketManager) GetCommentSubscribeInfo() map[string][]Subscriber {
	return w.getNoDetailSubscribe(CommentKey)
}
func (w *WebsocketManager) GetOrderSubscribeInfo() map[string][]Subscriber {
	return w.getNoDetailSubscribe(OrderKey)
}
func (w *WebsocketManager) GetBancorTradeSubscribeInfo() map[string][]Subscriber {
	return w.getNoDetailSubscribe(BancorTradeKey)
}
func (w *WebsocketManager) GetBancorDealSubscribeInfo() map[string][]Subscriber {
	return w.getNoDetailSubscribe(BancorDealKey)
}
func (w *WebsocketManager) GetIncomeSubscribeInfo() map[string][]Subscriber {
	return w.getNoDetailSubscribe(IncomeKey)
}
func (w *WebsocketManager) GetUnbondingSubscribeInfo() map[string][]Subscriber {
	return w.getNoDetailSubscribe(UnbondingKey)
}
func (w *WebsocketManager) GetRedelegationSubscribeInfo() map[string][]Subscriber {
	return w.getNoDetailSubscribe(RedelegationKey)
}
func (w *WebsocketManager) GetUnlockSubscribeInfo() map[string][]Subscriber {
	return w.getNoDetailSubscribe(UnlockKey)
}
func (w *WebsocketManager) GetTxSubscribeInfo() map[string][]Subscriber {
	return w.getNoDetailSubscribe(TxKey)
}

func (w *WebsocketManager) GetLockedSubscribeInfo() map[string][]Subscriber {
	return w.getNoDetailSubscribe(LockedKey)
}

// Push msgs----------------------------
func (w *WebsocketManager) sendEncodeMsg(subscriber Subscriber, typeKey string, info []byte) {
	if !w.SkipPushed {
		msg := []byte(fmt.Sprintf("{\"type\":\"%s\", \"payload\":%s}", typeKey, string(info)))
		if err := subscriber.WriteMsg(msg); err != nil {
			log.Errorf(err.Error())
			w.CloseWsConn(subscriber.(ImplSubscriber).Conn)
		}
	}
}
func (w *WebsocketManager) PushLockedSendMsg(subscriber Subscriber, info []byte) {
	w.sendEncodeMsg(subscriber, LockedKey, info)
}
func (w *WebsocketManager) PushSlash(subscriber Subscriber, info []byte) {
	w.sendEncodeMsg(subscriber, SlashKey, info)
}
func (w *WebsocketManager) PushHeight(subscriber Subscriber, info []byte) {
	w.sendEncodeMsg(subscriber, BlockInfoKey, info)
}
func (w *WebsocketManager) PushTicker(subscriber Subscriber, t []*Ticker) {
	payload, err := json.Marshal(t)
	if err != nil {
		log.Error(err)
		return
	}
	w.sendEncodeMsg(subscriber, TickerKey, payload)
}
func (w *WebsocketManager) PushDepthFullMsg(subscriber Subscriber, info []byte) {
	w.sendEncodeMsg(subscriber, DepthFull, info)
}
func (w *WebsocketManager) PushDepthWithChange(subscriber Subscriber, info []byte) {
	w.sendEncodeMsg(subscriber, DepthChange, info)
}
func (w *WebsocketManager) PushDepthWithDelta(subscriber Subscriber, delta []byte) {
	w.sendEncodeMsg(subscriber, DepthDelta, delta)
}
func (w *WebsocketManager) PushCandleStick(subscriber Subscriber, info []byte) {
	w.sendEncodeMsg(subscriber, KlineKey, info)
}
func (w *WebsocketManager) PushDeal(subscriber Subscriber, info []byte) {
	w.sendEncodeMsg(subscriber, DealKey, info)
}
func (w *WebsocketManager) PushCreateOrder(subscriber Subscriber, info []byte) {
	w.sendEncodeMsg(subscriber, CreateOrderKey, info)
}
func (w *WebsocketManager) PushFillOrder(subscriber Subscriber, info []byte) {
	w.sendEncodeMsg(subscriber, FillOrderKey, info)
}
func (w *WebsocketManager) PushCancelOrder(subscriber Subscriber, info []byte) {
	w.sendEncodeMsg(subscriber, CancelOrderKey, info)
}
func (w *WebsocketManager) PushBancorInfo(subscriber Subscriber, info []byte) {
	w.sendEncodeMsg(subscriber, BancorKey, info)
}
func (w *WebsocketManager) PushBancorTrade(subscriber Subscriber, info []byte) {
	w.sendEncodeMsg(subscriber, BancorTradeKey, info)
}
func (w *WebsocketManager) PushBancorDeal(subscriber Subscriber, info []byte) {
	w.sendEncodeMsg(subscriber, BancorDealKey, info)
}
func (w *WebsocketManager) PushIncome(subscriber Subscriber, info []byte) {
	w.sendEncodeMsg(subscriber, IncomeKey, info)
}
func (w *WebsocketManager) PushUnbonding(subscriber Subscriber, info []byte) {
	w.sendEncodeMsg(subscriber, UnbondingKey, info)
}
func (w *WebsocketManager) PushRedelegation(subscriber Subscriber, info []byte) {
	w.sendEncodeMsg(subscriber, RedelegationKey, info)
}
func (w *WebsocketManager) PushUnlock(subscriber Subscriber, info []byte) {
	w.sendEncodeMsg(subscriber, UnlockKey, info)
}
func (w *WebsocketManager) PushTx(subscriber Subscriber, info []byte) {
	w.sendEncodeMsg(subscriber, TxKey, info)
}
func (w *WebsocketManager) PushComment(subscriber Subscriber, info []byte) {
	w.sendEncodeMsg(subscriber, CommentKey, info)
}
