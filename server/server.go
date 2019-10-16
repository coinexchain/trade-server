package server

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coinexchain/trade-server/core"
	"github.com/pelletier/go-toml"
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
	consumer *TradeConsumer
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

	interval := svrConfig.GetDefault("interval", 60).(int)
	hub := core.NewHub(db, wsManager, int64(interval))
	restoreHub(&hub)

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

	var filePath string
	if backupToggle := svrConfig.GetDefault("backup-toggle", false).(bool); backupToggle {
		filePath = svrConfig.GetDefault("backup-file", "").(string)
		if len(filePath) == 0 {
			log.Fatal("backup data filePath is empty")
		}
	}
	consumer := NewConsumer(strings.Split(addrs, ","), DexTopic, filePath, &hub)
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
		// TODO: Compatible with old data. Delete later.
		// try to restore from file
		dumpFileName := dataDir + "/" + DumpFile
		if _, err := os.Stat(dumpFileName); err != nil {
			log.WithError(err).Infof("stat file fail: %s", dumpFileName)
			return
		}
		bz, err := ioutil.ReadFile(dumpFileName)
		if err != nil {
			log.WithError(err).Errorf("read from file fail %s", dumpFileName)
			return
		}
		data = bz
		if len(bz) > 16 && bz[0] != '{' {
			data = bz[16:]
			dataMD5 := md5.Sum(data)
			if !bytes.Equal(bz[:16], dataMD5[:]) {
				log.Error("hub data file is broken")
				return
			}
		}
	}
	hub4jo := &core.HubForJSON{}
	if err := json.Unmarshal(data, hub4jo); err != nil {
		log.WithError(err).Error("hub json unmarshal fail")
		return
	}
	hub.Load(hub4jo)
	log.Info("restore hub finish")
}
