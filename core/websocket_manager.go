package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

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

	SkipPushed   bool // skip the messages to be pushed to subscribers, used during startup
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

// Wrap this WsInterface with 'Conn' and return this 'Conn'
// The map 'wsConn2Conn' ensures for the same WsInterface, we only wrap it once
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
	//for topic := range c.allTopics {
	//	if conns, ok := w.topics2Conns[topic]; ok {
	//		delete(conns, c)
	//	}
	//}
	for topic := range w.topics2Conns { //Why? Do these code look more clear?
		delete(w.topics2Conns[topic], c)
	}
	delete(w.wsConn2Conn, c.WsIfc)
	if err := c.Close(); err != nil {
		log.WithError(err).Error(fmt.Sprintf("close connection failed"))
	}
}

func (w *WebsocketManager) AddSubscribeConn(c *Conn, topic string, params []string) error {
	//The modification to w and c should be atomic, so we need two locks before modification
	w.mtx.Lock()
	defer w.mtx.Unlock()
	c.mtx.Lock()
	defer c.mtx.Unlock()
	w.addConnWithTopic(topic, c)
	c.addTopicAndParams(topic, params)
	return nil
}

func (w *WebsocketManager) PushFullInfo(hub *Hub, c *Conn, topic string, params []string, count int) error {
	err := pushFullInformation(topic, params, count, ImplSubscriber{Conn: c}, hub)
	return err
}

func (w *WebsocketManager) addConnWithTopic(topic string, conn *Conn) {
	if len(w.topics2Conns[topic]) == 0 {
		w.topics2Conns[topic] = make(map[*Conn]struct{})
	}
	w.topics2Conns[topic][conn] = struct{}{}
}

func (w *WebsocketManager) RemoveSubscribeConn(c *Conn, topic string, params []string) error {
	//The modification to w and c should be atomic, so we need two locks before modification
	c.mtx.Lock()
	defer c.mtx.Unlock()
	w.mtx.Lock()
	defer w.mtx.Unlock()
	c.removeTopicAndParams(topic, params)
	w.removeConnWithTopic(topic, c)
	return nil
}

func (w *WebsocketManager) removeConnWithTopic(topic string, conn *Conn) {
	if conns, ok := w.topics2Conns[topic]; ok {
		if _, ok := conns[conn]; ok {
			if conn.topicHasEmptyParamSet(topic) {
				delete(conns, conn)
			}
		}
	}
	if len(w.topics2Conns[topic]) == 0 { // this topic has no Conns
		delete(w.topics2Conns, topic)
	}
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

// The key of the result map is market
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

// Get the Subscribers which doesn't need detail information
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

// The key of the result map is market
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

func (w *WebsocketManager) GetMarketSubscribeInfo() map[string][]Subscriber {
	return w.getNoDetailSubscribe(CreateMarketInfoKey)
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
func (w *WebsocketManager) GetDelegationRewards() map[string][]Subscriber {
	return w.getNoDetailSubscribe(DelegationRewardsKey)
}
func (w *WebsocketManager) GetValidatorCommissionInfo() map[string][]Subscriber {
	return w.getNoDetailSubscribe(ValidatorCommissionKey)
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
func (w *WebsocketManager) PushCreateMarket(subscriber Subscriber, info []byte) {
	w.sendEncodeMsg(subscriber, CreateMarketInfoKey, info)
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
func (w *WebsocketManager) PushValidatorCommissionInfo(subscriber Subscriber, info []byte) {
	w.sendEncodeMsg(subscriber, ValidatorCommissionKey, info)
}
func (w *WebsocketManager) PushDelegationRewards(subscriber Subscriber, info []byte) {
	w.sendEncodeMsg(subscriber, DelegationRewardsKey, info)
}
