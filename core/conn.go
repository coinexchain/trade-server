package core

import (
	"strings"
	"sync"

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

type Conn struct {
	Conn            WsInterface
	mtx             sync.RWMutex
	allTopics       map[string]struct{}            // all the topics (with or without params)
	topicWithParams map[string]map[string]struct{} // topic --> params

	lastError error
	msgChan   chan []byte
	close     bool
}

func NewConn(c WsInterface) *Conn {
	conn := &Conn{
		Conn:            c,
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

func (c *Conn) sendMsg() {
	for {
		msg, ok := <-c.msgChan
		if !ok {
			break
		}
		c.lastError = c.Conn.WriteMessage(websocket.TextMessage, msg)
	}
}

func (c *Conn) ReadMsg(manager *WebsocketManager) (message []byte, err error) {
	if _, message, err = c.Conn.ReadMessage(); err != nil {
		log.WithError(err).Error("read message failed")
		manager.CloseConn(c)
	}
	return
}

func (c *Conn) WriteMsg(v []byte) error {
	if c.lastError != nil {
		return c.lastError
	}
	if c.close {
		return nil
	}
	c.msgChan <- v
	return nil
}

func (c *Conn) PingHandler() func(string) error {
	return c.Conn.PingHandler()
}

func (c *Conn) addTopicAndParams(topic string, params []string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
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
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if len(params) != 0 {
		if len(params) == 1 {
			delete(c.topicWithParams[topic], params[0])
		} else {
			tmpVal := strings.Join(params, SeparateArgu)
			delete(c.topicWithParams[topic], tmpVal)
		}
	}
	if c.isCleanedTopic(topic) {
		delete(c.allTopics, topic)
		delete(c.topicWithParams, topic)
	}
}

func (c *Conn) isCleanedTopic(topic string) bool {
	return len(c.topicWithParams[topic]) == 0
}

func (c *Conn) Close() error {
	c.close = true
	close(c.msgChan)
	return c.Conn.Close()
}
