package database

import (
	"context"
	"fmt"
	"time"

	"github.com/labbsr0x/kafka2influxdb/web/config"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

// Kafka defines the methods that can be performed
type Kafka interface {
	NewKafka(webBuilder *config.WebBuilder) *DefaultKafka
	Connect() *DefaultKafka
	Close() error
	Consumer()
	Produce(message string)
	Reader(handler func(string)) error
}

// DefaultKafka a default Kafka interface implementation
type DefaultKafka struct {
	Addr      string
	Topic     string
	Partition int
	Messages  []string
	Client    *kafka.Conn
}

// NewKafka initializes a default configs from web builder
func NewKafka(webBuilder *config.WebBuilder) *DefaultKafka {
	instance := new(DefaultKafka)
	instance.Addr = webBuilder.KafkaAddr
	instance.Topic = webBuilder.KafkaTopic
	instance.Partition = webBuilder.KafkaPartition

	return instance
}

// Connect to Kafka
func (dk *DefaultKafka) Connect() *DefaultKafka {
	logrus.Debugf("Connecting to Kafka. Host: %s, Topic: %s, Partition: %s", dk.Addr, dk.Topic, dk.Partition)

	conn, err := kafka.DialLeader(context.Background(), "tcp", dk.Addr, dk.Topic, dk.Partition)
	if err != nil {
		logrus.Errorf("error connecting to Kafka: %s", err)
		panic(fmt.Sprintf("No error should happen when connecting to Kafka, but got err=%+v", err))
	}

	dk.Client = conn
	logrus.Debugf("connected to Kafka")

	return dk
}

// Close all opened connections
func (dk *DefaultKafka) Close() error {
	return dk.Client.Close()
}

// Consumer a Kafka topic
func (dk *DefaultKafka) Consumer() {
	dk.Messages = []string{}
	dk.Client.SetReadDeadline(time.Now().Add(10 * time.Second))
	batch := dk.Client.ReadBatch(10e3, 1e6) // fetch 10KB min, 1MB max

	b := make([]byte, 10e3) // 10KB max per message
	for {
		_, err := batch.Read(b)
		if err != nil {
			break
		}
		fmt.Println(string(b))
		dk.Messages = append(dk.Messages, string(b))
	}

	batch.Close()
	dk.Client.Close()
}

// Produce a message in a Kafka topic
func (dk *DefaultKafka) Produce(message string) {
	dk.Client.SetWriteDeadline(time.Now().Add(10 * time.Second))
	dk.Client.WriteMessages(kafka.Message{Value: []byte(message)})
	dk.Client.Close()
}

// Listen messages from Kafka topic
func (dk *DefaultKafka) Listen(handler func(string)) error {
	fmt.Sprintf("\n\tCreating Kafka reader with params\n. Host: %s, Topic: %s, Partition: %s", dk.Addr, dk.Topic, dk.Partition)
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{dk.Addr},
		Topic:     dk.Topic,
		Partition: dk.Partition,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			return err
			break
		}
		fmt.Printf("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
		handler(string(m.Value))
	}

	r.Close()

	return nil
}
