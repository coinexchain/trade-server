package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pelletier/go-toml"
)

const (
	ReadTimeout  = 10
	WriteTimeout = 10
	WaitTimeout  = 10
)

type TradeSever struct {
	http.Server
}

func NewServer(cfgFile string) *TradeSever {
	svrConfig, err := loadConfigFile(cfgFile)
	if err != nil {
		log.Printf("load config file fail:%v\n", err)
	}

	router := registerHandler()

	return &TradeSever{
		http.Server{
			Handler:      router,
			Addr:         fmt.Sprintf(":%d", svrConfig.GetDefault("port", 8000).(int)),
			WriteTimeout: WriteTimeout * time.Second,
			ReadTimeout:  ReadTimeout * time.Second,
		},
	}
}

func (ts *TradeSever) Start() {
	log.Printf("Server start... (%v)\n", ts.Addr)

	if err := ts.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("error occur:%v\n", err)
	}
}

func (ts *TradeSever) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), WaitTimeout*time.Second)
	defer cancel()

	if err := ts.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown failed. error:%v\n", err)
	}

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
