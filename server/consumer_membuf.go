package server

import (
	"sync"

	"github.com/coinexchain/trade-server/core"
	toml "github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
)



// func (dt *DirTail) Start(interval int64, consumeFunc func(line string, fileNum uint32, offset uint32)) {

	//	app.msgQueProducer.SendMsg([]byte("commit"), []byte("{}"))


type kvpair struct {
	key []byte
	value []byte
}

type TradeConsumerWithMemBuf struct {
	bufPair    [2][]kvpair
	mtx        sync.Mutex
	bufIdx     int
	hub        *core.Hub
	writer     MsgWriter
}

func NewConsumerWithMemBuf(svrConfig *toml.Tree, hub *core.Hub) (*TradeConsumerWithMemBuf, error) {
	var (
		err     error
		writer  MsgWriter
	)
	if writer, err = initBackupWriter(svrConfig); err != nil {
		log.WithError(err).Errorf("init backup writer failed")
		return nil, err
	}
	return &TradeConsumerWithMemBuf{
		hub:        hub,
		writer:     writer,
	}, nil
}

func (tc *TradeConsumerWithMemBuf) String() string {
	return "membuf-consumer"
}

func (tc *TradeConsumerWithMemBuf) switchBuf() {
	if tc.bufIdx == 0 {
		tc.bufIdx = 1
	} else {
		tc.bufIdx = 0
	}
}

func (tc *TradeConsumerWithMemBuf) PutMsg(k, v []byte) {
	tc.bufPair[tc.bufIdx] = append(tc.bufPair[tc.bufIdx], kvpair{key: k, value: v})
	if string(k) == "commit" {
		finishedList := tc.bufPair[tc.bufIdx]
		tc.switchBuf()
		tc.mtx.Lock()
		go tc.consume(finishedList)
	}
}

func (tc *TradeConsumerWithMemBuf) consume(kvList []kvpair) {
	for _, kv := range kvList {
		tc.hub.ConsumeMessage(string(kv.key), kv.value)
		if tc.writer != nil {
			if err := tc.writer.WriteKV(kv.key, kv.value); err != nil {
				log.WithError(err).Error("write file failed")
			}
		}
		log.WithFields(log.Fields{"key": kv.key, "value": string(kv.value)}).Debug("consume message")
	}
	tc.mtx.Unlock()
}

