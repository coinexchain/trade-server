package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8000", "http service address")

func main() {
	flag.Parse()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	fmt.Println(u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial: ", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	op := OpCommand{
		Op: "subscribe",
		Args: []string{
			"blockinfo",
			"slash",
			"depth:sdu1/cet",
			// "ticker:sdu1/cet",
			// "deal:sdu1/cet",
			"comment:cet",
			// "order:coinex16ur229a4xkj9e0xu06nqge9c23y70g7sl5vj98",
			// "kline:sdu1/cet:16",
			// "txs:coinex16ur229a4xkj9e0xu06nqge9c23y70g7sl5vj98",
			// "income:coinex16ur229a4xkj9e0xu06nqge9c23y70g7sl5vj98",
		},
	}

	bz, err := json.Marshal(op)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = c.WriteMessage(websocket.TextMessage, bz)
	if err != nil {
		log.Fatal(err)
		return
	}
	<-done

}

type OpCommand struct {
	Op   string
	Args []string
}
