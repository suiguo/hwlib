package kafka

import (
	"fmt"
	"sync"
	"time"

	"log"

	co_kafka "github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/suiguo/hwlib/logger"
)

// To restart kafka after an upgrade:
//
//	brew services restart kafka
//
// Or, if you don't want/need a background service you can just run:
//
//	/opt/homebrew/opt/kafka/bin/kafka-server-start /opt/homebrew/etc/kafka/server.properties
type KafkaType string
type ConumerOffset string

const (
	Smallest  ConumerOffset = "smallest"
	Earliest  ConumerOffset = "earliest"
	Beginning ConumerOffset = "beginning"
	Largest   ConumerOffset = "largest"
	Latest    ConumerOffset = "latest"
	End       ConumerOffset = "end"
	Error     ConumerOffset = "error"
)

// , earliest, beginning, largest, latest, end, error
const KafkaLogTag = "Kafka"
const (
	ALLType      KafkaType = "all"
	ProducerType KafkaType = "Produce"
	ConsumerType KafkaType = "Consumer"
)

type KafkaMsg struct {
	Partition int32
	Offset    int64
	Key       string
	Msg       []byte
	MetaData  string
}
type KafaClient interface {
	Producer
	Consumer
}
type Producer interface {
	Produce(topic string, msg *KafkaMsg) error
}
type Consumer interface {
	MessageChan() <-chan *co_kafka.Message
	Subscribe(topics ...string) error
}

type kafkaClient struct {
	sync.Once
	msgPopChan chan *co_kafka.Message
	logger.Logger
	producer *co_kafka.Producer
	consumer *co_kafka.Consumer
}

func (k *kafkaClient) Subscribe(topics ...string) error {
	if k.consumer == nil {
		return fmt.Errorf("consumer not int")
	}
	return k.consumer.SubscribeTopics(topics, nil)
}
func (k *kafkaClient) MessageChan() <-chan *co_kafka.Message {
	return k.msgPopChan
}
func (k *kafkaClient) Produce(topic string, msg *KafkaMsg) error {
	if k.producer == nil {
		return fmt.Errorf("producer not init")
	}
	topic_partition := co_kafka.TopicPartition{}
	if msg.MetaData != "" {
		topic_partition.Metadata = &msg.MetaData
	}
	if topic != "" {
		topic_partition.Topic = &topic
	}
	if msg.Offset > 0 {
		topic_partition.Offset.Set(msg.Offset)
	}
	tmp_msg := &co_kafka.Message{
		TopicPartition: topic_partition,
		// Key:            []byte(msg.Key),
		Value: msg.Msg,
	}
	if msg.Key != "" {
		tmp_msg.Key = []byte(msg.Key)
	}
	return k.producer.Produce(tmp_msg, nil)
}

// run
func (k *kafkaClient) run() {
	k.Once.Do(func() {
		for {
			if k.consumer == nil {
				time.Sleep(time.Second * 2)
			}
			msg, err := k.consumer.ReadMessage(time.Second)
			if err != nil || msg == nil {
				if err.(co_kafka.Error).Code() == co_kafka.ErrTimedOut {
					continue
				}
				if k.Logger != nil {
					k.Logger.Error(KafkaLogTag, "ReadMessage", err)
				} else {
					log.Println(KafkaLogTag, "ReadMessage", err)
				}
				continue
			}
			go func(data *co_kafka.Message) {
				k.msgPopChan <- data
			}(msg)
		}
	})
}
func GetKafkaByCfg(ktype KafkaType, consumer co_kafka.ConfigMap, producer co_kafka.ConfigMap, log logger.Logger) (KafaClient, error) {
	tmp := &kafkaClient{
		Logger:     log,
		msgPopChan: make(chan *co_kafka.Message, 1000),
	}
	var err error
	switch ktype {
	case ALLType:
		tmp.consumer, err = co_kafka.NewConsumer(&consumer)
		if err != nil {
			return nil, err
		}
		tmp.producer, err = co_kafka.NewProducer(&producer)
	case ConsumerType:
		tmp.consumer, err = co_kafka.NewConsumer(&consumer)
	case ProducerType:
		tmp.producer, err = co_kafka.NewProducer(&producer)
	}
	if err != nil {
		return nil, err
	}
	tmp.run()
	return tmp, err
}
func GetDefaultKafka(ktype KafkaType, server string, group_id string, offset ConumerOffset, log logger.Logger) (KafaClient, error) {
	tmp := &kafkaClient{
		Logger:     log,
		msgPopChan: make(chan *co_kafka.Message, 1000),
	}
	client_cfg := &co_kafka.ConfigMap{
		"bootstrap.servers": server,
		"group.id":          group_id,
	}
	if offset != "" {
		client_cfg.SetKey("auto.offset.reset", string(offset))
	} else {
		client_cfg.SetKey("auto.offset.reset", "earliest")
	}
	var err error
	switch ktype {
	case ALLType:
		tmp.consumer, err = co_kafka.NewConsumer(client_cfg)
		if err != nil {
			return nil, err
		}
		tmp.producer, err = co_kafka.NewProducer(&co_kafka.ConfigMap{
			"bootstrap.servers": server,
		})
	case ConsumerType:
		tmp.consumer, err = co_kafka.NewConsumer(client_cfg)
	case ProducerType:
		tmp.producer, err = co_kafka.NewProducer(&co_kafka.ConfigMap{
			"bootstrap.servers": server,
		})
	}
	if err != nil {
		return nil, err
	}
	go tmp.run()
	return tmp, err
}
