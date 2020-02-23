package server

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"

	"github.com/coinexchain/trade-server/core"
	log "github.com/sirupsen/logrus"
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
		if err = wsManager.AddSubscribeConn(subTopic, op.Depth, wsConn, hub); err != nil {
			log.WithError(err).Error(fmt.Sprintf("Subscribe topic (%s) failed ", subTopic))
			return
		}
	}
	return nil
}

func (op *OpCommand) handleUnSubscribe(wsManager *core.WebsocketManager, wsConn *core.Conn) (err error) {
	for _, subTopic := range op.Args {
		if err = wsManager.RemoveSubscribeConn(subTopic, wsConn); err != nil {
			log.WithError(err).Error(fmt.Sprintf("Unsubscribe topic (%s) failed ", subTopic))
			return
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
			wsManager.CloseConn(wsConn)
			return false
		}
		if err = wsConn.WriteMsg([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error()))); err != nil {
			wsManager.CloseConn(wsConn)
			return false
		}
	}
	return true
}
