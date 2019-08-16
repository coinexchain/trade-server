package core

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

const (
	SeparateArgu = ":"
	MinArguNum   = 0
	MaxArguNum   = 2
	NullParam    = ""
)

type Conn struct {
	websocket.Conn
	detail interface{}
}

func (c Conn) Detail() interface{} {
	return c.detail
}

type ConnWithParams map[string]map[Conn]struct{}

type WebsocketManager struct {
	sync.RWMutex

	subs           map[string]ConnWithParams    // topic --> params
	connWithTopics map[Conn]map[string]struct{} // conn --> topics
}

func NewWebSocketManager() *WebsocketManager {
	return &WebsocketManager{
		subs:           make(map[string]ConnWithParams),
		connWithTopics: make(map[Conn]map[string]struct{}),
	}
}

func (w *WebsocketManager) AddSubscribeConn(subscriptionTopic string, c Conn) error {
	w.Lock()
	defer w.Unlock()
	values := strings.Split(subscriptionTopic, SeparateArgu)
	if len(values) < MinArguNum+1 || len(values) > MaxArguNum+1 {
		return fmt.Errorf("Expect range of parameters [%d, %d], actual : %d ", MinArguNum, MaxArguNum, len(values)-1)
	}

	topic := values[0]
	params := values[1:]
	if len(params) == MinArguNum {
		w.subs[topic][NullParam][c] = struct{}{}
	} else if len(params) == MaxArguNum {
		c.detail = params[1]
		w.subs[topic][params[0]][c] = struct{}{}
	} else {
		w.subs[topic][params[0]][c] = struct{}{}
	}
	w.connWithTopics[c][topic] = struct{}{}

	return nil
}

func (w *WebsocketManager) CloseConn(c Conn) error {
	w.Lock()
	defer w.Unlock()

	topics, ok := w.connWithTopics[c]
	if !ok {
		panic("the remove conn not cache in websocketManager : ")
	}

	for topic := range topics {
		delete(w.subs, topic)
	}
	delete(w.connWithTopics, c)
	return c.Close()
}

func (w *WebsocketManager) RemoveSubscribeConn(c Conn, topic string) {
	w.Lock()
	defer w.Unlock()

	topics, ok := w.connWithTopics[c]
	if ok {
		if _, ok = topics[topic]; ok {
			delete(w.subs, topic)
		}
	}
}

func (w *WebsocketManager) GetSlashSubscribeInfo() []Subscriber {
	paramsWithConns := w.subs[Slash]
	res := make([]Subscriber, 0, len(paramsWithConns))
	for _, conns := range paramsWithConns {
		for con := range conns {
			res = append(res, con)
		}
	}

	return res
}
func (w *WebsocketManager) GetHeightSubscribeInfo() []Subscriber {
	return nil
}
func (w *WebsocketManager) GetTickerSubscribeInfo() []Subscriber {
	return nil
}
func (w *WebsocketManager) GetCandleStickSubscribeInfo() map[string][]Subscriber {
	return nil
}
func (w *WebsocketManager) GetDepthSubscribeInfo() map[string][]Subscriber {
	return nil
}
func (w *WebsocketManager) GetDealSubscribeInfo() map[string][]Subscriber {
	return nil
}
func (w *WebsocketManager) GetBancorInfoSubscribeInfo() map[string][]Subscriber {
	return nil
}
func (w *WebsocketManager) GetCommentSubscribeInfo() map[string][]Subscriber {
	return nil
}
func (w *WebsocketManager) GetOrderSubscribeInfo() map[string][]Subscriber {

	return nil
}
func (w *WebsocketManager) GetBancorTradeSubscribeInfo() map[string][]Subscriber {
	return nil
}
func (w *WebsocketManager) GetIncomeSubscribeInfo() map[string][]Subscriber {
	return nil
}
func (w *WebsocketManager) GetUnbondingSubscribeInfo() map[string][]Subscriber {
	return nil
}
func (w *WebsocketManager) GetRedelegationSubscribeInfo() map[string][]Subscriber {
	return nil
}
func (w *WebsocketManager) GetUnlockSubscribeInfo() map[string][]Subscriber {
	return nil
}
func (w *WebsocketManager) GetTxSubscribeInfo() map[string][]Subscriber {
	return nil
}

// Push msgs----------------------------

func (w *WebsocketManager) PushSlash(subscriber Subscriber, info []byte) {

}
func (w *WebsocketManager) PushHeight(subscriber Subscriber, info []byte) {

}
func (w *WebsocketManager) PushTicker(subscriber Subscriber, t []*Ticker) {

}
func (w *WebsocketManager) PushDepthSell(subscriber Subscriber, delta map[*PricePoint]bool) {

}
func (w *WebsocketManager) PushDepthBuy(subscriber Subscriber, delta map[*PricePoint]bool) {

}
func (w *WebsocketManager) PushCandleStick(subscriber Subscriber, cs *CandleStick) {

}
func (w *WebsocketManager) PushDeal(subscriber Subscriber, info []byte) {

}
func (w *WebsocketManager) PushCreateOrder(subscriber Subscriber, info []byte) {

}
func (w *WebsocketManager) PushFillOrder(subscriber Subscriber, info []byte) {

}
func (w *WebsocketManager) PushCancelOrder(subscriber Subscriber, info []byte) {

}
func (w *WebsocketManager) PushBancorInfo(subscriber Subscriber, info []byte) {

}
func (w *WebsocketManager) PushBancorTrade(subscriber Subscriber, info []byte) {

}
func (w *WebsocketManager) PushIncome(subscriber Subscriber, info []byte) {

}
func (w *WebsocketManager) PushUnbonding(subscriber Subscriber, info []byte) {

}
func (w *WebsocketManager) PushRedelegation(subscriber Subscriber, info []byte) {

}
func (w *WebsocketManager) PushUnlock(subscriber Subscriber, info []byte) {

}
func (w *WebsocketManager) PushTx(subscriber Subscriber, info []byte) {

}
func (w *WebsocketManager) PushComment(subscriber Subscriber, info []byte) {

}
