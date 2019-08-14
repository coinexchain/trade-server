package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/coinexchain/trade-server/server"
)

var (
	help    bool
	cfgFile string
)

func init() {
	flag.BoolVar(&help, "h", false, "display this help")
	flag.StringVar(&cfgFile, "c", "config.toml", "config file")
	flag.Usage = usage

	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	svr := server.NewServer(cfgFile)
	go svr.Start()

	waitForSignal()

	svr.Stop()
}

func waitForSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func usage() {
	_, _ = fmt.Println("Options:")
	flag.PrintDefaults()
}
