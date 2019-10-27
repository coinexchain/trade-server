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

	var data interface{}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s\n", message)
			if err := json.Unmarshal(message, &data); err != nil {
				panic(err)
			}
		}
	}()

	op := OpCommand{
		Op: "subscribe",
		Args: []string{
			// "blockinfo",
			//"slash",
			//"send_lock_coins:coinex1rafnyd9j9gc9cwu5q5uflefpdn62awyl7rvh8t",
			//"depth:abc/cet:all",
			"bancor-deal:yofo/cet",
			"ticker:B:yofo/cet",
			"kline:B:yofo/cet:1min",
			// "depth:abc/cet:10",
			//"bancor-trade:cettest1svtmefnz5gxqzna9z5s670td27vehzchweqv83",
			//"unlock:coinex1tlegt4y40m3qu3dd4zddmjf6u3rswdqk8xxvzw",
			// "unlock:coinex1kc2nguz9xfttfpav4drldh2w96xyzrnqss9scw",
			// "ticker:abc/cet",
			// "deal:abc/cet",
			// "comment:cet",
			//"order:coinex1tlegt4y40m3qu3dd4zddmjf6u3rswdqk8xxvzw",
			// "kline:abc/cet:1min",
			//"txs:coinex18rdsh78t4ds76p58kum34rye2pmrt3hj8z2ehg",
			// "txs:coinex1avmxlmztzxap20hawpc85h3uzj3277ja88wec2",
			// "txs:coinex1j0awxx9lf32y235esjkwvs8h36whqj7f699f8z",
			// "txs:coinex1zyvvtlp2k2guuetqu3w06qxrr8w07f03s56kyj",
			// "income:coinex1x6rhu5m53fw8qgpwuljauaptvxyur57zym4jly",
			// "redelegation:coinex18rdsh78t4ds76p58kum34rye2pmrt3hj8z2ehg",
			// "unbonding:coinex1tlegt4y40m3qu3dd4zddmjf6u3rswdqk8xxvzw",
		},
		Depth: 30,
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
	Op    string   `json:"op"`
	Args  []string `json:"args"`
	Depth int      `json:"depth"`
}
