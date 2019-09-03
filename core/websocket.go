package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	SeparateArgu = ":"
	MinArguNum   = 0
	MaxArguNum   = 2
)

type ImplSubscriber struct {
	*websocket.Conn
	value interface{}
}

func (i ImplSubscriber) Detail() interface{} {
	return i.value
}

func (i ImplSubscriber) WriteMsg(v []byte) error {
	return i.WriteMessage(websocket.TextMessage, v)
}

type Conn struct {
	*websocket.Conn
	topicWithParams map[string]map[string]struct{} // topic --> params
}

func NewConn(c *websocket.Conn) *Conn {
	return &Conn{
		Conn:            c,
		topicWithParams: make(map[string]map[string]struct{}),
	}
}

type WebsocketManager struct {
	sync.RWMutex

	connWithTopics map[*Conn]map[string]struct{} // conn --> topics
	topicAndConns  map[string]map[*Conn]struct{}
}

func NewWebSocketManager() *WebsocketManager {
	return &WebsocketManager{
		topicAndConns:  make(map[string]map[*Conn]struct{}),
		connWithTopics: make(map[*Conn]map[string]struct{}),
	}
}

func (w *WebsocketManager) AddConn(c *Conn) {
	w.Lock()
	defer w.Unlock()
	w.connWithTopics[c] = make(map[string]struct{})
}

func (w *WebsocketManager) CloseConn(c *Conn) error {
	w.Lock()
	defer w.Unlock()
	topics, ok := w.connWithTopics[c]
	if !ok {
		panic("the remove conn not cache in websocketManager ")
	}

	for topic := range topics {
		conns, ok := w.topicAndConns[topic]
		if ok {
			delete(conns, c)
		}
	}
	delete(w.connWithTopics, c)
	return c.Close()
}

func (w *WebsocketManager) AddSubscribeConn(subscriptionTopic string, c *Conn) error {
	w.Lock()
	defer w.Unlock()
	values := strings.Split(subscriptionTopic, SeparateArgu)
	if len(values) < 1 || len(values) > MaxArguNum+1 {
		return fmt.Errorf("Expect range of parameters [%d, %d], actual : %d ", MinArguNum, MaxArguNum, len(values)-1)
	}

	topic := values[0]
	params := values[1:]
	if !checkTopicValid(topic) {
		log.Errorf("The subscribed topic [%s] is illegal ", topic)
		return fmt.Errorf("The subscribed topic [%s] is illegal ", topic)
	}

	if len(params) != 0 {
		if len(c.topicWithParams[topic]) == 0 {
			c.topicWithParams[topic] = make(map[string]struct{})
		}
		if len(params) == 1 {
			c.topicWithParams[topic][params[0]] = struct{}{}
		} else {
			c.topicWithParams[topic][strings.Join(params, SeparateArgu)] = struct{}{}
		}
	}
	if len(w.topicAndConns[topic]) == 0 {
		w.topicAndConns[topic] = make(map[*Conn]struct{})
	}
	w.topicAndConns[topic][c] = struct{}{}
	w.connWithTopics[c][topic] = struct{}{}
	return nil
}

func (w *WebsocketManager) RemoveSubscribeConn(topic string, c *Conn) error {
	w.Lock()
	defer w.Unlock()
	if !checkTopicValid(topic) {
		log.Errorf("The subscribed topic [%s] is illegal ", topic)
		return fmt.Errorf("The subscribed topic [%s] is illegal ", topic)
	}
	if topics, ok := w.connWithTopics[c]; ok {
		if _, ok := topics[topic]; ok {
			delete(topics, topic)
		}
	}
	if conns, ok := w.topicAndConns[topic]; ok {
		if _, ok := conns[c]; ok {
			delete(conns, c)
		}
	}
	return nil
}

func checkTopicValid(topic string) bool {
	switch topic {
	case BlockInfoKey, SlashKey, TickerKey,
		KlineKey, DepthKey, DealKey, BancorKey,
		BancorTradeKey, CommentKey, OrderKey,
		IncomeKey, UnbondingKey, RedelegationKey,
		UnlockKey, TxKey, LockedKey:
		return true
	default:
		return false

	}
}

func (w *WebsocketManager) GetSlashSubscribeInfo() []Subscriber {
	w.RLock()
	defer w.RUnlock()
	conns := w.topicAndConns[SlashKey]
	res := make([]Subscriber, 0, len(conns))
	for conn := range conns {
		res = append(res, ImplSubscriber{Conn: conn.Conn})
	}
	return res
}

func (w *WebsocketManager) GetHeightSubscribeInfo() []Subscriber {
	w.RLock()
	defer w.RUnlock()
	conns := w.topicAndConns[BlockInfoKey]
	res := make([]Subscriber, 0, len(conns))
	for conn := range conns {
		res = append(res, ImplSubscriber{Conn: conn.Conn})
	}
	return res
}

