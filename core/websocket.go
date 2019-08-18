package core

import (
	"encoding/json"
	"fmt"
	"os"
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

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.WarnLevel)

	if _, err := os.Stat("log"); err != nil && os.IsNotExist(err) {
		if err = os.Mkdir("log", 0755); err != nil {
			log.Fatal(err)
		}
	}

	file, err := os.OpenFile("log/log.out", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
	log.SetReportCaller(true)
}

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
	topicWithParams map[string][]string // topic --> params
}

func NewConn(c *websocket.Conn) *Conn {
	return &Conn{
		Conn:            c,
		topicWithParams: make(map[string][]string),
	}
}

type ConnWithParams map[string]map[*Conn]struct{}

type WebsocketManager struct {
	sync.RWMutex

	subs           map[string]ConnWithParams     // topic --> params
	connWithTopics map[*Conn]map[string]struct{} // conn --> topics

	topicAndConns map[string]map[*Conn]struct{}
}

func NewWebSocketManager() *WebsocketManager {
	return &WebsocketManager{
		subs:           make(map[string]ConnWithParams),
		connWithTopics: make(map[*Conn]map[string]struct{}),
	}
}

func (w *WebsocketManager) CloseConn(c *Conn) error {
	w.Lock()
	defer w.Unlock()

	topics, ok := w.connWithTopics[c]
	if !ok {
		panic("the remove conn not cache in websocketManager : ")
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
	if len(values) < MinArguNum+1 || len(values) > MaxArguNum+1 {
		return fmt.Errorf("Expect range of parameters [%d, %d], actual : %d ", MinArguNum, MaxArguNum, len(values)-1)
	}

	topic := values[0]
	params := values[1:]
	if len(params) != MinArguNum {
		c.topicWithParams[topic] = append(c.topicWithParams[topic], params...)
	}
	w.topicAndConns[topic][c] = struct{}{}
	return nil
}

func (w *WebsocketManager) RemoveSubscribeConn(c *Conn, topic string) {
	w.Lock()
	defer w.Unlock()

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
}

func (w *WebsocketManager) GetSlashSubscribeInfo() []Subscriber {
	conns := w.topicAndConns[SlashKey]
	res := make([]Subscriber, 0, len(conns))
	for conn := range conns {
		res = append(res, ImplSubscriber{Conn: conn.Conn})
	}
	return res
}

func (w *WebsocketManager) GetHeightSubscribeInfo() []Subscriber {
	conns := w.topicAndConns[BlockInfoKey]
	res := make([]Subscriber, 0, len(conns))
	for conn := range conns {
		res = append(res, ImplSubscriber{Conn: conn.Conn})
	}
	return res
}

func (w *WebsocketManager) GetTickerSubscribeInfo() []Subscriber {
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
	conns := w.topicAndConns[KlineKey]
	res := make(map[string][]Subscriber)
	for conn := range conns {
		params := conn.topicWithParams[KlineKey]
		for i := 0; i < len(params); i += 2 {
			res[params[i]] = append(res[params[i]], ImplSubscriber{
				Conn:  conn.Conn,
				value: params[i+1],
			})
		}
	}

	return res
}

func (w *WebsocketManager) getNoDetailSubscribe(topic string) map[string][]Subscriber {
	conns := w.topicAndConns[topic]
	res := make(map[string][]Subscriber)
	for conn := range conns {
		params := conn.topicWithParams[topic]
		for _, param := range params {
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

// Push msgs----------------------------

func assembleJSON(typeKey string, info []byte) []byte {
	m := make(map[string]string)
	m[Type] = typeKey
	m[Payload] = ""
	bz, _ := json.Marshal(m)
	msg := string(bz[:len(bz)-3]) + string(info) + "}"
	return []byte(msg)
}

func sendEncodeMsg(subscriber Subscriber, typeKey string, info []byte) {
	msg := assembleJSON(typeKey, info)
	if err := subscriber.WriteMsg(msg); err != nil {
		log.Errorf(err.Error())
	}
}

func (w *WebsocketManager) PushSlash(subscriber Subscriber, info []byte) {
	sendEncodeMsg(subscriber, SlashKey, info)
}

func (w *WebsocketManager) PushHeight(subscriber Subscriber, info []byte) {
	sendEncodeMsg(subscriber, BlockInfoKey, info)
}
func (w *WebsocketManager) PushTicker(subscriber Subscriber, t []*Ticker) {
	type ticker struct {
		Type    string
		Payload []*Ticker
	}

	msg := ticker{Type: TickerKey, Payload: t}
	bz, err := json.Marshal(msg)
	if err != nil {
		log.Error(err)
		return
	}
	if err := subscriber.WriteMsg(bz); err != nil {
		log.Error(err)
		return
	}
}
func (w *WebsocketManager) PushDepthSell(subscriber Subscriber, delta []byte) {
	if err := subscriber.WriteMsg(delta); err != nil {
		log.Error(err)
	}
}
func (w *WebsocketManager) PushDepthBuy(subscriber Subscriber, delta []byte) {
	if err := subscriber.WriteMsg(delta); err != nil {
		log.Error(err)
	}
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
