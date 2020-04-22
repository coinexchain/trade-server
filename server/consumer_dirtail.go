package server

import (
	"strings"

	"github.com/coinexchain/dirtail"
	"github.com/coinexchain/trade-server/core"
	toml "github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
)

type TradeConsumerWithDirTail struct {
	dirName    string
	filePrefix string
	dt         *dirtail.DirTail
	hub        *core.Hub
	writer     MsgWriter
}

func NewConsumerWithDirTail(svrConfig *toml.Tree, hub *core.Hub) (Consumer, error) {
	var (
		dataDir string
		err     error
		writer  MsgWriter
	)
	dataDir = svrConfig.GetDefault("dir", "").(string)
	if writer, err = initBackupWriter(svrConfig); err != nil {
		log.WithError(err).Errorf("init backup writer failed")
		return nil, err
	}
	return &TradeConsumerWithDirTail{
		dirName:    dataDir,
		filePrefix: FilePrefix,
		hub:        hub,
		writer:     writer,
	}, nil
}

func (tc *TradeConsumerWithDirTail) String() string {
	return "dir-consumer"
}

func (tc *TradeConsumerWithDirTail) GetDumpHeight() int64 {
	if dump, err := GetHubDumpData(tc.hub); err == nil && dump != nil {
		return dump.CurrBlockHeight
	}
	return 0
}

func (tc *TradeConsumerWithDirTail) Consume() {
	offset := tc.hub.LoadOffset(0)
	fileOffset := uint32(offset)
	fileNum := uint32(offset >> 32)
	tc.dt = dirtail.NewDirTail(tc.dirName, tc.filePrefix, "", fileNum, fileOffset)
	tc.dt.Start(600, func(line string, fileNum uint32, fileOffset uint32) {
		offset := (int64(fileNum) << 32) | int64(fileOffset)
		tc.hub.UpdateOffset(0, offset)

		divIdx := strings.Index(line, "#")
		key := line[:divIdx]
		value := []byte(line[divIdx+1:])
		tc.hub.ConsumeMessage(key, value)
		if tc.writer != nil {
			if err := tc.writer.WriteKV([]byte(key), value); err != nil {
				log.WithError(err).Error("write file failed")
			}
		}
		log.WithFields(log.Fields{"key": key, "value": string(value), "offset": offset}).Debug("consume message")
	})
}

func (tc *TradeConsumerWithDirTail) Close() {
	tc.dt.Stop()
	if tc.writer != nil {
		if err := tc.writer.Close(); err != nil {
			log.WithError(err).Error("file close failed")
		}
	}
	log.Info("Consumer close")
}
