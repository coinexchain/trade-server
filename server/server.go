package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/coinexchain/trade-server/core"
	toml "github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
	dbm "github.com/tendermint/tm-db"
)

const (
	ReadTimeout  = 10
	WriteTimeout = 10
	WaitTimeout  = 10

	DexTopic = "coinex-dex"
	DbName   = "dex-trade"
	DumpFile = "hub"
)

var (
	dataDir        string
	certDir        string
	serverCertName = "server.crt"
	serverKeyName  = "server.key"
	httpsToggle    bool
)

type TradeSever struct {
	httpSvr  *http.Server
	hub      *core.Hub
	consumer Consumer
}

func NewServer(svrConfig *toml.Tree) *TradeSever {
	// websocket manager
	wsManager := core.NewWebSocketManager()

	// hub
	dataDir = svrConfig.GetDefault("data-dir", "data").(string)
	db, err := newLevelDB(DbName, dataDir)
	if err != nil {
		log.WithError(err).Fatal("open db fail")
	}

	interval := svrConfig.GetDefault("interval", int64(60)).(int64)
	monitorInterval := svrConfig.GetDefault("monitorinterval", int64(0)).(int64)
	hub := core.NewHub(db, wsManager, interval, monitorInterval)
	restoreHub(hub)

	//https toggle
	httpsToggle = svrConfig.GetDefault("https-toggle", false).(bool)
	if httpsToggle {
		certDir = svrConfig.GetDefault("cert-dir", "cert").(string)
		if _, err := os.Stat(certDir); err != nil && os.IsNotExist(err) {
			if err = os.Mkdir(certDir, 0755); err != nil {
				fmt.Print(err)
				log.WithError(err).Fatal("certificate path not exist")
			}
		}
	}

	// http server
	proxy := svrConfig.GetDefault("proxy", false).(bool)
	lcd := svrConfig.GetDefault("lcd", "").(string)
	router, err := registerHandler(hub, wsManager, proxy, lcd)
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
	consumer := NewConsumer(svrConfig, hub)
	return &TradeSever{
		httpSvr:  httpSvr,
		consumer: consumer,
		hub:      hub,
	}
}

func (ts *TradeSever) Start() {
	log.WithField("addr", ts.httpSvr.Addr).Info("Server start...")

	// start consumer
	go ts.consumer.Consume()

	// start http server
	go func() {
		if httpsToggle {
			certPath := certDir + "/" + serverCertName
			keyPath := certDir + "/" + serverKeyName
			if err := ts.httpSvr.ListenAndServeTLS(certPath, keyPath); err != nil && err != http.ErrServerClosed {
				log.WithError(err).Fatal("https server listen and serve error")
			}
		} else {
			if err := ts.httpSvr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.WithError(err).Fatal("http server listen and serve error")
			}
		}
	}()
}

func (ts *TradeSever) Stop() {
	// stop http server
	ctx, cancel := context.WithTimeout(context.Background(), WaitTimeout*time.Second)
	defer cancel()
	if err := ts.httpSvr.Shutdown(ctx); err != nil {
		log.WithError(err).Error("http server shutdown failed")
	}

	// stop hub (before closing consumer)
	ts.hub.Close()

	// stop consumer
	ts.consumer.Close()
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

func restoreHub(hub *core.Hub) {
	data := hub.LoadDumpData()
	if data == nil {
		log.Info("dump data does not exist, init new hub")
		return
	}
	hub4jo := &core.HubForJSON{}
	if err := json.Unmarshal(data, hub4jo); err != nil {
		log.WithError(err).Error("hub json unmarshal fail")
		return
	}
	hub.Load(hub4jo)
	log.Info("restore hub finish")
}
