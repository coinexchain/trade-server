package server

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/Shopify/sarama"
)

type TradeConsumer struct {
	sarama.Consumer
	topic    string
	stopChan chan byte
}

func init() {
	sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
}

func NewConsumer(addrs []string, topic string) (*TradeConsumer, error) {
	consumer, err := sarama.NewConsumer(addrs, nil)
	if err != nil {
		panic(err)
	}

	return &TradeConsumer{
		Consumer: consumer,
		topic:    topic,
		stopChan: make(chan byte, 1),
	}, nil
}

func (tc *TradeConsumer) Consume() {
	defer close(tc.stopChan)
	partitionList, err := tc.Partitions(tc.topic)
	if err != nil {
		panic(err)
	}

	wg := &sync.WaitGroup{}
	for partition := range partitionList {
		pc, err := tc.ConsumePartition(tc.topic, int32(partition), sarama.OffsetNewest)
		if err != nil {
			log.Printf("Failed to start consumer for partition %d: %s\n", partition, err)
			continue
		}
		wg.Add(1)

		go func(sarama.PartitionConsumer) {
			log.Printf("PartitionConsumer %v start\n", partition)
			defer func() {
				pc.AsyncClose()
				wg.Done()
				log.Printf("PartitionConsumer %v close\n", partition)
			}()

			signals := make(chan os.Signal, 1)
			signal.Notify(signals, os.Interrupt)

			for {
				select {
				case msg := <-pc.Messages():
					fmt.Printf("Partition:%d, Offset:%d, Key:%s, Value:%s\n", msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
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
