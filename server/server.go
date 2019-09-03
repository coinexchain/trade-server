package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coinexchain/trade-server/core"
	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
	dbm "github.com/tendermint/tm-db"
)

type askInfo struct {
	amount int
	price  int
}

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

func NewServer(svrConfig *toml.Tree) *TradeSever {
	// websocket manager
	wsManager := core.NewWebSocketManager()

	// hub
	dataDir := svrConfig.GetDefault("data-dir", "data").(string)
	db, err := newLevelDB(DbName, dataDir)
	if err != nil {
		log.WithError(err).Fatal("open db fail")
	}
	hub := core.NewHub(db, wsManager)

	// http server
	proxy := svrConfig.GetDefault("proxy", false).(bool)
	lcd := svrConfig.GetDefault("lcd", "").(string)
	router, err := registerHandler(&hub, wsManager, proxy, lcd)
	if err != nil {
		log.WithError(err).Fatal("registerHandler fail")
	}
	httpSvr := &http.Server{
		Addr:         fmt.Sprintf(":%d", svrConfig.GetDefault("port", 8000).(int64)),
		Handler:      router,
		ReadTimeout:  ReadTimeout * time.Second,
		WriteTimeout: WriteTimeout * time.Second,
	}

	// consumer
	addrs := svrConfig.GetDefault("kafka-addrs", "").(string)
	if len(addrs) == 0 {
		log.Fatal("kafka address is empty")
	}
	consumer, err := NewConsumer(strings.Split(addrs, ","), DexTopic, &hub)
	if err != nil {
		log.WithError(err).Fatal("create consumer error")
	}

	return &TradeSever{
		httpSvr:  httpSvr,
		consumer: consumer,
		hub:      &hub,
	}
}

func (ts *TradeSever) Start() {
	log.WithField("addr", ts.httpSvr.Addr).Info("Server start...")

	// start consumer
	go ts.consumer.Consume()

	// start http server
	go func() {
		if err := ts.httpSvr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("http server listen and serve error")
		}
	}()
}

func (ts *TradeSever) Stop() {
	// stop http server
	ctx, cancel := context.WithTimeout(context.Background(), WaitTimeout*time.Second)
	defer cancel()
	if err := ts.httpSvr.Shutdown(ctx); err != nil {
		log.WithError(err).Fatal("http server shutdown failed")
	}

	// stop consumer
	ts.consumer.Close()

	// stop hub
	ts.hub.Close()

	log.Info("Server stop...")
}

func newLevelDB(name string, dir string) (db dbm.DB, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("couldn't create db: %v", r)
		}
	}()
	return dbm.NewDB(name, dbm.GoLevelDBBackend, dir), err
}