func (w *WebsocketManager) GetTickerSubscribeInfo() []Subscriber {
	w.RLock()
	defer w.RUnlock()
	conns := w.topicAndConns[TickerKey]
	res := make([]Subscriber, 0, len(conns))
	for conn := range conns {
		res = append(res, ImplSubscriber{
			Conn:  conn.Conn,
			value: conn.topicWithParams[TickerKey],
		})
	}

	return res
}

func (w *WebsocketManager) GetCandleStickSubscribeInfo() map[string][]Subscriber {
	w.RLock()
	defer w.RUnlock()
	conns := w.topicAndConns[KlineKey]
	res := make(map[string][]Subscriber)
	for conn := range conns {
		params := conn.topicWithParams[KlineKey]
		for p := range params {
			vals := strings.Split(p, SeparateArgu)
			res[vals[0]] = append(res[vals[0]], ImplSubscriber{
				Conn:  conn.Conn,
				value: vals[1],
			})
		}
	}

	return res
}

func (w *WebsocketManager) getNoDetailSubscribe(topic string) map[string][]Subscriber {
	w.RLock()
	defer w.RUnlock()
	conns := w.topicAndConns[topic]
	res := make(map[string][]Subscriber)
	for conn := range conns {
		for param := range conn.topicWithParams[topic] {
			res[param] = append(res[param], ImplSubscriber{
				Conn: conn.Conn,
			})
		}
	}

	return res
}

func (w *WebsocketManager) GetDepthSubscribeInfo() map[string][]Subscriber {
	return w.getNoDetailSubscribe(DepthKey)
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
func sendEncodeMsg(subscriber Subscriber, typeKey string, info []byte) {
	msg := []byte(fmt.Sprintf("{\"type\":\"%s\", \"payload\":%s}", typeKey, string(info)))
	if err := subscriber.WriteMsg(msg); err != nil {
		log.Errorf(err.Error())
	}
}

func (w *WebsocketManager) PushLockedSendMsg(subscriber Subscriber, info []byte) {
	sendEncodeMsg(subscriber, LockedKey, info)
}

func (w *WebsocketManager) PushSlash(subscriber Subscriber, info []byte) {
	sendEncodeMsg(subscriber, SlashKey, info)
}

func (w *WebsocketManager) PushHeight(subscriber Subscriber, info []byte) {
	sendEncodeMsg(subscriber, BlockInfoKey, info)
}
func (w *WebsocketManager) PushTicker(subscriber Subscriber, t []*Ticker) {
	payload, err := json.Marshal(t)
	if err != nil {
		log.Error(err)
		return
	}
	sendEncodeMsg(subscriber, TickerKey, payload)
}
func (w *WebsocketManager) PushDepthSell(subscriber Subscriber, delta []byte) {
	sendEncodeMsg(subscriber, DepthKey, delta)
}
func (w *WebsocketManager) PushDepthBuy(subscriber Subscriber, delta []byte) {
	sendEncodeMsg(subscriber, DepthKey, delta)
}
func (w *WebsocketManager) PushCandleStick(subscriber Subscriber, info []byte) {
	sendEncodeMsg(subscriber, KlineKey, info)
}
func (w *WebsocketManager) PushDeal(subscriber Subscriber, info []byte) {
	sendEncodeMsg(subscriber, DealKey, info)
}
func (w *WebsocketManager) PushCreateOrder(subscriber Subscriber, info []byte) {
	sendEncodeMsg(subscriber, OrderKey, info)
}
func (w *WebsocketManager) PushFillOrder(subscriber Subscriber, info []byte) {
	sendEncodeMsg(subscriber, OrderKey, info)
}
func (w *WebsocketManager) PushCancelOrder(subscriber Subscriber, info []byte) {
	sendEncodeMsg(subscriber, OrderKey, info)
}
func (w *WebsocketManager) PushBancorInfo(subscriber Subscriber, info []byte) {
	sendEncodeMsg(subscriber, BancorKey, info)
}
func (w *WebsocketManager) PushBancorTrade(subscriber Subscriber, info []byte) {
	sendEncodeMsg(subscriber, BancorTradeKey, info)
}
func (w *WebsocketManager) PushIncome(subscriber Subscriber, info []byte) {
	sendEncodeMsg(subscriber, IncomeKey, info)
}
func (w *WebsocketManager) PushUnbonding(subscriber Subscriber, info []byte) {
	sendEncodeMsg(subscriber, UnbondingKey, info)
}
func (w *WebsocketManager) PushRedelegation(subscriber Subscriber, info []byte) {
	sendEncodeMsg(subscriber, RedelegationKey, info)
}
func (w *WebsocketManager) PushUnlock(subscriber Subscriber, info []byte) {
	sendEncodeMsg(subscriber, UnlockKey, info)
}
func (w *WebsocketManager) PushTx(subscriber Subscriber, info []byte) {
	sendEncodeMsg(subscriber, TxKey, info)
}
func (w *WebsocketManager) PushComment(subscriber Subscriber, info []byte) {
	sendEncodeMsg(subscriber, CommentKey, info)
}
