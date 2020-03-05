package server

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/coinexchain/trade-server/core"
	toml "github.com/pelletier/go-toml"
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

func NewKafkaConsumer(svrConfig *toml.Tree, topic string, hub *core.Hub) (*TradeConsumer, error) {
	var (
		writer   MsgWriter
		err      error
		consumer sarama.Consumer
	)
	if writer, err = initBackupWriter(svrConfig); err != nil {
		return nil, err
	}
	if consumer, err = newKafka(svrConfig); err != nil {
		return nil, err
	}
	return &TradeConsumer{
		Consumer: consumer,
		topic:    topic,
		stopChan: make(chan byte, 1),
		quitChan: make(chan byte, 1),
		hub:      hub,
		writer:   writer,
	}, nil
}

func newKafka(svrConfig *toml.Tree) (sarama.Consumer, error) {
	addrs := svrConfig.GetDefault("kafka-addrs", "").(string)
	if len(addrs) == 0 {
		log.Error("kafka address is empty")
		return nil, fmt.Errorf("kafka address is empty")
	}
	sarama.Logger = log.StandardLogger()
	consumer, err := sarama.NewConsumer(strings.Split(addrs, ","), nil)
	if err != nil {
		log.WithError(err).Error("create consumer error")
		return nil, err
	}
	return consumer, nil

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

func (tc *TradeConsumer) String() string {
	return "kafka-consumer"
}
