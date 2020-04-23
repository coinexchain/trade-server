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
	toml "github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
)

var (
	newFlag *flag.FlagSet
	help    bool
	cfgFile string
	version bool
)

var ReleaseVersion = "v0.2.3-alpha"

func init() {
	newFlag = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	newFlag.BoolVar(&help, "h", false, "display this help")
	newFlag.StringVar(&cfgFile, "c", "config.toml", "config file")
	newFlag.BoolVar(&version, "v", false, "show the trade-server version")
	newFlag.Usage = usage
}

func main() {
	if !isBeginService() {
		return
	}
	if svrConfig := initConfigAndLog(); svrConfig != nil {
		startService(svrConfig)
	}
	log.Info("trade-server exit")
}

func isBeginService() bool {
	if err := newFlag.Parse(os.Args[1:]); err != nil {
		return false
	}
	if help {
		newFlag.Usage()
		return false
	}
	if version {
		fmt.Println(ReleaseVersion)
		return false
	}
	return true
}

func initConfigAndLog() *toml.Tree {
	svrConfig, err := loadConfigFile(cfgFile)
	if err != nil {
		fmt.Printf("Load config file fail:%v\n", err)
		return nil
	}
	if err = utils.InitLog(svrConfig); err != nil {
		fmt.Printf("Init log fail:%v\n", err)
		os.Exit(1)
	}
	return svrConfig
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

func startService(svrConfig *toml.Tree) {
	if svr := server.NewServer(svrConfig, nil); svr != nil {
		svr.Start(svrConfig)
		waitForSignal()
		svr.Stop()
	}
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
