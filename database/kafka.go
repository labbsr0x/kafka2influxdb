package database

import (
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/labbsr0x/kafka2influxdb/web/config"

	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

// Kafka defines the methods that can be performed
type Kafka interface {
	NewKafka(webBuilder *config.WebBuilder) *DefaultKafka
	Connect() *DefaultKafka
	Close() error
	ListenGroup(handler func(string))
	consume(handler func(string)) (chan *sarama.ConsumerMessage, chan *sarama.ConsumerError)
}

// DefaultKafka a default Kafka interface implementation
type DefaultKafka struct {
	Addr                string
	Topic               string
	Partition           int
	Messages            []string
	Client              sarama.Consumer
	WithSASL            bool
	KerberosConfigPath  string
	KerberosServiceName string
	KerberosUsername    string
	KerberosPassword    string
	KerberosRealm       string
}

// NewKafka initializes a default configs from web builder
func NewKafka(webBuilder *config.WebBuilder) *DefaultKafka {
	instance := new(DefaultKafka)
	instance.Addr = webBuilder.KafkaAddr
	instance.Topic = webBuilder.KafkaTopic
	instance.WithSASL = webBuilder.WithSASL
	instance.KerberosConfigPath = webBuilder.KerberosConfigPath
	instance.KerberosServiceName = webBuilder.KerberosServiceName
	instance.KerberosUsername = webBuilder.KerberosUsername
	instance.KerberosPassword = webBuilder.KerberosPassword
	instance.KerberosRealm = webBuilder.KerberosRealm

	return instance
}

// Connect to Kafka
func (dk *DefaultKafka) Connect() *DefaultKafka {
	config := sarama.NewConfig()
	config.ClientID = "interactws-consumer"
	config.Consumer.Return.Errors = true

	//Check if SASL is enabled
	if dk.WithSASL {
		config.Version = sarama.V2_0_0_0
		config.Net.SASL.Enable = dk.WithSASL
		config.Net.SASL.Handshake = true
		config.Net.SASL.Mechanism = sarama.SASLTypeGSSAPI
		config.Net.SASL.GSSAPI = sarama.GSSAPIConfig{
			AuthType:           sarama.KRB5_USER_AUTH,
			KerberosConfigPath: dk.KerberosConfigPath,
			ServiceName:        dk.KerberosServiceName,
			Username:           dk.KerberosUsername,
			Password:           dk.KerberosPassword,
			Realm:              dk.KerberosRealm,
		}
	}

	// For debug SASL_PLAINTEXT USING KERBEROS
	// saramaBroker := sarama.NewBroker(dk.Addr)
	// saramaBroker.Open(config)
	// fmt.Println(saramaBroker.Connected())

	client, err := sarama.NewConsumer([]string{dk.Addr}, config)
	if err != nil {
		logrus.Errorf("Error creating consumer client: %v", err)
		panic(fmt.Sprintf("Error creating consumer client: %v", err))
	}

	dk.Client = client

	return dk
}

// Listen all messages from Kafka topic list
func (dk *DefaultKafka) ListenGroup(handler func([]byte) error) {
	defer func() {
		if err := dk.Client.Close(); err != nil {
			logrus.Errorf("Error on closing connection: %v", err)
			panic(fmt.Sprintf("Error on closing connection: %v", err))
		}
	}()

	consumer, errors := dk.consume(handler)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// Count how many message processed
	msgCount := 0

	// Get signnal for finish
	doneCh := make(chan struct{})
	go func() {
		for {
			select {
			case <-consumer:
				msgCount++
			case consumerError := <-errors:
				msgCount++
				logrus.Debugf("Received consumerError\n\t Topic: %s, Partition: %s, Error: %v", string(consumerError.Topic), string(consumerError.Partition), consumerError.Err)
				doneCh <- struct{}{}
			case <-signals:
				logrus.Debugf("Interrupt is detected")
				doneCh <- struct{}{}
			}
		}
	}()

	<-doneCh
	logrus.Debugf("Processed %d messages", msgCount)
}

func (dk *DefaultKafka) consume(handler func(string, []byte) error) (chan *sarama.ConsumerMessage, chan *sarama.ConsumerError) {
	consumers := make(chan *sarama.ConsumerMessage)
	errors := make(chan *sarama.ConsumerError)
	topics, _ := dk.Client.Topics()

	for _, topic := range topics {
		if strings.Contains(topic, dk.Topic) {
			partitions, _ := dk.Client.Partitions(topic)

			for _, partition := range partitions {
				consumer, err := dk.Client.ConsumePartition(topic, partition, sarama.OffsetOldest)
				if nil != err {
					logrus.Errorf("Topic %v partitions: %v", topic, err)
					panic(fmt.Sprintf("Topic %v partitions: %v", topic, err))
				}

				go func(topic string, consumer sarama.PartitionConsumer) {
					for {
						select {
						case consumerError := <-consumer.Errors():
							errors <- consumerError

						case msg := <-consumer.Messages():
							consumers <- msg
							logrus.Debugf("Got message on topic (%s): %s", topic, msg.Value)
							handler(topic, msg.Value)
						}
					}
				}(topic, consumer)
			}
		}
	}

	return consumers, errors
}
