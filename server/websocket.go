package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	"github.com/coinexchain/trade-server/core"
)

const (
	Subscribe   = "subscribe"
	Unsubscribe = "unsubscribe"
	Ping        = "ping"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type OpCommand struct {
	Op    string   `json:"op"`
	Args  []string `json:"args"`
	Depth int      `json:"depth"`
}

func ServeWsHandleFn(wsManager *core.WebsocketManager, hub *core.Hub) http.HandlerFunc {
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
					log.WithError(err).Error("read message failed")
					err = wsManager.CloseConn(wsConn)
					if err != nil {
						log.WithError(err).Error("close websocket failed")
					}
					break
				}

				var command OpCommand
				if err := json.Unmarshal(message, &command); err != nil {
					log.WithError(err).Error("unmarshal message failed: ", string(message))
					continue
				}
				switch command.Op {
				case Subscribe:
					for _, subTopic := range command.Args {
						err = wsManager.AddSubscribeConn(subTopic, command.Depth, wsConn, hub)
						if err != nil {
							log.WithError(err).Error(fmt.Sprintf("Subscribe topic (%s) failed ", subTopic))
							break
						}
					}
				case Unsubscribe:
					for _, subTopic := range command.Args {
						err = wsManager.RemoveSubscribeConn(subTopic, wsConn)
						if err != nil {
							log.WithError(err).Error(fmt.Sprintf("Unsubscribe topic (%s) failed ", subTopic))
							break
						}
					}
				case Ping:
					if err = wsConn.PingHandler()(`{"type":"pong"}`); err != nil {
						log.WithError(err).Error(fmt.Sprintf("pong message failed"))
					}
				default:
					log.Errorf("Unknown operation : %v", command.Op)
				}

				if err != nil {
					switch err.(type) {
					case *websocket.CloseError:
						log.Error(err)
						err = wsManager.CloseConn(wsConn)
						if err != nil {
							log.WithError(err).Error(fmt.Sprintf("Connection closed failed"))
						}
					default:
						err = wsConn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
					}
				}
			}
		}()
	}
}
