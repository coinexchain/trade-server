package server

import (
	"net/http"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	"github.com/coinexchain/trade-server/core"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ServeWsHandleFn(wsManager *core.WebsocketManager, hub *core.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.WithError(err).Errorf("upgrader http request to websocket failed")
			return
		}
		wsConn := wsManager.AddWsConn(c)

		go handleConnIncomingMsg(hub, wsManager, wsConn)
	}
}

func handleConnIncomingMsg(hub *core.Hub, wsManager *core.WebsocketManager, wsConn *core.Conn) {
	for {
		var (
			err     error
			message []byte
			command *OpCommand
		)
		if message, err = wsConn.ReadMsg(wsManager); err != nil {
			break
		}
		if command = NewCommand(message); command == nil {
			continue
		}
		if !command.HandleCommand(hub, wsManager, wsConn) {
			break
		}
	}
}
