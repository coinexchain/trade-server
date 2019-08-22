package server

import (
	"encoding/json"
	"net/http"

	"github.com/coinexchain/trade-server/core"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	Subscribe   = "subscribe"
	Unsubscribe = "unsubscribe"
	Ping        = "ping"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type OpCommand struct {
	Op   string   `json:"op"`
	Args []string `json:"args"`
}

func ServeWsHandleFn(wsManager *core.WebsocketManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Error(err)
			return
		}
		wsConn := core.NewConn(c)
		wsManager.AddConn(wsConn)

		go func() {
			for {
				_, message, err := wsConn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.WithError(err).Error("unexpected close error")
					}
					// TODO. will handle the error
					_ = wsManager.CloseConn(wsConn)
					break
				}

				// TODO. will handler the error
				var command OpCommand
				if err := json.Unmarshal(message, &command); err != nil {
					log.WithError(err).Error("unmarshal fail")
					continue
				}

				// TODO. will handler the error
				switch command.Op {
				case Subscribe:
					for _, subTopic := range command.Args {
						_ = wsManager.AddSubscribeConn(subTopic, wsConn)
					}
				case Unsubscribe:
					for _, subTopic := range command.Args {
						_ = wsManager.RemoveSubscribeConn(subTopic, wsConn)
					}
				case Ping:
					_ = wsConn.PongHandler()(`{"type": "pong"}`)
				}

			}
		}()
	}
}
