package core

import (
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type WsInterface interface {
	Close() error
	WriteMessage(msgType int, data []byte) error
	ReadMessage() (messageType int, p []byte, err error)
	SetPingHandler(func(string) error)
	PingHandler() func(string) error
}

// Conn models the connection to a client through websocket
type Conn struct {
	WsIfc           WsInterface
	mtx             sync.RWMutex
	allTopics       map[string]struct{}            // all the topics (with or without params)
	topicWithParams map[string]map[string]struct{} // topic --> params

	lastError atomic.Value
	msgChan   chan []byte
	closed    bool
}

func NewConn(c WsInterface) *Conn {
	conn := &Conn{
		WsIfc:           c,
		msgChan:         make(chan []byte),
		allTopics:       make(map[string]struct{}),
		topicWithParams: make(map[string]map[string]struct{}),
	}
	c.SetPingHandler(func(appData string) error {
		return conn.WriteMsg([]byte(appData))
	})
	go conn.sendMsg()
	return conn
}

// Since WsIfc.WriteMessage is blocking, we'd like to wrap it into a goroutine
func (c *Conn) sendMsg() {
	for {
		msg, ok := <-c.msgChan
		if !ok {
			break
		}

		if c.lastError.Load() == nil {
			if err := c.WsIfc.WriteMessage(websocket.TextMessage, msg); err != nil {
				c.lastError.Store(err)
			}
		}
	}
}

// Since WsIfc.ReadMessage is blocking, this function is only used in the handleConn goroutine
func (c *Conn) ReadMsg(manager *WebsocketManager) (message []byte, err error) {
	if _, message, err = c.WsIfc.ReadMessage(); err != nil {
		log.WithError(err).Error("read message failed")
		manager.CloseWsConn(c)
	}
	return
}

func (c *Conn) WriteMsg(v []byte) error {
	if val := c.lastError.Load(); val != nil {
		return val.(error)
	}
	if c.closed { // Why? Should we return an error when writing to a closed Conn?
		return nil
	}
	c.msgChan <- v
	return nil
}

func (c *Conn) PingHandler() func(string) error {
	return c.WsIfc.PingHandler()
}

func (c *Conn) addTopicAndParams(topic string, params []string) {
	c.allTopics[topic] = struct{}{}
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
}

func (c *Conn) removeTopicAndParams(topic string, params []string) {
	if len(params) != 0 {
		if len(params) == 1 {
			delete(c.topicWithParams[topic], params[0])
		} else {
			tmpVal := strings.Join(params, SeparateArgu)
			delete(c.topicWithParams[topic], tmpVal) // Why? The order of the params is important?
		}
	}
	if c.topicHasEmptyParamSet(topic) {
		delete(c.allTopics, topic)
		delete(c.topicWithParams, topic)
	}
}

// if the param set of this topic is empty
func (c *Conn) topicHasEmptyParamSet(topic string) bool {
	return len(c.topicWithParams[topic]) == 0
}

func (c *Conn) Close() error {
	c.closed = true
	close(c.msgChan)
	return c.WsIfc.Close()
}
