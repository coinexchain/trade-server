package server

import (
	"context"
	"fmt"
	"github.com/coinexchain/trade-server/core"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
	dbm "github.com/tendermint/tm-db"
)

const (
	ReadTimeout  = 10
	WriteTimeout = 10
	WaitTimeout  = 10

	DexTopic = "coinex-dex"
	DbName   = "dex-trade"
)

type TradeSever struct {
	httpSvr  *http.Server
	hub      *core.Hub
	consumer *TradeConsumer
}

func NewServer(cfgFile string) *TradeSever {
	svrConfig, err := loadConfigFile(cfgFile)
	if err != nil {
		log.Printf("load config file fail:%v\n", err)
	}

	// hub
	dataDir := svrConfig.GetDefault("data-dir", "data").(string)
	db, err := newLevelDB(DbName, dataDir)
	if err != nil {
		log.Fatalf("open db fail. %v\n", err)
	}
	// TODO: SubscribeManager
	hub := core.NewHub(db, TestSubscribeManager{})

	// http server
	router := registerHandler()
	httpSvr := &http.Server{
		Addr:         fmt.Sprintf(":%d", svrConfig.GetDefault("port", 8000).(int64)),
		Handler:      router,
		ReadTimeout:  ReadTimeout * time.Second,
		WriteTimeout: WriteTimeout * time.Second,
	}

	// consumer
	addrs := svrConfig.Get("kafka-addrs").(string)
	if len(addrs) == 0 {
		log.Fatalln("kafka address is empty")
	}
	consumer, err := NewConsumer(strings.Split(addrs, ","), DexTopic, &hub)
	if err != nil {
		log.Fatalf("create consumer error:%v\n", err)
	}

	return &TradeSever{
		httpSvr:  httpSvr,
		consumer: consumer,
		hub:      &hub,
	}
}

func (ts *TradeSever) Start() {
	log.Printf("Server start... (%v)\n", ts.httpSvr.Addr)

	// start consumer
	go ts.consumer.Consume()

	// start http server
	go func() {
		if err := ts.httpSvr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("error occur:%v\n", err)
		}
	}()
}

func (ts *TradeSever) Stop() {
	// stop http server
	ctx, cancel := context.WithTimeout(context.Background(), WaitTimeout*time.Second)
	defer cancel()
	if err := ts.httpSvr.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown failed. error:%v\n", err)
	}

	// stop consumer
	ts.consumer.Close()

	ts.hub.Close()

	log.Println("Server stop...")
}

func loadConfigFile(cfgFile string) (*toml.Tree, error) {
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		log.Printf("%s does not exist\n", cfgFile)
		return toml.Load(``)
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

func newLevelDB(name string, dir string) (db dbm.DB, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("couldn't create db: %v", r)
		}
	}()
	return dbm.NewDB(name, dbm.GoLevelDBBackend, dir), err
}
