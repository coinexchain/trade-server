package server

import (
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/coinexchain/trade-server/core"
)

type TradeConsumer struct {
	sarama.Consumer
	topic    string
	stopChan chan byte
	hub      *core.Hub
}

func init() {
	sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
}

func NewConsumer(addrs []string, topic string, hub *core.Hub) (*TradeConsumer, error) {
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
	log.Printf("Partition size:%v\n", len(partitionList))

	wg := &sync.WaitGroup{}
	for _, partition := range partitionList {
		offset := tc.hub.LoadOffset(partition)
		if offset == 0 {
			// start from the oldest offset
			offset = sarama.OffsetOldest
		} else {
			// start from next offset
			offset += 1
		}
		pc, err := tc.ConsumePartition(tc.topic, partition, offset)
		if err != nil {
			log.Printf("Failed to start consumer for partition %d: %s\n", partition, err)
			continue
		}
		wg.Add(1)

		go func(sarama.PartitionConsumer) {
			log.Printf("PartitionConsumer %v start from %v\n", partition, offset)
			defer func() {
				pc.AsyncClose()
				wg.Done()
				log.Printf("PartitionConsumer %v close in %v\n", partition, offset)
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
				case <-signals:
					return
				}
			}
		}(pc)
	}

	wg.Wait()
}

func (tc *TradeConsumer) Close() {
	<-tc.stopChan
	if err := tc.Consumer.Close(); err != nil {
		log.Fatalln(err)
	}
	log.Println("Consumer close")
}
