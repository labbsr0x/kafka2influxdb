package database

import (
	"fmt"
	"testing"

	"github.com/docker/docker/pkg/testutil/assert"
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
