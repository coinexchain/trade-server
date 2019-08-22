package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	dbm "github.com/tendermint/tm-db"
	"github.com/coinexchain/trade-server/core"
)

func toStr(payload [][]byte) string {
	out := make([]string, len(payload))
	for i := 0; i < len(out); i++ {
		out[i] = string(payload[i])
	}
	return strings.Join(out, "\n")
}

func T(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func main() {
	file, err := os.Open("/Users/a/GoNew/src/github.com/coinexchain/trade-server/docs/dex_msgs_data.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	db := dbm.NewMemDB()
	subMan := core.GetSubscribeManager("coinex1x6rhu5m53fw8qgpwuljauaptvxyur57zym4jly","coinex1yj66ancalgk7dz3383s6cyvdd0nd93q0tk4x0c")
	hub := core.NewHub(db, subMan)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		divIdx := strings.Index(line, "#")
		msgType := line[:divIdx]
		msg := line[divIdx+1:]
		//fmt.Printf("%s %s\n", msgType, msg)
		hub.ConsumeMessage(msgType, []byte(msg))
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	unixTime := T("2019-08-21T08:02:06.647266Z").Unix()
	data := hub.QueryCandleStick("abc/cet", core.Minute, unixTime, 0, 20)
	fmt.Printf("here %s %d\n", toStr(data), unixTime)
}
