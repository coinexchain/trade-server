package server

import (
	toml "github.com/pelletier/go-toml"

	"github.com/coinexchain/trade-server/core"
)

const FilePrefix = "backup-"

type Consumer interface {
	Consume()
	Close()
	String() string
}

func NewConsumer(svrConfig *toml.Tree, hub *core.Hub) Consumer {
	var (
		dir      bool
		consumer Consumer
	)
	dir = svrConfig.GetDefault("dir-mode", false).(bool)
	if dir {
		consumer = NewConsumerWithDirTail(svrConfig, hub)
	} else {
		consumer = NewKafkaConsumer(svrConfig, DexTopic, hub)
	}
	return consumer
}
