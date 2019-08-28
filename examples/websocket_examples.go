// +build examples

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8000", "http service address")

func main() {
	flag.Parse()
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
			// "blockinfo",
			"slash",
			"send_lock_coins:coinex1rafnyd9j9gc9cwu5q5uflefpdn62awyl7rvh8t",
			// "depth:abc/cet",
			// "bancor-trade:coinex1x6rhu5m53fw8qgpwuljauaptvxyur57zym4jly",
			// "unlock:coinex1tlegt4y40m3qu3dd4zddmjf6u3rswdqk8xxvzw",
			// "unlock:coinex1kc2nguz9xfttfpav4drldh2w96xyzrnqss9scw",
			// "ticker:abc/cet",
			// "deal:abc/cet",
			// "comment:cet",
			// "order:coinex1x6rhu5m53fw8qgpwuljauaptvxyur57zym4jly",
			// "kline:abc/cet:16",
			// "txs:coinex1x6rhu5m53fw8qgpwuljauaptvxyur57zym4jly",
			// "income:coinex1x6rhu5m53fw8qgpwuljauaptvxyur57zym4jly",
			// "redelegation:coinex18rdsh78t4ds76p58kum34rye2pmrt3hj8z2ehg",
			// "unbonding:coinex1tlegt4y40m3qu3dd4zddmjf6u3rswdqk8xxvzw",
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
