package server

import (
	"strings"
	"sync"

	toml "github.com/pelletier/go-toml"

	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"

	"github.com/coinexchain/dirtail"
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

type TradeConsumer struct {
	sarama.Consumer
	topic    string
	stopChan chan byte
	quitChan chan byte
	hub      *core.Hub
	writer   MsgWriter
}

func (tc *TradeConsumer) String() string {
	return "kafka-consumer"
}

func NewKafkaConsumer(svrConfig *toml.Tree, topic string, hub *core.Hub) *TradeConsumer {
	var (
		writer   MsgWriter
		consumer sarama.Consumer
	)
	if writer = initBackupWriter(svrConfig); writer == nil {
		return nil
	}
	if consumer = newKafka(svrConfig); consumer == nil {
		return nil
	}
	return &TradeConsumer{
		Consumer: consumer,
		topic:    topic,
		stopChan: make(chan byte, 1),
		quitChan: make(chan byte, 1),
		hub:      hub,
		writer:   writer,
	}
}

func newKafka(svrConfig *toml.Tree) sarama.Consumer {
	addrs := svrConfig.GetDefault("kafka-addrs", "").(string)
	if len(addrs) == 0 {
		log.Error("kafka address is empty")
		return nil
	}
	sarama.Logger = log.StandardLogger()
	consumer, err := sarama.NewConsumer(strings.Split(addrs, ","), nil)
	if err != nil {
		log.WithError(err).Error("create consumer error")
		return nil
	}
	return consumer

}

func (tc *TradeConsumer) Consume() {
	defer close(tc.stopChan)
	partitionList, err := tc.Partitions(tc.topic)
	if err != nil {
		panic(err)
	}
	log.WithField("size", len(partitionList)).Info("consumer partitions")

	wg := &sync.WaitGroup{}
	for _, partition := range partitionList {
		offset := tc.hub.LoadOffset(partition)
		if offset == 0 {
			// start from the oldest offset
			offset = sarama.OffsetOldest
		} else {
			// start from next offset
			offset++
		}
		pc, err := tc.ConsumePartition(tc.topic, partition, offset)
		if err != nil {
			log.WithError(err).Errorf("Failed to start consumer for partition %d", partition)
			continue
		}
		wg.Add(1)

		go func(pc sarama.PartitionConsumer, partition int32) {
			log.WithFields(log.Fields{"partition": partition, "offset": offset}).Info("PartitionConsumer start")
			defer func() {
				pc.AsyncClose()
				wg.Done()
				log.WithFields(log.Fields{"partition": partition, "offset": offset}).Info("PartitionConsumer close")
			}()

			for {
				select {
				case msg := <-pc.Messages():
					// update offset, and then commit to db
					tc.hub.UpdateOffset(msg.Partition, msg.Offset)
					tc.hub.ConsumeMessage(string(msg.Key), msg.Value)
					offset = msg.Offset
					if tc.writer != nil {
						if err := tc.writer.WriteKV(msg.Key, msg.Value); err != nil {
							log.WithError(err).Error("write file failed")
						}
					}
					log.WithFields(log.Fields{"key": string(msg.Key), "value": string(msg.Value), "offset": offset}).Debug("consume message")
				case <-tc.quitChan:
					return
				}
			}
		}(pc, partition)
	}

	wg.Wait()
}

func (tc *TradeConsumer) Close() {
	close(tc.quitChan)
	<-tc.stopChan
	if err := tc.Consumer.Close(); err != nil {
		log.WithError(err).Error("consumer close failed")
	}
	if tc.writer != nil {
		if err := tc.writer.Close(); err != nil {
			log.WithError(err).Error("file close failed")
		}
	}

	log.Info("Consumer close")
}

// ======================================
type TradeConsumerWithDirTail struct {
	dirName    string
	filePrefix string
	dt         *dirtail.DirTail
	hub        *core.Hub
	writer     MsgWriter
}

func NewConsumerWithDirTail(svrConfig *toml.Tree, hub *core.Hub) Consumer {
	var (
		dataDir string
		writer  MsgWriter
	)
	dataDir = svrConfig.GetDefault("dir", "").(string)
	if writer = initBackupWriter(svrConfig); writer == nil {
		return nil
	}
	return &TradeConsumerWithDirTail{
		dirName:    dataDir,
		filePrefix: FilePrefix,
		hub:        hub,
		writer:     writer,
	}
}

func initBackupWriter(svrConfig *toml.Tree) MsgWriter {
	var (
		err          error
		writer       MsgWriter
		backFilePath string
	)
	if backupToggle := svrConfig.GetDefault("backup-toggle", false).(bool); backupToggle {
		if backFilePath = svrConfig.GetDefault("backup-file", "").(string); len(backFilePath) == 0 {
			log.Error("backup data filePath is empty")
			return nil
		}
	}
	if len(backFilePath) != 0 {
		if writer, err = NewFileMsgWriter(backFilePath); err != nil {
			log.WithError(err).Error("create writer error")
			return nil
		}
	}
	return writer

}

func (tc *TradeConsumerWithDirTail) String() string {
	return "dir-consumer"
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
