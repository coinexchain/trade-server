package server

import (
	"github.com/coinexchain/trade-server/core"
	toml "github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
)

const FilePrefix = "backup-"

type Consumer interface {
	Consume()
	Close()
	String() string
}

func NewConsumer(svrConfig *toml.Tree, hub *core.Hub) (Consumer, error) {
	var (
		dir      bool
		err      error
		consumer Consumer
	)
	dir = svrConfig.GetDefault("dir-mode", false).(bool)
	if dir {
		if consumer, err = NewConsumerWithDirTail(svrConfig, hub); err != nil {
			log.WithError(err).Errorf("NewConsumerWithDirTail failed")
		}
	} else {
		if consumer, err = NewKafkaConsumer(svrConfig, DexTopic, hub); err != nil {
			log.WithError(err).Errorf("NewKafkaConsumer failed")
		}
	}
	return consumer, err
}
