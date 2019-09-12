package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/coinexchain/trade-server/server"
	"github.com/coinexchain/trade-server/utils"
	"github.com/pelletier/go-toml"
)

var (
	newFlag *flag.FlagSet
	help    bool
	cfgFile string
)

func init() {
	newFlag = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	newFlag.BoolVar(&help, "h", false, "display this help")
	newFlag.StringVar(&cfgFile, "c", "config.toml", "config file")
	newFlag.Usage = usage
}

func main() {
	_ = newFlag.Parse(os.Args[1:])
	if help {
		newFlag.Usage()
		return
	}

	svrConfig, err := loadConfigFile(cfgFile)
	if err != nil {
		fmt.Printf("Load config file fail:%v\n", err)
		os.Exit(1)
	}

	if err = utils.InitLog(svrConfig); err != nil {
		fmt.Printf("Init log fail:%v\n", err)
		os.Exit(1)
	}

	svr := server.NewServer(svrConfig)
	svr.Start()

	waitForSignal()

	svr.Stop()
}

func loadConfigFile(cfgFile string) (*toml.Tree, error) {
	if _, err := os.Stat(cfgFile); err != nil {
		return nil, err
	}

	bz, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return nil, err
	}

	tree, err := toml.LoadBytes(bz)
	if err != nil {
		return nil, err
	}

	return tree, nil
}

func waitForSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-c
}

func usage() {
	_, _ = fmt.Println("Options:")
	newFlag.PrintDefaults()
}
