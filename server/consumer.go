package server

import (
	"sync"

	"github.com/Shopify/sarama"
	"github.com/coinexchain/trade-server/core"
	log "github.com/sirupsen/logrus"
)

type TradeConsumer struct {
	sarama.Consumer
	topic    string
	stopChan chan byte
	quitChan chan byte
	hub      *core.Hub
	writer   MsgWriter
}

func NewConsumer(addrs []string, topic string, writeConfig string, hub *core.Hub) *TradeConsumer {
	// set logger
	sarama.Logger = log.StandardLogger()

	consumer, err := sarama.NewConsumer(addrs, nil)
	if err != nil {
		log.WithError(err).Fatal("create consumer error")
	}
	var writer MsgWriter
	if len(writeConfig) != 0 {
		if writer, err = NewFileMsgWriter(writeConfig); err != nil {
			log.WithError(err).Fatal("create writer error")
		}
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
