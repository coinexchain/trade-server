package server

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"

	"github.com/coinexchain/trade-server/core"
	log "github.com/sirupsen/logrus"
)

const (
	Subscribe   = "subscribe"
	Unsubscribe = "unsubscribe"
	Ping        = "ping"
)

type OpCommand struct {
	Op    string   `json:"op"`
	Args  []string `json:"args"`
	Depth int      `json:"depth"`
}

func NewCommand(msg []byte) *OpCommand {
	var command OpCommand
	if err := json.Unmarshal(msg, &command); err != nil {
		log.WithError(err).Error("json unmarshal message failed: ", string(msg))
		return nil
	}
	return &command
}

func (op *OpCommand) HandleCommand(hub *core.Hub, wsManager *core.WebsocketManager, wsConn *core.Conn) bool {
	var err error
	switch op.Op {
	case Subscribe:
		err = op.handleSubscribe(hub, wsManager, wsConn)
	case Unsubscribe:
		err = op.handleUnSubscribe(wsManager, wsConn)
	case Ping:
		err = op.handlePing(wsConn)
	default:
		op.defaultHandle()
	}
	return op.handlerExecOpErr(err, wsManager, wsConn)
}

func (op *OpCommand) handleSubscribe(hub *core.Hub, wsManager *core.WebsocketManager, wsConn *core.Conn) (err error) {
	for _, subTopic := range op.Args {
		topic, params, err := core.GetTopicAndParams(subTopic)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Parse subscribe topic (%s) failed ", subTopic))
			return err
		}
		if err = wsManager.PushFullInfo(hub, wsConn, topic, params, op.Depth); err != nil {
			log.WithError(err).Error(fmt.Sprintf("Push full info failed; topic (%s)", subTopic))
			return err
		}
		if err = wsManager.AddSubscribeConn(wsConn, topic, params); err != nil {
			log.WithError(err).Error(fmt.Sprintf("Add subscribe conn to wsManager failed; topic (%s)  ", subTopic))
			return err
		}
	}
	return nil
}

func (op *OpCommand) handleUnSubscribe(wsManager *core.WebsocketManager, wsConn *core.Conn) error {
	for _, subTopic := range op.Args {
		topic, params, err := core.GetTopicAndParams(subTopic)
		if err != nil {
			log.Error(err)
			return err
		}
		if err = wsManager.RemoveSubscribeConn(wsConn, topic, params); err != nil {
			log.WithError(err).Error(fmt.Sprintf("Unsubscribe topic (%s) failed ", subTopic))
			return err
		}
	}
	return nil
}

func (op *OpCommand) handlePing(wsConn *core.Conn) (err error) {
	if err = wsConn.PingHandler()(`{"type":"pong"}`); err != nil {
		log.WithError(err).Error(fmt.Sprintf("pong message failed"))
		return
	}
	return nil
}

func (op *OpCommand) defaultHandle() {
	log.Errorf("Unknown operation : %v", op.Op)
}

func (op *OpCommand) handlerExecOpErr(err error, wsManager *core.WebsocketManager, wsConn *core.Conn) bool {
	if err != nil {
		if _, ok := err.(*websocket.CloseError); ok {
			log.WithError(err).Errorf("handle command[%s] failed", op.Op)
			wsManager.CloseWsConn(wsConn)
			return false
		}
		if err = wsConn.WriteMsg([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error()))); err != nil {
			wsManager.CloseWsConn(wsConn)
			return false
		}
	}
	return true
}
