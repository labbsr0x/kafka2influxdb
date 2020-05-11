package database

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/docker/docker/pkg/testutil/assert"
	"github.com/sirupsen/logrus"
)

// To run this test, you will need to run the docker/docker-compose.yml file and
// add to your hosts file the mapping broker to 127.0.0.1
func TestKafkaConnectSASLKerberos(t *testing.T) {
	instance := new(DefaultKafka)
	instance.Addr = "broker:9093"
	instance.Topic = "owner"
	instance.WithSASL = true
	instance.KerberosConfigPath = "../docker/kerberos-data/krb5.conf"
	instance.KerberosServiceName = "kafka"
	instance.KerberosUsername = "kafka/kafka2influx"
	instance.KerberosPassword = "kafka2influx1234"
	instance.KerberosRealm = "KERBEROS"

	client := instance.Connect()
	result, err := client.Client.Topics()
	if err != nil {
		panic(1)
	}

	fmt.Printf("List Topics: %s\n", result)
	assert.EqualStringSlice(t, result, []string{"__confluent.support.metrics"})
}

func TestConsumeInfoFromTopic(t *testing.T) {
	instance := new(DefaultKafka)
	instance.Addr = "broker.kafka-kerberos_default:9093"
	instance.Topic = "owner"
	instance.WithSASL = true
	instance.KerberosConfigPath = "../docker/kerberos-data/krb5.conf"
	instance.KerberosServiceName = "kafka"
	instance.KerberosUsername = "kafka/kafka2influx"
	instance.KerberosPassword = "influx1234"
	instance.KerberosRealm = "KERBEROS"

	client := instance.Connect()
	topics, _ := client.Client.Topics()

	consumers := make(chan *sarama.ConsumerMessage)
	errors := make(chan *sarama.ConsumerError)

	for _, topic := range topics {
		if strings.Contains(topic, client.Topic) {
			partitions, _ := client.Client.Partitions(topic)

			for _, partition := range partitions {
				consumer, err := client.Client.ConsumePartition(topic, partition, sarama.OffsetOldest)
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
							fmt.Println(msg.Value)
						}
					}
				}(topic, consumer)
			}
		}
	}
}
