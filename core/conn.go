package core

import (
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Conn struct {
	*websocket.Conn
	topicWithParams map[string]map[string]struct{} // topic --> params
	allTopics       map[string]struct{}            // all the topics (with or without params)

	lastError error
	mtx       sync.RWMutex
}

func NewConn(c *websocket.Conn) *Conn {
	conn := &Conn{
		Conn:            c,
		allTopics:       make(map[string]struct{}),
		topicWithParams: make(map[string]map[string]struct{}),
	}
	c.SetPingHandler(func(appData string) error {
		return conn.WriteMsg([]byte(appData))
	})
	return conn
}

func (c *Conn) WriteMsg(v []byte) error {
	c.mtx.Lock()
	if c.lastError != nil {
		return c.lastError
	}
	go func() {
		c.lastError = c.WriteMessage(websocket.TextMessage, v)
		c.mtx.Unlock()
	}()
	return nil
}

func (c *Conn) ReadMsg(manager *WebsocketManager) (message []byte, err error) {
	if _, message, err = c.ReadMessage(); err != nil {
		log.WithError(err).Error("read message failed")
		manager.CloseConn(c)
	}
	return
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
