package server

import (
	"os"
	"os/signal"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/coinexchain/trade-server/core"
	log "github.com/sirupsen/logrus"
)

type TradeConsumer struct {
	sarama.Consumer
	topic    string
	stopChan chan byte
	hub      *core.Hub
}

func NewConsumer(addrs []string, topic string, hub *core.Hub) (*TradeConsumer, error) {
	// set logger
	sarama.Logger = log.StandardLogger()

	consumer, err := sarama.NewConsumer(addrs, nil)
	if err != nil {
		panic(err)
	}

	return &TradeConsumer{
		Consumer: consumer,
		topic:    topic,
		stopChan: make(chan byte, 1),
		hub:      hub,
	}, nil
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
			log.Errorf("Failed to start consumer for partition %d: %s", partition, err)
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

			signals := make(chan os.Signal, 1)
			signal.Notify(signals, os.Interrupt)

			for {
				select {
				case msg := <-pc.Messages():
					// update offset, and then commit to db
					tc.hub.UpdateOffset(msg.Partition, msg.Offset)
					tc.hub.ConsumeMessage(string(msg.Key), msg.Value)
					offset = msg.Offset
					log.WithFields(log.Fields{"key": string(msg.Key), "value": string(msg.Value)}).Debug("consume message")
				case <-signals:
					return
				}
			}
		}(pc, partition)
	}

	wg.Wait()
}

func (tc *TradeConsumer) Close() {
	<-tc.stopChan
	if err := tc.Consumer.Close(); err != nil {
		log.WithError(err).Fatal("consumer close")
	}
	log.Info("Consumer close")
}
